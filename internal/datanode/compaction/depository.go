package compaction

import (
	"context"
	"strings"
	"sync"

	"github.com/milvus-io/milvus/internal/datanode/allocator"
	"github.com/milvus-io/milvus/internal/datanode/io"
	"github.com/milvus-io/milvus/internal/proto/datapb"
	"github.com/milvus-io/milvus/internal/proto/etcdpb"

	"github.com/milvus-io/milvus/pkg/common"
	"github.com/milvus-io/milvus/pkg/log"

	"github.com/cockroachdb/errors"
	"go.uber.org/zap"
)

type Buffer interface {
	Write(row Row) error
	Full() bool
}

func NewResult(planID, segmentID UniqueID) *datapb.CompactionResult {
	return &datapb.CompactionResult{
		PlanID:              planID,
		SegmentID:           segmentID,
		InsertLogs:          make([]*datapb.FieldBinlog, 0),
		Deltalogs:           make([]*datapb.FieldBinlog, 0),
		Field2StatslogPaths: make([]*datapb.FieldBinlog, 0),
	}
}

type SegmentManager struct {
	segmentID UniqueID

	mutable MutableBuffer // delta data and insert data are guarenteed to arrived in order

	meta   *etcdpb.CollectionMeta
	result *datapb.CompactionResult
}

func NewSegmentManager(planID, segmentID UniqueID, planType datapb.CompactionType, meta *etcdpb.CollectionMeta, chunkManagerRoot string) *SegmentManager {
	var mutable MutableBuffer
	switch planType {
	case 0: // TODO legacy
	case 100: // single delta
		mutable = NewDeltaBuffer(chunkManagerRoot)
	case 200: // single insert
		mutable = NewInsertBuffer(MaxBufferSize, meta.GetSchema(), chunkManagerRoot)
	case 300: // single stats
		// mutable = &StatsBuffer{} // TODO
	case datapb.CompactionType_MixCompaction: // minor merge delta/insert
		mutable = NewDeltaBuffer(chunkManagerRoot)

	}
	m := SegmentManager{
		segmentID: segmentID,
		mutable:   mutable,
		meta:      meta,
		result:    NewResult(planID, segmentID),
	}

	return &m
}

func (s *SegmentManager) Refresh(meta *etcdpb.CollectionMeta, chunkManagerRoot string) error {
	switch s.mutable.(type) {
	// case *StatsBuffer: TODO
	case *DeltaBuffer: // TODO
		s.SetMutable(NewDeltaBuffer(chunkManagerRoot))
	case *InsertBuffer:
		s.SetMutable(NewInsertBuffer(MaxBufferSize, meta.GetSchema(), chunkManagerRoot))
	}
	return nil
}

// RowMatchBuffer return false on and only on BinlogRow with DeltaBuffer
// because we initialize mixed buffer to DeltaBuffer
func (s *SegmentManager) RowMatchBuffer(row Row) bool {
	_, isDeltaRow := row.(*DeltalogRow)
	_, isDeltaBuf := s.mutable.(*DeltaBuffer)

	_, isBinRow := row.(*InsertRow)
	_, isInsertBuf := s.mutable.(*InsertBuffer)

	return (isDeltaRow && isDeltaBuf) || (isBinRow && isInsertBuf)
}

func (s *SegmentManager) SetMutable(mu MutableBuffer) {
	s.mutable = nil
	s.mutable = mu
}

func (s *SegmentManager) appendInsertPaths(paths map[UniqueID]*datapb.FieldBinlog) {
	for idx := range s.result.GetField2StatslogPaths() {
		fieldBinlogs := s.result.Field2StatslogPaths[idx]
		fieldID := fieldBinlogs.GetFieldID()

		targetBinlogs, ok := paths[fieldID]
		if !ok {
			continue // TODO, should not happen
		}
		fieldBinlogs.Binlogs = append(fieldBinlogs.Binlogs, targetBinlogs.GetBinlogs()...)
	}
}

func (s *SegmentManager) appendDeltaPaths(paths map[UniqueID]*datapb.FieldBinlog) {
	targetBinlogs, ok := paths[0] // TODO
	if !ok {
		return // TODO, should not happen
	}
	s.result.Deltalogs[0].Binlogs = append(s.result.Deltalogs[0].Binlogs, targetBinlogs.GetBinlogs()...)
}

// TODO
func (s *SegmentManager) appendStatsPaths(paths map[UniqueID]*datapb.FieldBinlog) {
	return
}

