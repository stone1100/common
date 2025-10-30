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

package models

import (
	"bytes"
)

// KV represents a key-value.
type KV struct {
	Key   []byte
	Value []byte
}

// KVs sorts key values, then computes the hash
type KVs struct {
	Values []KV
	Count  int
}

func (items KVs) Len() int { return items.Count }

func (items KVs) Swap(i, j int) { items.Values[i], items.Values[j] = items.Values[j], items.Values[i] }

func (items KVs) Less(i, j int) bool {
	return bytes.Compare(items.Values[i].Key, items.Values[j].Key) < 0
}
