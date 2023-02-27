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
	"context"
	"sync"
	"time"

	"github.com/milvus-io/milvus-proto/go-api/commonpb"
	"github.com/milvus-io/milvus/internal/datanode/allocator"
	"github.com/milvus-io/milvus/internal/datanode/io"
	"github.com/milvus-io/milvus/internal/datanode/resource"
	"github.com/milvus-io/milvus/internal/proto/datapb"
	"github.com/milvus-io/milvus/internal/proto/etcdpb"
	"github.com/milvus-io/milvus/internal/types"

	"github.com/milvus-io/milvus/pkg/log"
	"github.com/milvus-io/milvus/pkg/util/timerecord"
	"github.com/milvus-io/milvus/pkg/util/typeutil"

	"github.com/samber/lo"
	"go.uber.org/zap"
)

type (
	UniqueID  = typeutil.UniqueID
	Timestamp = typeutil.Timestamp
)

// CompactionFactory controls the schedule and executing of compaction work lines
// factory.CreateCompactionWorkLine()
// TODO between partitions
type CompactionFactory struct {
	lineGuard sync.Mutex
	workLines map[int64]*CompactionWorkLine // planID -> CompactionWorkLine
	backlogs  chan *CompactionWorkLine
	completed []UniqueID

	memoryPool *resource.MemoryPool
	binlogIO   *io.BinlogIO

	resources  *resource.GlobalResources
	inventory  *Inventory
	depository *Depository
	datacoord  types.DataCoord
	allocator  allocator.Allocator
}

func NewCompactionFactory(binlogIO *io.BinlogIO, resources *resource.GlobalResources, allocator allocator.Allocator) *CompactionFactory {
	// 50: 50, TODO
	pool := resource.NewMemoryPool(resources.GetMemoryInBytes() / 2) // Should be > MinimumMemoryForCompaction
	f := CompactionFactory{
		workLines:  make(map[UniqueID]*CompactionWorkLine), // worklines executing
		backlogs:   make(chan *CompactionWorkLine, 100),    // worklines in backlogs TODO: configurable
		completed:  make([]UniqueID, 0),
		binlogIO:   binlogIO,
		memoryPool: pool,
		inventory:  NewInventory(pool),
		depository: NewDepository(binlogIO, allocator),
		allocator:  allocator,
	}

	return &f
}

func (f *CompactionFactory) CreateWorkLine(plan *datapb.CompactionPlan, meta *etcdpb.CollectionMeta, targetParitionID UniqueID) error {
	newSegmentIDs, err := f.assignNewSegments(plan)
	if err != nil {
		return err
	}

	line, err := NewCompactionWorkLine(plan, meta, newSegmentIDs, targetParitionID, f.binlogIO)
	if err != nil {
		return err
	}

	f.backlogs <- line

	return nil
}

func (f *CompactionFactory) assignNewSegments(plan *datapb.CompactionPlan) ([]UniqueID, error) {
	// TODO GOOSE: enable to assign multiple segments or keep the original segmentID
	startID, _, err := f.allocator.Alloc(1)
	if err != nil {
		return nil, err
	}

	return []UniqueID{startID}, nil
}

func (f *CompactionFactory) Start(ctx context.Context) {
	// Schedule Worklines by resources
	log.Info("CompactionFactory started, YX")
	for {
		select {
		case <-ctx.Done():
			return
		case line := <-f.backlogs:
			fetchID := f.trySchedule(line)
			go f.executeWorkLine(line, fetchID)
		}
	}
}

const (
	MinimumMemoryForCompaction uint64 = 50 * 1024 * 1024 // 50 MB
)

