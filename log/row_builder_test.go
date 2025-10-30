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
	"testing"

	flatbuffers "github.com/google/flatbuffers/go"

	"github.com/lindb/common/proto/gen/v1/flatLogV1"
)

func Test_RowBuilder(t *testing.T) {
	rb := CreateRowBuilder()

	rb.AddMessage([]byte("message")).
		AddTimestamp(123).
		AddField([]byte("key1"), []byte("value1")).
		AddField([]byte("key2"), []byte("value2"))

	data, _ := rb.Build()
	log := flatLogV1.Log{}
	fmt.Println(string(data))
	log.Init(data, flatbuffers.GetUOffsetT(data))
	fmt.Println(string(log.Message()))
	fmt.Println(log.Timestamp())
	c := log.FieldsLength()
	fmt.Println(c)
	var f flatLogV1.Field
	for i := range c {
		log.Fields(&f, i)
		fmt.Println(string(f.Name()))
		fmt.Println(string(f.Value()))
	}
}
