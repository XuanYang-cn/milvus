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
	"container/heap"
	"fmt"
	"math"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/milvus-io/milvus/internal/datanode/util"

	"github.com/milvus-io/milvus-proto/go-api/commonpb"
	"github.com/milvus-io/milvus-proto/go-api/msgpb"
	"github.com/milvus-io/milvus-proto/go-api/schemapb"
	"github.com/milvus-io/milvus/internal/proto/datapb"
	"github.com/milvus-io/milvus/pkg/log"
	"github.com/milvus-io/milvus/pkg/util/etcd"
	"github.com/milvus-io/milvus/pkg/util/paramtable"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain(t *testing.M) {
	rand.Seed(time.Now().Unix())
	// init embed etcd
	embedetcdServer, tempDir, err := etcd.StartTestEmbedEtcdServer()
	if err != nil {
		log.Fatal(err.Error())
	}
	defer os.RemoveAll(tempDir)
	defer embedetcdServer.Close()

	addrs := etcd.GetEmbedEtcdEndpoints(embedetcdServer)
	// setup env for etcd endpoint
	os.Setenv("etcd.endpoints", strings.Join(addrs, ","))

	path := "/tmp/milvus_ut/rdb_data"
	os.Setenv("ROCKSMQ_PATH", path)
	defer os.RemoveAll(path)

	Params.Init()
	// change to specific channel for test
	paramtable.Get().Save(Params.EtcdCfg.Endpoints.Key, strings.Join(addrs, ","))

	code := t.Run()
	os.Exit(code)
}

func genTestCollectionSchema(dim int64, vectorType schemapb.DataType) *schemapb.CollectionSchema {
	floatVecFieldSchema := &schemapb.FieldSchema{
		FieldID:  100,
		Name:     "vec",
		DataType: vectorType,
		TypeParams: []*commonpb.KeyValuePair{
			{
				Key:   "dim",
				Value: fmt.Sprintf("%d", dim),
			},
		},
	}
	schema := &schemapb.CollectionSchema{
		Name: "collection-0",
		Fields: []*schemapb.FieldSchema{
			floatVecFieldSchema,
		},
	}
	return schema
}

func TestBufferData(t *testing.T) {
	paramtable.Get().Save(Params.DataNodeCfg.FlushInsertBufferSize.Key, strconv.FormatInt(16*(1<<20), 10)) // 16 MB
	tests := []struct {
		isValid bool

		indim         int64
		expectedLimit int64
		vectorType    schemapb.DataType

		description string
	}{
		{true, 1, 4194304, schemapb.DataType_FloatVector, "Smallest of the DIM"},
		{true, 128, 32768, schemapb.DataType_FloatVector, "Normal DIM"},
		{true, 32768, 128, schemapb.DataType_FloatVector, "Largest DIM"},
		{true, 4096, 32768, schemapb.DataType_BinaryVector, "Normal binary"},
		{false, 0, 0, schemapb.DataType_FloatVector, "Illegal DIM"},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			idata, err := newBufferData(genTestCollectionSchema(test.indim, test.vectorType))

			if test.isValid {
				assert.NoError(t, err)
				assert.NotNil(t, idata)

				assert.Equal(t, test.expectedLimit, idata.limit)
				assert.Zero(t, idata.size)

				capacity := idata.effectiveCap()
				assert.Equal(t, test.expectedLimit, capacity)
			} else {
				assert.Error(t, err)
				assert.Nil(t, idata)
			}
		})
	}
}

func TestBufferData_updateTimeRange(t *testing.T) {
	paramtable.Get().Save(Params.DataNodeCfg.FlushInsertBufferSize.Key, strconv.FormatInt(16*(1<<20), 10)) // 16 MB

	type testCase struct {
		tag string

		trs        []util.TimeRange
		expectFrom Timestamp
		expectTo   Timestamp
	}

	cases := []testCase{
		{
			tag:        "no input range",
			expectTo:   0,
			expectFrom: math.MaxUint64,
		},
		{
			tag: "single range",
			trs: []util.TimeRange{
				util.NewTimeRange(100, 200),
			},
			expectFrom: 100,
			expectTo:   200,
		},
		{
			tag: "multiple range",
			trs: []util.TimeRange{
				util.NewTimeRange(150, 250),
				util.NewTimeRange(100, 200),
				util.NewTimeRange(50, 180),
			},
			expectFrom: 50,
			expectTo:   250,
		},
	}

	for _, tc := range cases {
		t.Run(tc.tag, func(t *testing.T) {
			bd, err := newBufferData(genTestCollectionSchema(16, schemapb.DataType_FloatVector))
			require.NoError(t, err)
			for _, tr := range tc.trs {
				bd.updateTimeRange(tr)
			}

			assert.Equal(t, tc.expectFrom, bd.tsFrom)
			assert.Equal(t, tc.expectTo, bd.tsTo)
		})
	}
}

