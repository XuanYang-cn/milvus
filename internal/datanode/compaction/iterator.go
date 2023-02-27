package compaction

import (
	"sync"
	"time"

	"github.com/milvus-io/milvus-proto/go-api/schemapb"
	"github.com/milvus-io/milvus/internal/storage"
	"github.com/milvus-io/milvus/pkg/common"
	"github.com/milvus-io/milvus/pkg/log"

	"github.com/cockroachdb/errors"
	"go.uber.org/atomic"
)

var (
	// ErrNoMoreRecord is the error that the iterator does not have next record.
	ErrNoMoreRecord = errors.New("no more record")
	// ErrDisposed is the error that the iterator is disposed.
	ErrDisposed = errors.New("iterator is disposed")
)

type Row interface {
}

type InsertRow struct {
	ID        int64
	PK        storage.PrimaryKey
	Timestamp Timestamp
	Value     map[storage.FieldID]interface{}
}

type DeltalogRow struct {
	Pk        storage.PrimaryKey
	Timestamp Timestamp
}

// TODO
type StatsRow struct {
	stats storage.PrimaryKeyStats
}

type Label struct {
	segmentID UniqueID
	shard     string
}

type LabeledRowData struct {
	label *Label
	data  Row
}

const InvalidID int64 = -1

func (l *LabeledRowData) GetSegmentID() UniqueID {
	if l.label == nil {
		return InvalidID
	}

	return l.label.segmentID
}

func NewLabeledRowData(data Row, label *Label) *LabeledRowData {
	return &LabeledRowData{
		label: label,
		data:  data,
	}
}

type Iterator interface {
	HasNext() bool
	Next() (*LabeledRowData, error)
	Dispose()
	WaitForDisposed() <-chan struct{} // wait until the iterator is disposed
}

type InterDynamicIterator interface {
	Iterator
	Add(iter Iterator)
	ReadOnly() bool
	SetReadOnly()
}

type BinlogIterator struct {
	disposed     atomic.Bool // 0: false, 1: true
	disposedCh   chan struct{}
	disposedOnce sync.Once

	data      *storage.InsertData
	label     *Label
	pkFieldID int64
	pkType    schemapb.DataType
	pos       int
}

var _ Iterator = (*BinlogIterator)(nil)
var _ Iterator = (*DeltalogIterator)(nil)

// NewInsertBinlogIterator creates a new iterator
func NewInsertBinlogIterator(v [][]byte, pkFieldID UniqueID, pkType schemapb.DataType, label *Label) (*BinlogIterator, error) {
	blobs := make([]*storage.Blob, len(v))
	for i := range blobs {
		blobs[i] = &storage.Blob{Value: v[i]}
	}

	reader := storage.NewInsertCodec()
	_, _, iData, err := reader.Deserialize(blobs)

	if err != nil {
		return nil, err
	}

	return &BinlogIterator{
		disposedCh: make(chan struct{}),
		data:       iData,
		pkFieldID:  pkFieldID,
		pkType:     pkType,
		label:      label}, nil
}

// HasNext returns true if the iterator have unread record
func (i *BinlogIterator) HasNext() bool {
	return !i.isDisposed() && i.hasNext()
}

func (i *BinlogIterator) Next() (*LabeledRowData, error) {
	if i.isDisposed() {
		return nil, ErrDisposed
	}

	if !i.hasNext() {
		return nil, ErrNoMoreRecord
	}

	m := make(map[int64]interface{})
	for fieldID, fieldData := range i.data.Data {
		m[fieldID] = fieldData.GetRow(i.pos)
	}

	pk, err := storage.GenPrimaryKeyByRawData(i.data.Data[i.pkFieldID].GetRow(i.pos), i.pkType)
	if err != nil {
		return nil, err
	}

	row := &InsertRow{
		ID:        i.data.Data[common.RowIDField].GetRow(i.pos).(int64),
		Timestamp: uint64(i.data.Data[common.TimeStampField].GetRow(i.pos).(int64)),
		PK:        pk,
		Value:     m,
	}
	i.pos++
	return NewLabeledRowData(row, i.label), nil
}

// Dispose disposes the iterator
func (i *BinlogIterator) Dispose() {
	i.disposed.CompareAndSwap(false, true)
	i.disposedOnce.Do(func() {
		close(i.disposedCh)
	})
}

func (i *BinlogIterator) hasNext() bool {
	_, ok := i.data.Data[common.RowIDField]
	if !ok {
		return false
	}
	return i.pos < i.data.Data[common.RowIDField].RowNum()
}

func (i *BinlogIterator) isDisposed() bool {
	return i.disposed.Load()
}

// Disposed wait forever for the iterator to dispose
func (i *BinlogIterator) WaitForDisposed() <-chan struct{} {
	<-i.disposedCh
	notifyCh := make(chan struct{})
	close(notifyCh)
	return notifyCh
}

type DeltalogIterator struct {
	disposeCh    chan struct{}
	disposedOnce sync.Once
	disposed     atomic.Bool

	data      *storage.DeleteData
	label     *Label
	pkFieldID int64
	pkType    schemapb.DataType
	pos       int
}

