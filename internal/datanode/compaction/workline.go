package compaction

import (
	"errors"
	"sync"

	"github.com/milvus-io/milvus/internal/datanode/io"
	"github.com/milvus-io/milvus/internal/proto/datapb"
	"github.com/milvus-io/milvus/internal/proto/etcdpb"

	"github.com/milvus-io/milvus/pkg/log"

	"github.com/milvus-io/milvus-proto/go-api/schemapb"
	"go.uber.org/zap"
)

type SegmentCart struct {
	SegmentID UniqueID
	Segment   *datapb.CompactionSegmentBinlogs // TODO

	PkID   UniqueID
	PkType schemapb.DataType

	iterator InterDynamicIterator // TODO
	done     chan struct{}

	cloaseOnce sync.Once
}

func (c *SegmentCart) Close() {
	c.cloaseOnce.Do(func() {
		c.done <- struct{}{}
	})
}

// CompactionWorkLine is in charge of one compaction.
// Submit segment carts to inventory for downloading and deserializing.
// Pip in streaming data when segment cart is ready.
// Dispatch streaming data to new segment.
// Pipe out streaming data to depository for serializing and uploading.
type CompactionWorkLine struct {
	Plan           *datapb.CompactionPlan
	CompactionType datapb.CompactionType

	meta *etcdpb.CollectionMeta

	InWorker  Worker
	OutPacker *Packer
	binlogIO  *io.BinlogIO

	PipeDispatch chan *LabeledRowData
	PipeOut      chan *LabeledRowData
	Dispatcher   Dispatcher // life time of dispatcher is managed by WorkLine

	PartitionID   UniqueID
	FinalSegments []UniqueID
}

func NewCompactionWorkLine(plan *datapb.CompactionPlan, meta *etcdpb.CollectionMeta, targetedSegmentIDs []UniqueID, targetPartitionID UniqueID, binlogIO *io.BinlogIO) (*CompactionWorkLine, error) {
	line := &CompactionWorkLine{
		Plan:           plan,
		CompactionType: plan.GetType(),
		FinalSegments:  targetedSegmentIDs,
		PartitionID:    targetPartitionID,
		meta:           meta,
		binlogIO:       binlogIO,

		PipeDispatch: make(chan *LabeledRowData, 1000),
		PipeOut:      make(chan *LabeledRowData, 1000),
	}

	dispatcher, err := line.getDispatcher(plan, targetedSegmentIDs)
	if err != nil {
		return nil, err
	}

	line.Dispatcher = dispatcher

	return line, nil
}

func (l *CompactionWorkLine) Assemble(in *Inventory, out *Depository, fetchID UniqueID) error {
	worker, err := in.ProvideWorker(l.meta, l.Plan, l.PipeDispatch, fetchID)
	if err != nil {
		return err
	}
	packer := out.ProvidePacker(l)

	l.InWorker = worker
	l.OutPacker = packer
	return nil
}

func (l *CompactionWorkLine) Run() error {
	var (
		err     error
		errChan = make(chan error, 10)
		done    = make(chan struct{})

		lineWaiter sync.WaitGroup
	)

	lineWaiter.Add(3)
	go l.InWorker.Work(&lineWaiter, l.binlogIO, errChan)
	go l.Dispatcher.Dispatch(&lineWaiter, errChan)
	go l.OutPacker.Pack(&lineWaiter, errChan)
	go func() {
		select {
		case err = <-errChan:
			log.Warn("executing comapction wrong", zap.Error(err))
			l.InWorker.Close()
			l.Dispatcher.Done()
			l.OutPacker.ForceDone()
		case <-done:
		}
	}()

	lineWaiter.Wait()
	close(done)
	return err
}

func (l *CompactionWorkLine) getDispatcher(plan *datapb.CompactionPlan, segmentIDs []UniqueID) (Dispatcher, error) {
	if len(segmentIDs) < 1 {
		return nil, errors.New("invalid compaction type")
	}

	return NewDispatcher(l.PipeDispatch, l.PipeOut, segmentIDs...), nil
}