func (s *SegmentManager) AppendPaths(paths map[UniqueID]*datapb.FieldBinlog) {
	// TODO varify correctness of paths
	var one string
	for _, fbinlog := range paths {
		one = fbinlog.GetBinlogs()[0].GetLogPath()
	}

	if strings.Contains(one, common.SegmentInsertLogPath) {
		s.appendInsertPaths(paths)
		return
	}

	if strings.Contains(one, common.SegmentDeltaLogPath) {
		s.appendDeltaPaths(paths)
		return
	}

	if strings.Contains(one, common.SegmentStatslogPath) {
		s.appendStatsPaths(paths)
		return
	}
}

type Packer struct {
	planID      UniqueID
	planType    datapb.CompactionType
	meta        *etcdpb.CollectionMeta
	partitionID UniqueID

	binlogIO           *io.BinlogIO
	immutables         chan map[string][]byte
	closeImmutableOnce sync.Once
	pipeIn             <-chan *LabeledRowData

	segmentGuard sync.Mutex
	segments     map[UniqueID]*SegmentManager // SegmentID -> SegmentManager
	allocator    allocator.Allocator

	forceDoneOnce sync.Once
	forceDone     chan struct{}
}

func (p *Packer) ForceDone() {
	p.forceDoneOnce.Do(func() {
		close(p.forceDone)
	})
}

func (p *Packer) Pack(lineWaiter *sync.WaitGroup, errChan chan<- error) {
	log.Info("YX: Start packing", zap.Any("planID", p.planID))
	defer lineWaiter.Done()

	var packerWaiter sync.WaitGroup
	packerWaiter.Add(2)
	go p.Upload(&packerWaiter, errChan)
	go p.Buffer(&packerWaiter, errChan)
	packerWaiter.Wait()
	log.Info("YX: Finish packing", zap.Any("planID", p.planID))
}

func (p *Packer) AddImmutable(kvs map[string][]byte) {
	p.immutables <- kvs
}

func (p *Packer) Upload(packerWaiter *sync.WaitGroup, errChan chan<- error) {
	log := log.With(zap.Int64("planID", p.planID))
	log.Info("YX: Packer start uploading")
	defer packerWaiter.Done()
	var ctx = context.TODO()
	for {
		select {
		case immutable, ok := <-p.immutables:
			if !ok {
				log.Info("YX: Packer finish uploading")
				return
			}
			log.Info("YX: Packer uploading immutable")
			if err := p.binlogIO.Upload(ctx, immutable); err != nil {
				log.Warn("Packer fail to upload", zap.Error(err))
				errChan <- err
				return
			}
		case <-p.forceDone: // force stopped by other goroutine
			log.Warn("Packer upload force stopped by other goroutine")
			return
		}
	}

}

func (p *Packer) Buffer(packerWaiter *sync.WaitGroup, errChan chan<- error) {
	log.Info("YX: Start buffering", zap.Any("planID", p.planID))
	defer packerWaiter.Done()
	var err error

	for {
		select {
		case row, ok := <-p.pipeIn:
			if !ok {
				log.Info("YX: Finish buffering, freeze all left in buffer", zap.Any("planID", p.planID))
				if err := p.freezeAll(); err != nil {
					log.Warn("Packer freeze all error", zap.Error(err))
					errChan <- err
				}
				// TODO clear the buffer
				return
			}
			if err = p.buffer(row); err != nil {
				log.Warn("Packer buffer segment wrong", zap.Error(err))
				errChan <- err
				return
			}
		case <-p.forceDone: // force stopped by other goroutine
			log.Warn("Packer buffer force stopped by other goroutine")
			return
		}
	}
}

func (p *Packer) freezeAll() error {
	p.segmentGuard.Lock()
	defer p.segmentGuard.Unlock()
	for ID, seg := range p.segments {
		if seg.mutable.Empty() {
			continue
		}
		immutable, paths, err := seg.mutable.Freeze(ID, p.partitionID, p.meta, p.allocator)
		if err != nil {
			return err // TODO wrapper error
		}
		p.AddImmutable(immutable)
		seg.AppendPaths(paths)
	}

	p.closeImmutableOnce.Do(func() {
		close(p.immutables)
	})
	return nil
}

func (p *Packer) buffer(labeled *LabeledRowData) error {
	p.segmentGuard.Lock()
	seg, ok := p.segments[labeled.GetSegmentID()]
	p.segmentGuard.Unlock()
	if !ok {
		return errors.Newf("No segment in the packer, segmentID=%d", labeled.GetSegmentID())
	}

	// Buffer mix when merging delete with inserts
	// TODO Minor compaction and major compaction
	if p.planType == datapb.CompactionType_MixCompaction {
		return p.bufferSegmentMix(seg, labeled.data)
	}

	return p.bufferSegment(seg, labeled.data)
}