func (f *CompactionFactory) trySchedule(line *CompactionWorkLine) UniqueID {
	log := log.With(
		zap.Int64("PlanID", line.Plan.GetPlanID()),
		zap.String("PlanType", line.CompactionType.String()),
		zap.String("Channel", line.Plan.GetChannel()),
	)
	log.Debug("YX: todo remove")
	var (
		estimated = f.EstimateMemoryUsage(line)
		ticker    = time.NewTicker(30 * time.Second) // TODO Configure
		available uint64
	)

	scheduled := func() bool {
		available = f.memoryPool.GetAvailableSize()

		if available < MinimumMemoryForCompaction {
			log.Warn("Unable to schedule compaction workline: available memory size less than MinimumMemoryForCompaction",
				zap.Uint64("available memory", available),
				zap.Uint64("minimum memory for compaction", MinimumMemoryForCompaction),
			)
			return false
		}

		executingCounts := f.getExecutingCompactionCounts()

		// For parallazation, available must be greater than estimated
		if executingCounts > 0 && available <= estimated {
			log.Warn("Unable to schedule parallel compaction workline: available memory size less than estimated size",
				zap.Int("executing compaction counts", executingCounts),
				zap.Uint64("available memory", available),
				zap.Uint64("estimated memory", estimated),
			)
			return false
		}

		log.Info("Schedule compaction workline: sufficient memory size or no executing compactions",
			zap.Int("executing compaction counts", executingCounts),
			zap.Uint64("estimated memory in MB", estimated/(1024*1024)),
			zap.Uint64("available memory in MB", available/(1024*1024)),
		)
		return true
	}

	for !scheduled() {
		<-ticker.C
	}

	fetchID, err := f.memoryPool.PreFetch(lo.Min([]uint64{available, estimated}))
	if err != nil {
		// TODO: should not happend at all
		return -1
	}
	return fetchID
}

func CalculateFieldBinlogSize(fbinlogs []*datapb.FieldBinlog) uint64 {
	var size uint64
	for _, fbinlog := range fbinlogs {
		for _, binlog := range fbinlog.GetBinlogs() {
			size += uint64(binlog.GetLogSize())
		}
	}
	return size
}

func (f *CompactionFactory) EstimateMemoryUsage(line *CompactionWorkLine) uint64 {
	// get segment binlog size
	var size uint64

	for _, segbinlogs := range line.Plan.SegmentBinlogs {
		// compactionType should be varified valid before
		switch line.CompactionType {
		case 0: // TODO: merge stats logs
			size += CalculateFieldBinlogSize(segbinlogs.GetField2StatslogPaths())
		case 1: // TODO: merge delta logs
			size += CalculateFieldBinlogSize(segbinlogs.GetDeltalogs())
		case 2: // TODO: merge insert logs
			size += CalculateFieldBinlogSize(segbinlogs.GetFieldBinlogs())
		case 3: // TODO: minor compaction
			size += CalculateFieldBinlogSize(segbinlogs.GetDeltalogs())
			size += CalculateFieldBinlogSize(segbinlogs.GetFieldBinlogs())
		case 4: // TODO: major compaction
		}
	}

	return size * 2
}

func (f *CompactionFactory) getExecutingCompactionCounts() int {
	f.lineGuard.Lock()
	defer f.lineGuard.Unlock()

	return len(f.workLines)
}

func (f *CompactionFactory) executeWorkLine(line *CompactionWorkLine, fetchID UniqueID) {
	tr := timerecord.NewTimeRecorder("YX")
	defer f.memoryPool.Return(fetchID)

	planID := line.Plan.GetPlanID()

	f.lineGuard.Lock()
	f.workLines[planID] = line
	f.lineGuard.Unlock()
	defer f.remove(line)

	line.Assemble(f.inventory, f.depository, fetchID)

	err := line.Run()
	if err != nil {
		log.Warn("compaction wrong", zap.Error(err))
	}

	f.completed = append(f.completed, planID)
	tr.Record("compaction finished")
}

func (f *CompactionFactory) remove(line *CompactionWorkLine) {
	f.lineGuard.Lock()
	delete(f.workLines, line.Plan.GetPlanID())
	f.lineGuard.Unlock()
}

func (f *CompactionFactory) GetResults(planID UniqueID) []*datapb.CompactionResult {
	results := f.depository.GetResults(planID)
	f.depository.RemovePacker(planID)
	return results
}

func (f *CompactionFactory) GetAllResults() []*datapb.CompactionStateResult {
	var results []*datapb.CompactionStateResult
	for _, planID := range f.completed {
		rsts := f.GetResults(planID)
		for _, rst := range rsts {
			results = append(results, &datapb.CompactionStateResult{
				State:  commonpb.CompactionState_Completed,
				PlanID: planID,
				Result: rst,
			})
		}
	}
	f.completed = []UniqueID{}
	return results
}