func NewDeltalogIterator(v [][]byte, pkFieldID UniqueID, pkType schemapb.DataType, label *Label) (*DeltalogIterator, error) {
	blobs := make([]*storage.Blob, len(v))
	for i := range blobs {
		blobs[i] = &storage.Blob{Value: v[i]}
	}

	reader := storage.NewDeleteCodec()
	_, _, dData, err := reader.Deserialize(blobs)
	if err != nil {
		return nil, err
	}
	return &DeltalogIterator{data: dData, pkFieldID: pkFieldID, pkType: pkType, label: label}, nil
}

func (d *DeltalogIterator) HasNext() bool {
	return !d.isDisposed() && d.hasNext()
}

func (d *DeltalogIterator) Next() (*LabeledRowData, error) {
	if d.isDisposed() {
		return nil, ErrDisposed
	}

	if !d.hasNext() {
		return nil, ErrNoMoreRecord
	}

	row := &DeltalogRow{
		Pk:        d.data.Pks[d.pos],
		Timestamp: d.data.Tss[d.pos],
	}
	d.pos++

	return NewLabeledRowData(row, d.label), nil
}

func (d *DeltalogIterator) Dispose() {
	d.disposed.CompareAndSwap(false, true)
	d.disposedOnce.Do(func() {
		close(d.disposeCh)
	})
}

func (d *DeltalogIterator) hasNext() bool {
	return int64(d.pos) < d.data.RowCount
}

func (d *DeltalogIterator) isDisposed() bool {
	return d.disposed.Load()
}

func (d *DeltalogIterator) WaitForDisposed() <-chan struct{} {
	<-d.disposeCh
	notifyCh := make(chan struct{})
	close(notifyCh)
	return notifyCh
}

// DynamicIterator can add, remove iterators while iterating.
// A ReadOnly DynamicIterator cannot add new iterators. And readers of the iterator
// will wait forever if the iterator is NOT ReadyOnly.
//
// So whoever WRITES iterators into DynamicIterator should set it to
// ReadOnly when there're no iterators to add, so that the readers WON'T WAIT FOREVER.
//
// Why we need DynamicIterator? Now we store lots of binlog files for 1 segment, usually too many
// files that we could easily hit the S3 limit. So we need to downloading files in batch, ending with
// the need for DynamicIterator.
type DynamicIterator[T Iterator] struct {
	backlogs chan T
	active   Iterator

	readonly atomic.Bool
	disposed atomic.Bool
}

var _ InterDynamicIterator = &DynamicIterator[*BinlogIterator]{}
var _ InterDynamicIterator = &DynamicIterator[*DeltalogIterator]{}

func NewDynamicBinlogIterator() *DynamicIterator[*BinlogIterator] {

	iter := DynamicIterator[*BinlogIterator]{
		backlogs: make(chan *BinlogIterator, 100), // TODO configurable
	}
	return &iter
}

func NewDynamicDeltalogIterator() *DynamicIterator[*DeltalogIterator] {
	iter := DynamicIterator[*DeltalogIterator]{
		backlogs: make(chan *DeltalogIterator, 100), // TODO configurable
	}
	return &iter
}

func (d *DynamicIterator[T]) HasNext() bool {
	return !d.isDisposed() && d.hasNext()
}

func (d *DynamicIterator[T]) Next() (*LabeledRowData, error) {
	if d.isDisposed() {
		return nil, ErrDisposed
	}

	if !d.hasNext() {
		return nil, ErrNoMoreRecord
	}

	return d.active.Next()
}

func (d *DynamicIterator[T]) Dispose() {
	if d.isDisposed() {
		return
	}

	d.active.Dispose()

	for b := range d.backlogs {
		b.Dispose()
	}

	d.disposed.CompareAndSwap(false, true)
}

func (d *DynamicIterator[T]) Add(iter Iterator) {
	if d.ReadOnly() {
		log.Warn("Adding iterators to an readonly dynamic iterator")
		return
	}
	d.backlogs <- iter.(T)
}

func (d *DynamicIterator[T]) ReadOnly() bool {
	return d.readonly.Load()
}

func (d *DynamicIterator[T]) SetReadOnly() {
	d.readonly.CompareAndSwap(false, true)
}

// WaitForDisposed does nothing, do not use
func (d *DynamicIterator[T]) WaitForDisposed() <-chan struct{} {
	notifyCh := make(chan struct{})
	close(notifyCh)
	return notifyCh
}

func (d *DynamicIterator[T]) isDisposed() bool {
	return d.disposed.Load()
}

func (d *DynamicIterator[T]) empty() bool {
	// No active iterator or the active iterator doesn't have next
	// and nothing in the backlogs
	return (d.active == nil || !d.active.HasNext()) || len(d.backlogs) > 0
}

