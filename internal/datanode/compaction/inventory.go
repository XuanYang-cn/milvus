package compaction

import (
	"context"
	"sync"

	"github.com/milvus-io/milvus/internal/datanode/io"
	"github.com/milvus-io/milvus/internal/datanode/resource"
	"github.com/milvus-io/milvus/internal/proto/datapb"
	"github.com/milvus-io/milvus/internal/proto/etcdpb"

	"github.com/milvus-io/milvus/pkg/log"
	"github.com/milvus-io/milvus/pkg/util/tsoutil"
	"github.com/milvus-io/milvus/pkg/util/typeutil"

	"github.com/cockroachdb/errors"
	"github.com/milvus-io/milvus-proto/go-api/schemapb"
	"go.uber.org/zap"
)

type Inventory struct {
	memoryPool *resource.MemoryPool
}

func NewInventory(pool *resource.MemoryPool) *Inventory {
	in := Inventory{
		memoryPool: pool,
	}

	return &in
}

func (i *Inventory) ProvideWorker(meta *etcdpb.CollectionMeta, plan *datapb.CompactionPlan, pipeDispatch chan *LabeledRowData, fetchID UniqueID) (Worker, error) {

	limitedMemory, err := i.memoryPool.GetPreFetchSize(fetchID)
	if err != nil {
		// Should never happen
		return nil, err
	}
	worker := NewWorker(meta, plan, pipeDispatch, limitedMemory)
	return worker, nil
}
func (i *Inventory) ReleaseWorker() {}

type Worker interface {
	Work(wg *sync.WaitGroup, io *io.BinlogIO, errChan chan<- error)
	Close()
}

func NewWorker(meta *etcdpb.CollectionMeta, plan *datapb.CompactionPlan, pipeDispatch chan *LabeledRowData, limitedMemory uint64) Worker {
	switch plan.GetType() {
	// TODO: other types
	case datapb.CompactionType_MixCompaction: // TODO Minor Compaction
		return &MixedWorker{
			segmentCarts:  createSegmentCarts(meta, plan),
			pipeDispatch:  pipeDispatch,
			limitedMemory: limitedMemory}
	}
	return nil // TODO
}

func createSegmentCarts(meta *etcdpb.CollectionMeta, plan *datapb.CompactionPlan) map[UniqueID]*SegmentCart {
	var pkID int64
	var pkType schemapb.DataType

	for _, fs := range meta.GetSchema().GetFields() {
		if fs.GetIsPrimaryKey() && fs.GetFieldID() >= 100 && typeutil.IsPrimaryFieldType(fs.GetDataType()) {
			pkID = fs.GetFieldID()
			pkType = fs.GetDataType()
			break
		}
	}

	carts := make(map[UniqueID]*SegmentCart)
	for _, segBinlogs := range plan.GetSegmentBinlogs() {
		cart := SegmentCart{
			SegmentID: segBinlogs.GetSegmentID(),
			Segment:   segBinlogs,
			PkID:      pkID,
			PkType:    pkType,
		}

		switch plan.GetType() {
		case 0: // TODO:  SingleCompaction merge statslog
		case 100: // TODO:  SingleCompaction merge insertlogs
			cart.iterator = NewDynamicBinlogIterator()
		case 200: // TODO:  SingleCompaction merge deltalogs
			cart.iterator = NewDynamicDeltalogIterator()
		case 300: // TODO:  SingleCompaction merge statslogs
			// cart.iterator = NewDynamicStatslogIterator()
		case datapb.CompactionType_MixCompaction: // TODO MinorCompaction, merge small segments and delete data physically
			cart.iterator = NewFilterIterator(plan.GetTimetravel(), tsoutil.GetCurrentTime(), plan.GetCollectionTtl())
		case 400: // TODO: Major compaction
		}

		carts[segBinlogs.GetSegmentID()] = &cart
	}

	return carts
}

// StatsLogWorker handles statslogs of one segment
type StatsLogWorker struct {
	segmentCart *SegmentCart
}

// InsertLogWorker handles insertlogs of one segment
type InsertLogWorker struct {
	segmentCart *SegmentCart
}

// DeltaLogWorker handles deltalogs of one segment
type DeltaLogWorker struct {
	segmentCart *SegmentCart
}

