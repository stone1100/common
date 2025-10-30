// Licensed to LinDB under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. LinDB licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package log

import (
	"fmt"
	"sync"

	flatbuffers "github.com/google/flatbuffers/go"

	"github.com/lindb/common/models"
	"github.com/lindb/common/pkg/fasttime"
	"github.com/lindb/common/proto/gen/v1/flatLogV1"
)

var rowBuilderPool sync.Pool

type RowBuilder struct {
	timestamp int64
	message   []byte

	rowKVs models.KVs

	// context for building flat log
	flatBuilder *flatbuffers.Builder
	keys        []flatbuffers.UOffsetT
	values      []flatbuffers.UOffsetT
	fields      []flatbuffers.UOffsetT
}

// NewRowBuilder picks a row builder from pool for building flat log.
func NewRowBuilder() (
	rb *RowBuilder,
	releaseFunc func(rb *RowBuilder),
) {
	releaseFunc = func(rb *RowBuilder) { rowBuilderPool.Put(rb) }
	item := rowBuilderPool.Get()
	if item != nil {
		builder := item.(*RowBuilder)
		builder.Reset()
	}
	return &RowBuilder{
		flatBuilder: flatbuffers.NewBuilder(1536),
	}, releaseFunc
}

// CreateRowBuilder creates a new row builder, not reused builder.
func CreateRowBuilder() *RowBuilder {
	return &RowBuilder{flatBuilder: flatbuffers.NewBuilder(1536)}
}

// AddTimestamp adds timestamp for log entry.
func (rb *RowBuilder) AddTimestamp(ts int64) *RowBuilder {
	rb.timestamp = ts
	return rb
}

// AddMessage adds message for log entry.
func (rb *RowBuilder) AddMessage(message []byte) *RowBuilder {
	rb.message = append(rb.message[:0], message...)
	return rb
}

// AddField appends a key-value pair
// Return false if field is invalid
func (rb *RowBuilder) AddField(key, value []byte) *RowBuilder {
	if len(key) == 0 || len(value) == 0 {
		return rb
	}
	rb.rowKVs.Count++

	if rb.rowKVs.Count > len(rb.rowKVs.Values) {
		rb.rowKVs.Values = append(rb.rowKVs.Values, models.KV{})
	}
	kvIdx := rb.rowKVs.Count - 1
	// copy key
	rb.rowKVs.Values[kvIdx].Key = append(rb.rowKVs.Values[kvIdx].Key[:0], key...)
	// copy value
	rb.rowKVs.Values[kvIdx].Value = append(rb.rowKVs.Values[kvIdx].Value[:0], value...)
	return rb
}

func (rb *RowBuilder) Build() ([]byte, error) {
	if len(rb.message) == 0 {
		return nil, fmt.Errorf("message is empty")
	}

	for i := 0; i < rb.rowKVs.Count; i++ {
		rb.keys = append(rb.keys, rb.flatBuilder.CreateByteString(rb.rowKVs.Values[i].Key))
		rb.values = append(rb.values, rb.flatBuilder.CreateByteString(rb.rowKVs.Values[i].Value))
	}

	// building key values vector
	for i := range len(rb.keys) {
		flatLogV1.FieldStart(rb.flatBuilder)
		flatLogV1.FieldAddName(rb.flatBuilder, rb.keys[i])
		flatLogV1.FieldAddValue(rb.flatBuilder, rb.values[i])
		rb.fields = append(rb.fields, flatLogV1.FieldEnd(rb.flatBuilder))
	}

	flatLogV1.LogStartFieldsVector(rb.flatBuilder, rb.rowKVs.Count)
	for i := rb.rowKVs.Count - 1; i >= 0; i-- {
		rb.flatBuilder.PrependUOffsetT(rb.fields[i])
	}
	fields := rb.flatBuilder.EndVector(rb.rowKVs.Count)

	message := rb.flatBuilder.CreateByteString(rb.message)

	flatLogV1.LogStart(rb.flatBuilder)
	if rb.timestamp == 0 {
		rb.timestamp = fasttime.UnixMilliseconds()
	}
	flatLogV1.LogAddMessage(rb.flatBuilder, message)
	flatLogV1.LogAddTimestamp(rb.flatBuilder, rb.timestamp)
	flatLogV1.LogAddFields(rb.flatBuilder, fields)
	end := flatLogV1.LogEnd(rb.flatBuilder)
	// size prefix encoding
	rb.flatBuilder.Finish(end)

	return rb.flatBuilder.FinishedBytes(), nil
}

func (rb *RowBuilder) Reset() {
	// reset flat builder context
	rb.flatBuilder.Reset()
	rb.timestamp = 0
	rb.message = rb.message[:0]

	// reset kvs context
	rb.rowKVs.Count = 0

	rb.keys = rb.keys[:0]
	rb.values = rb.values[:0]
	rb.fields = rb.fields[:0]
}
