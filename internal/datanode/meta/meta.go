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

package meta

import (
	"context"

	"github.com/milvus-io/milvus-proto/go-api/commonpb"
	"github.com/milvus-io/milvus-proto/go-api/msgpb"
	"github.com/milvus-io/milvus-proto/go-api/schemapb"
	"github.com/milvus-io/milvus/internal/proto/datapb"
	"github.com/milvus-io/milvus/internal/storage"
	"github.com/milvus-io/milvus/pkg/util/paramtable"
	"github.com/milvus-io/milvus/pkg/util/typeutil"
)

// Params from config.yaml
var Params *paramtable.ComponentParam = paramtable.Get()

type (
	primaryKey        = storage.PrimaryKey
	int64PrimaryKey   = storage.Int64PrimaryKey
	varCharPrimaryKey = storage.VarCharPrimaryKey
	UniqueID          = typeutil.UniqueID
	Timestamp         = typeutil.Timestamp
)

var (
	newInt64PrimaryKey   = storage.NewInt64PrimaryKey
	newVarCharPrimaryKey = storage.NewVarCharPrimaryKey
)

// Channel is DataNode unique replication
type Channel interface {
	getCollectionID() UniqueID
	getCollectionSchema(collectionID UniqueID, ts Timestamp) (*schemapb.CollectionSchema, error)
	getCollectionAndPartitionID(segID UniqueID) (collID, partitionID UniqueID, err error)
	getChannelName(segID UniqueID) string

	listAllSegmentIDs() []UniqueID
	listNotFlushedSegmentIDs() []UniqueID
	addSegment(req addSegmentReq) error
	listPartitionSegments(partID UniqueID) []UniqueID
	filterSegments(partitionID UniqueID) []*Segment
	listNewSegmentsStartPositions() []*datapb.SegmentStartPosition
	transferNewSegments(segmentIDs []UniqueID)
	updateSegmentPKRange(segID UniqueID, ids storage.FieldData)
	mergeFlushedSegments(ctx context.Context, seg *Segment, planID UniqueID, compactedFrom []UniqueID) error
	hasSegment(segID UniqueID, countFlushed bool) bool
	removeSegments(segID ...UniqueID)
	listCompactedSegmentIDs() map[UniqueID][]UniqueID
	listSegmentIDsToSync(ts Timestamp) []UniqueID
	setSegmentLastSyncTs(segID UniqueID, ts Timestamp)

	updateSegmentRowNumber(segID UniqueID, numRows int64)
	updateSegmentMemorySize(segID UniqueID, memorySize int64)
	InitPKstats(ctx context.Context, s *Segment, statsBinlogs []*datapb.FieldBinlog, ts Timestamp) error
	RollPKstats(segID UniqueID, stats []*storage.PrimaryKeyStats)
	getSegmentStatisticsUpdates(segID UniqueID) (*commonpb.SegmentStats, error)
	segmentFlushed(segID UniqueID)

	getChannelCheckpoint(ttPos *msgpb.MsgPosition) *msgpb.MsgPosition

	getCurInsertBuffer(segmentID UniqueID) (*BufferData, bool)
	setCurInsertBuffer(segmentID UniqueID, buf *BufferData)
	rollInsertBuffer(segmentID UniqueID)
	evictHistoryInsertBuffer(segmentID UniqueID, endPos *msgpb.MsgPosition)

	getCurDeleteBuffer(segmentID UniqueID) (*DelDataBuf, bool)
	setCurDeleteBuffer(segmentID UniqueID, buf *DelDataBuf)
	rollDeleteBuffer(segmentID UniqueID)
	evictHistoryDeleteBuffer(segmentID UniqueID, endPos *msgpb.MsgPosition)

	// getTotalMemorySize returns the sum of memory sizes of segments.
	getTotalMemorySize() int64
	forceToSync()
}

type addSegmentReq struct {
	segType                    datapb.SegmentType
	segID, collID, partitionID UniqueID
	numOfRows                  int64
	startPos, endPos           *msgpb.MsgPosition
	statsBinLogs               []*datapb.FieldBinlog
	recoverTs                  Timestamp
	importing                  bool
}