// MixedWorker can handle both deltalogs and binlogs
// Often used by the merge of inserts and deletes
type MixedWorker struct {
	cartGuard     sync.Mutex // TODO
	segmentCarts  map[UniqueID]*SegmentCart
	pipeDispatch  chan<- *LabeledRowData
	limitedMemory uint64
}

func (w *MixedWorker) Close() {
	return
}

func (w *MixedWorker) Work(wg *sync.WaitGroup, io *io.BinlogIO, errChan chan<- error) {
	defer wg.Done()

	var iterWaiter sync.WaitGroup
	var notifyDone = make(chan struct{})
	var notifyOnce sync.Once
	iterWaiter.Add(len(w.segmentCarts))
	for _, cart := range w.segmentCarts {
		go func() {
			defer iterWaiter.Done()
			cart := cart
			for cart.iterator.HasNext() {
				select {
				case <-notifyDone:
					log.Info("Receive signal to return", zap.Any("cart", cart.SegmentID))
					return
				default:
				}

				labeled, err := cart.iterator.Next()
				if err != nil {
					log.Warn("iterator error", zap.Error(err))
					notifyOnce.Do(func() { close(notifyDone) })
					errChan <- err
					return
				}
				w.pipeDispatch <- labeled
			}
		}()
		if err := w.fillDeltaIterator(cart, io); err != nil {
			log.Warn("fillDeltaIterator error", zap.Error(err))
			notifyOnce.Do(func() { close(notifyDone) })
			errChan <- err
			return
		}
		if err := w.fillInsertIterator(cart, io); err != nil {
			log.Warn("fillInsertIterator error", zap.Error(err))
			notifyOnce.Do(func() { close(notifyDone) })
			errChan <- err
			return
		}
	}
	iterWaiter.Wait()
	close(w.pipeDispatch)
}

func (w *MixedWorker) fillDeltaIterator(cart *SegmentCart, binlogIO *io.BinlogIO) error {
	batches := CalculateBatchesByResources(w.limitedMemory, cart.Segment.GetDeltalogs())

	var err error
	var ctx = context.TODO()
	var lastIter Iterator
	waitForDisposed := func(iter Iterator) error {
		if iter == nil {
			return nil
		}
		select {
		case <-ctx.Done():
			return errors.Newf("fillDeltaIterator ctx done")
		// wait for iterator disposed
		case <-iter.WaitForDisposed():
			return nil
		}
	}
	for _, b := range batches {
		err = waitForDisposed(lastIter)
		if err != nil {
			return err
		}
		log.Info("compaction: YX: download delta batches", zap.Any("paths", len(b.paths)), zap.Any("size", b.size))
		raw, err := binlogIO.Download(ctx, b.paths)
		if err != nil {
			log.Warn("download paths error", zap.Strings("paths", b.paths), zap.Error(err))
			return err
		}

		lastIter, err = NewDeltalogIterator(raw, cart.PkID, cart.PkType, &Label{cart.SegmentID, cart.Segment.GetInsertChannel()})
		if err != nil {
			return err
		}
		cart.iterator.Add(lastIter)
	}

	cart.iterator.SetReadOnly()
	return nil
}

func (w *MixedWorker) fillInsertIterator(cart *SegmentCart, binlogIO *io.BinlogIO) error {
	batches := CalculateInsertBatchesByResources(w.limitedMemory, cart.Segment.GetFieldBinlogs())

	var ctx = context.TODO()
	var err error
	var lastIter Iterator
	waitForDisposed := func(iter Iterator) error {
		if iter == nil {
			return nil
		}
		select {
		case <-ctx.Done():
			return errors.Newf("fillInsertIterator ctx done")
		// wait for iterator disposed
		case <-iter.WaitForDisposed():
			return nil
		}
	}

	for idx, b := range batches {
		err = waitForDisposed(lastIter)
		if err != nil {
			return err
		}

		log.Info("compaction: YX: download insert batches", zap.Any("paths", len(b.paths)), zap.Any("size", b.size),
			zap.Any("batch idx", idx))
		raw, err := binlogIO.Download(ctx, b.paths)
		if err != nil {
			log.Warn("download paths error", zap.Strings("paths", b.paths), zap.Any("size", b.size), zap.Error(err))
			return err
		}

		lastIter, err = NewInsertBinlogIterator(raw, cart.PkID, cart.PkType, &Label{cart.SegmentID, cart.Segment.GetInsertChannel()})
		if err != nil {
			return err
		}
		cart.iterator.Add(lastIter)
	}

	cart.iterator.SetReadOnly()
	return nil
}

