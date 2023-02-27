package compaction

import (
	"sync"

	"github.com/milvus-io/milvus/pkg/log"
)

type Dispatcher interface {
	Dispatch(waiter *sync.WaitGroup, errChan chan<- error)
	Done()
}

// make sure this labeled rowdata is valid before change label
func (d *LabeledRowData) ChangeSegment(segmentID UniqueID) {
	d.label.segmentID = segmentID
}

func (d *LabeledRowData) SetLabel(l *Label) {
	d.label = l
}

type PkRange struct {
	largestPk  interface{}
	samllestPk interface{}
	numRow     int64
}

type SingleDispatcher struct {
	in, out         chan *LabeledRowData
	targetSegmentID UniqueID

	doneOnce sync.Once
	done     chan struct{}
}

var _ Dispatcher = (*SingleDispatcher)(nil)

func NewDispatcher(in, out chan *LabeledRowData, segmentIDs ...UniqueID) Dispatcher {
	// TODO multi dispatcher
	// if len(segmentIDs) >= 2 {
	//     return NewMultiRandomDispatcher(segmentIDs, in, out)
	// }

	return NewSingleDispatcher(segmentIDs[0], in, out)
}

func NewSingleDispatcher(segmentID UniqueID, in, out chan *LabeledRowData) *SingleDispatcher {
	dispatcher := SingleDispatcher{
		in:              in,
		out:             out,
		targetSegmentID: segmentID,
		done:            make(chan struct{}),
	}

	return &dispatcher
}

func (s *SingleDispatcher) Dispatch(lineWaiter *sync.WaitGroup, errChan chan<- error) {
	defer lineWaiter.Done()
	for {
		select {
		case labeled, ok := <-s.in:
			if !ok {
				close(s.out)
				log.Info("SingleDispatcher finish dispatching")
				return
			}
			labeled.ChangeSegment(s.targetSegmentID)
			s.out <- labeled
		case <-s.done:
			log.Info("SingleDispatcher closed unexpectedly")
			return
		}
	}
}

func (s *SingleDispatcher) Done() {
	s.doneOnce.Do(func() {
		close(s.done)
	})
}

// type MultiRandomDispatcher struct {
//     in, out  chan *LabeledRowData
//     segments map[UniqueID]*PkRange // TODO
//
//     deleteBuf []*dData // TODO
//
//     doneOnce sync.Once
//     done     chan struct{}
// }
//
// var _ Dispatcher = (*MultiRandomDispatcher)(nil)
//
// func NewMultiRandomDispatcher(segmentIDs []UniqueID, in, out chan *LabeledRowData) *MultiRandomDispatcher {
//     return &MultiRandomDispatcher{
//         in:        in,
//         out:       out,
//         deleteBuf: make([]*dData, 0), // TODO
//         done:      make(chan struct{}),
//     }
// }
//
// func (m *MultiRandomDispatcher) Dispatch(lineWaiter *sync.WaitGroup) {
//     for {
//         select {
//         case labeled := <-m.in:
//             switch labeled.data.(type) {
//             case *BinlogRow:
//                 m.dispatchInsert(labeled)
//             case *DeltalogRow:
//                 m.dispatchDelete(labeled)
//             }
//         case <-m.done:
//             lineWaiter.Done()
//             return
//         }
//     }
// }
//
// func (m *MultiRandomDispatcher) Done() {
//     m.doneOnce.Do(func() {
//         close(m.done)
//     })
// }
//
// func (m *MultiRandomDispatcher) dispatchInsert(row *LabeledRowData) {
//     newLabel := row.label // TODO random pick
//     row.SetLabel(newLabel)
//     m.out <- row
// }
//
// func (m *MultiRandomDispatcher) dispatchDelete(row *LabeledRowData) {
//     // TODO store in buffer
//     // wait for insert done
//     // TODO: m.deleteBuf.Buff(row.data)
// }
