// Licensed to the LF AI & Data foundation under one
// or more contributor license agreements. See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership. The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License. You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package resource

import (
	"github.com/milvus-io/milvus/pkg/util/hardware"
)

type GlobalResources struct {
}

func New() *GlobalResources {
	return &GlobalResources{}
}

func (gr *GlobalResources) GetMemoryInBytes() uint64 {
	return hardware.GetMemoryCount()
}

func (gr *GlobalResources) GetFreeMemoryInMB() uint64 {
	return hardware.GetFreeMemoryCount() / (1 << 20)
}

func (gr *GlobalResources) GetFreeMemoryInBytes() uint64 {
	return hardware.GetFreeMemoryCount()
}

func (gr *GlobalResources) GetCPUUsage() float64 {
	return hardware.GetCPUUsage()
}

// Get zero now
func (gr *GlobalResources) GetIOPS() int64 {
	// TODO
	return 0
}

// Do nothing now
func (gr *GlobalResources) SubmitIO(count int) {
	// now := time.Now()
	// TODO
	return
}