// TODO
func (w *MixedWorker) fillStatsIterator(iter InterDynamicIterator, binlogIO *io.BinlogIO, binlogs []*datapb.FieldBinlog) error {
	return nil
}

// only apply to deltalogs and statslogs
func CalculateBatchesByResources(limitedMemory uint64, fieldBinlogs []*datapb.FieldBinlog) []*batch {
	var (
		allSize    uint64
		allBinlogs []*datapb.Binlog
	)

	// flatting fieldBinlogs to allBinlogs, because delta/stats binlogs has one field
	for _, binlogs := range fieldBinlogs {
		for _, b := range binlogs.GetBinlogs() {
			allSize += uint64(b.GetLogSize())
		}
		allBinlogs = append(allBinlogs, binlogs.GetBinlogs()...)
	}

	var (
		batches  = make([]*batch, 0)
		tmpBatch = &batch{paths: make([]string, 0)}
	)

	for i := range allBinlogs {
		var (
			targetPaths = []string{allBinlogs[i].GetLogPath()}
			targetSize  = uint64(allBinlogs[i].GetLogSize())
		)

		if tmpBatch.mightReachLimit(limitedMemory, targetSize, len(targetPaths)) && tmpBatch.notEmpty() {
			batches = append(batches, tmpBatch)
			tmpBatch = &batch{paths: make([]string, 0)}
		}

		tmpBatch.add(targetPaths, targetSize)
	}
	if tmpBatch.notEmpty() {
		batches = append(batches, tmpBatch)
	}

	return batches
}

// only apply to inset binlogs
func CalculateInsertBatchesByResources(limitedMemory uint64, fieldBinlogs []*datapb.FieldBinlog) []*batch {
	var numOfSync int
	for _, b := range fieldBinlogs {
		if b != nil {
			numOfSync = len(b.GetBinlogs())
			break
		}
	}

	var (
		allBinlogs = make([][]*datapb.Binlog, numOfSync, numOfSync)
		allSize    uint64
	)
	for idx := 0; idx < numOfSync; idx++ {
		var batchBinlogs []*datapb.Binlog
		for _, f := range fieldBinlogs {
			binlog := f.GetBinlogs()[idx]
			batchBinlogs = append(batchBinlogs, binlog)
			allSize += uint64(binlog.GetLogSize())
		}
		allBinlogs[idx] = batchBinlogs
	}

	var (
		batches  = make([]*batch, 0)
		tmpBatch = &batch{paths: make([]string, 0)}
	)

	for i := range allBinlogs {
		var (
			syncPaths []string
			syncSize  uint64
		)
		for j := range allBinlogs[i] {
			syncPaths = append(syncPaths, allBinlogs[i][j].GetLogPath())
			syncSize += uint64(allBinlogs[i][j].GetLogSize())
		}

		if tmpBatch.mightReachLimit(limitedMemory, syncSize, len(syncPaths)) && tmpBatch.notEmpty() {
			batches = append(batches, tmpBatch)
			tmpBatch = &batch{paths: make([]string, 0)}
		}

		tmpBatch.add(syncPaths, syncSize)
	}
	if tmpBatch.notEmpty() {
		batches = append(batches, tmpBatch)
	}

	return batches
}

type batch struct {
	paths []string
	size  uint64
}

const CountLimit int = 1024 // TODO Configurable

func (b *batch) mightReachLimit(sizeLimit uint64, targetSize uint64, targetCount int) bool {
	return b.size+targetSize >= sizeLimit || len(b.paths)+targetCount >= CountLimit
}

func (b *batch) add(paths []string, size uint64) {
	b.paths = append(b.paths, paths...)
	b.size += size
}

func (b *batch) notEmpty() bool {
	return len(b.paths) > 0
}