func TestPriorityQueueString(t *testing.T) {
	item := &Item{
		segmentID:  0,
		memorySize: 1,
	}

	assert.Equal(t, "<segmentID=0, memorySize=1>", item.String())

	pq := &PriorityQueue{}
	heap.Push(pq, item)
	assert.Equal(t, "[<segmentID=0, memorySize=1>]", pq.String())
}

func Test_CompactSegBuff(t *testing.T) {
	channelSegments := make(map[UniqueID]*Segment)
	delBufferManager := &DeltaBufferManager{
		channel: &ChannelMeta{
			segments: channelSegments,
		},
		delBufHeap: &PriorityQueue{},
	}
	//1. set compactTo and compactFrom
	targetSeg := &Segment{segmentID: 3333}
	targetSeg.setType(datapb.SegmentType_Flushed)

	seg1 := &Segment{
		segmentID:   1111,
		compactedTo: targetSeg.segmentID,
	}
	seg1.setType(datapb.SegmentType_Compacted)

	seg2 := &Segment{
		segmentID:   2222,
		compactedTo: targetSeg.segmentID,
	}
	seg2.setType(datapb.SegmentType_Compacted)

	channelSegments[seg1.segmentID] = seg1
	channelSegments[seg2.segmentID] = seg2
	channelSegments[targetSeg.segmentID] = targetSeg

	//2. set up deleteDataBuf for seg1 and seg2
	delDataBuf1 := newDelDataBuf(seg1.segmentID)
	delDataBuf1.EntriesNum++
	delDataBuf1.updateStartAndEndPosition(nil, &msgpb.MsgPosition{Timestamp: 50})
	delBufferManager.updateMeta(seg1.segmentID, delDataBuf1)
	heap.Push(delBufferManager.delBufHeap, delDataBuf1.item)

	delDataBuf2 := newDelDataBuf(seg2.segmentID)
	delDataBuf2.EntriesNum++
	delDataBuf2.updateStartAndEndPosition(nil, &msgpb.MsgPosition{Timestamp: 50})
	delBufferManager.updateMeta(seg2.segmentID, delDataBuf2)
	heap.Push(delBufferManager.delBufHeap, delDataBuf2.item)

	//3. test compact
	delBufferManager.UpdateCompactedSegments()

	//4. expect results in two aspects:
	//4.1 compactedFrom segments are removed from delBufferManager
	//4.2 compactedTo seg is set properly with correct entriesNum
	_, seg1Exist := delBufferManager.Load(seg1.segmentID)
	_, seg2Exist := delBufferManager.Load(seg2.segmentID)
	assert.False(t, seg1Exist)
	assert.False(t, seg2Exist)
	assert.Equal(t, int64(2), delBufferManager.GetEntriesNum(targetSeg.segmentID))

	// test item of compactedToSegID is correct
	targetSegBuf, ok := delBufferManager.Load(targetSeg.segmentID)
	assert.True(t, ok)
	assert.NotNil(t, targetSegBuf.item)
	assert.Equal(t, targetSeg.segmentID, targetSegBuf.item.segmentID)

	//5. test roll and evict (https://github.com/milvus-io/milvus/issues/20501)
	delBufferManager.channel.rollDeleteBuffer(targetSeg.segmentID)
	_, segCompactedToExist := delBufferManager.Load(targetSeg.segmentID)
	assert.False(t, segCompactedToExist)
	delBufferManager.channel.evictHistoryDeleteBuffer(targetSeg.segmentID, &msgpb.MsgPosition{
		Timestamp: 100,
	})
	cp := delBufferManager.channel.getChannelCheckpoint(&msgpb.MsgPosition{
		Timestamp: 200,
	})
	assert.Equal(t, Timestamp(200), cp.Timestamp) // evict all buffer, use ttPos as cp
}