const MaxBufferSize int64 = 16 * 1024 * 1024

func (p *Packer) bufferSegmentMix(s *SegmentManager, row Row) error {
	if !s.RowMatchBuffer(row) {
		// When InsertRow not matching with DeltaBuffer
		// 1. freeze all delete data and
		// 2. refresh mutable to InsertBuffer
		if !s.mutable.Empty() {
			immutable, paths, err := s.mutable.Freeze(s.segmentID, p.partitionID, p.meta, p.allocator)
			if err != nil {
				return err
			}

			p.AddImmutable(immutable)
			s.AppendPaths(paths)
		}
		s.SetMutable(NewInsertBuffer(MaxBufferSize, p.meta.GetSchema(), p.binlogIO.RootPath()))
	}

	return p.bufferSegment(s, row)
}

func (p *Packer) bufferSegment(s *SegmentManager, row Row) error {
	if s.mutable.Full() {
		immutable, paths, err := s.mutable.Freeze(s.segmentID, p.partitionID, p.meta, p.allocator)
		if err != nil {
			return err
		}
		s.AppendPaths(paths)
		p.AddImmutable(immutable)
		s.Refresh(p.meta, p.binlogIO.RootPath())
		log.Info("YX: mutable full, change to immutable",
			zap.Any("planID", p.planID),
			zap.Any("segmentID", s.segmentID),
		)
	}

	return s.mutable.Write(row)
}

type Depository struct {
	packerGuard sync.Mutex
	Packers     map[UniqueID]*Packer // PlanID to packer

	allocator allocator.Allocator
	binlogIO  *io.BinlogIO
}

func NewDepository(binlogIO *io.BinlogIO, allocator allocator.Allocator) *Depository {
	d := Depository{
		binlogIO:  binlogIO,
		allocator: allocator,
		Packers:   make(map[UniqueID]*Packer),
	}
	return &d
}

func (d *Depository) ProvidePacker(line *CompactionWorkLine) *Packer {
	var (
		planID   = line.Plan.GetPlanID()
		planType = line.Plan.GetType()
	)

	p := &Packer{
		planID:      planID,
		planType:    planType,
		allocator:   d.allocator,
		meta:        line.meta,
		binlogIO:    line.binlogIO,
		partitionID: line.PartitionID,

		pipeIn:     line.PipeOut,
		immutables: make(chan map[string][]byte, 10), // TODO
		segments:   make(map[UniqueID]*SegmentManager),

		forceDone: make(chan struct{}),
	}

	for _, sID := range line.FinalSegments {
		manager := NewSegmentManager(planID, sID, planType, p.meta, d.binlogIO.RootPath())

		p.segmentGuard.Lock()
		p.segments[sID] = manager
		p.segmentGuard.Unlock()
	}

	d.packerGuard.Lock()
	d.Packers[planID] = p
	d.packerGuard.Unlock()

	return p
}

// GetResults returns the results of a compaction plan and remove the packer from depository
// TODO
func (d *Depository) GetResults(planID UniqueID) []*datapb.CompactionResult {
	d.packerGuard.Lock()
	p, ok := d.Packers[planID]
	d.packerGuard.Unlock()
	if !ok {
		log.Warn("No results for the plan", zap.Int64("planID", planID))
		return nil
	}

	var results []*datapb.CompactionResult
	p.segmentGuard.Lock()
	for _, s := range p.segments {
		results = append(results, s.result)
	}
	p.segmentGuard.Unlock()

	d.packerGuard.Lock()
	delete(d.Packers, planID)
	d.packerGuard.Unlock()
	log.Info("Remove the packer for plan", zap.Int64("planID", planID))

	return results
}

func (d *Depository) RemovePacker(planID UniqueID) {
	d.packerGuard.Lock()
	p, ok := d.Packers[planID]
	d.packerGuard.Unlock()
	if !ok {
		return
	}

	p.ForceDone()

	d.packerGuard.Lock()
	delete(d.Packers, planID)
	d.packerGuard.Unlock()
}

func (d *Depository) Close() {
	d.packerGuard.Lock()
	for planID, p := range d.Packers {
		p.ForceDone()
		delete(d.Packers, planID)
	}
	d.packerGuard.Unlock()
}