// waitForData waits for readonly or not empty
func (d *DynamicIterator[T]) waitForActive() bool {
	if d.active != nil {
		d.active.Dispose()
		d.active = nil
	}

	// d.active == nil here
	ticker := time.NewTicker(time.Second)
	timer := time.NewTimer(1 * time.Hour) // TODO configurable
	for !d.ReadOnly() && len(d.backlogs) == 0 {
		select {
		case <-ticker.C:
		case <-timer.C:
			log.Warn("Timeout after waiting for 1h")
			return false
		}
	}

	if d.ReadOnly() {
		return false
	}

	d.active = <-d.backlogs
	return true
}

func (d *DynamicIterator[T]) hasNext() bool {
	if d.active != nil && d.active.HasNext() {
		return true
	}
	return d.waitForActive()
}

// FilterIterator deals with insert binlogs and delta binlogs
// While iterating through insert data and delta data, FilterIterator
// filters the data by various Filters, for example, ExpireFilter, TimetravelFilter, and
// most importantly DeletionFilter.
//
// DeletionFilter is constructed from the delta data, thus FilterIterator needs to go through
// ALL delta data from a segment before processing insert data. See hasNext() for details.
type FilterIterator struct {
	disposed atomic.Bool

	addOnce sync.Once
	filters []Filter

	iIter *DynamicIterator[*BinlogIterator]
	dIter *DynamicIterator[*DeltalogIterator]

	deleted   map[interface{}]Timestamp
	nextCache *LabeledRowData
}

var _ InterDynamicIterator = (*FilterIterator)(nil)

func NewFilterIterator(travelTs, now Timestamp, ttl int64) *FilterIterator {
	i := FilterIterator{
		iIter:   NewDynamicBinlogIterator(),
		dIter:   NewDynamicDeltalogIterator(),
		deleted: make(map[interface{}]Timestamp),
	}

	i.filters = []Filter{
		TimeTravelFilter(travelTs),
		ExpireFilter(ttl, now),
	}
	return &i
}

func (f *FilterIterator) HasNext() bool {
	return !f.isDisposed() && f.hasNext()
}

func (f *FilterIterator) Next() (*LabeledRowData, error) {
	if f.isDisposed() {
		return nil, ErrDisposed
	}

	if !f.hasNext() {
		return nil, ErrNoMoreRecord
	}

	tmp := f.nextCache
	f.nextCache = nil
	return tmp, nil
}

func (f *FilterIterator) isDisposed() bool {
	return f.disposed.Load()
}

func (f *FilterIterator) hasNext() bool {
	if f.nextCache != nil {
		return true
	}

	var (
		filtered bool = true
		labeled  *LabeledRowData
	)

	// Get Next until the row isn't filtered
	for f.dIter.HasNext() && filtered {
		labeled, _ = f.dIter.Next()
		for _, filter := range f.filters {
			filtered = filter(labeled)
		}

		// Add delete data for physical deletion
		data, _ := labeled.data.(*DeltalogRow)
		if filtered {
			f.deleted[data.Pk.GetValue()] = data.Timestamp
		}
	}

	// Add a new deletion filter for insertLogs
	// when iterate deltaLogs done for the first time.
	if !f.dIter.HasNext() && filtered {
		f.addDeletionFilterOnce()
	}

	// Get Next until the row isn't filtered or no data
	for f.iIter.HasNext() && filtered {
		labeled, _ = f.iIter.Next()
		for _, filter := range f.filters {
			filtered = filter(labeled)
		}
	}

	if filtered {
		return false
	}

	f.nextCache = labeled
	return true
}

func (f *FilterIterator) addDeletionFilterOnce() {
	f.addOnce.Do(func() {
		f.filters = append(f.filters, DeletionFilter(f.deleted))
	})
}

func (f *FilterIterator) Dispose() {
	f.iIter.Dispose()
	f.dIter.Dispose()
	f.disposed.CompareAndSwap(false, true)
}

// WaitForDisposed does nothing, do not use
func (f *FilterIterator) WaitForDisposed() <-chan struct{} {
	notifyCh := make(chan struct{})
	close(notifyCh)
	return notifyCh
}

func (f *FilterIterator) Add(iter Iterator) {
	switch iter.(type) {
	case *BinlogIterator:
		f.iIter.Add(iter)
	case *DeltalogIterator:
		f.dIter.Add(iter)
		// TODO default, unexpected behaviour, should not happen
	}
}

func (f *FilterIterator) ReadOnly() bool {
	return f.iIter.ReadOnly() && f.dIter.ReadOnly()
}

func (f *FilterIterator) SetReadOnly() {
	if !f.dIter.ReadOnly() {
		f.dIter.SetReadOnly()
	} else {
		f.iIter.SetReadOnly()
	}
}

// TODO: Later
type PkIterator struct {
	disposed  atomic.Bool
	pk        storage.FieldData
	pKfieldID int64
	pkType    schemapb.DataType
	pos       int
}

// TODO: Later
type DynamicPkIterator struct {
	iterators  []*PkIterator
	tmpRecords []*LabeledRowData
	nextRecord *LabeledRowData
}
