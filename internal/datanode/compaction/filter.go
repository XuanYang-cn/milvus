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

package compaction

import (
	"time"

	"github.com/milvus-io/milvus/pkg/util/tsoutil"
)

type Filter func(row Row) bool

// affect deltaRow only
func TimeTravelFilter(travelTs Timestamp) Filter {
	return func(row Row) bool {
		deltaRow, ok := row.(*DeltalogRow)
		if !ok {
			return false
		}

		return deltaRow.Timestamp < travelTs
	}
}

// affect binlogRow only
func ExpireFilter(ttl int64, now Timestamp) Filter {
	return func(row Row) bool {
		if ttl <= 0 {
			return false
		}

		binRow, ok := row.(*InsertRow)
		if !ok {
			return false
		}

		ts := binRow.Timestamp
		pts, _ := tsoutil.ParseTS(ts)
		expireTime := pts.Add(time.Duration(ttl))

		pnow, _ := tsoutil.ParseTS(now)
		return expireTime.Before(pnow)
	}
}

// affect binlogRow only
func DeletionFilter(deleted map[interface{}]Timestamp) Filter {
	return func(row Row) bool {
		binRow, ok := row.(*InsertRow)
		if !ok {
			return false
		}

		ts, ok := deleted[binRow.PK.GetValue()]
		if ok && binRow.Timestamp < ts {
			return true
		}

		return false
	}
}
