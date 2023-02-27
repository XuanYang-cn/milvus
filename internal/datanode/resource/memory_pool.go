package resource

import (
	"sync"

	"github.com/cockroachdb/errors"
	"github.com/milvus-io/milvus/pkg/util/typeutil"
	"go.uber.org/atomic"
)

type UniqueID = typeutil.UniqueID

// MemoryPool fact, any change MUST follow the following fact:
// Sum(PreFetched.Values) + Available == TotalMemory, can be checked by check()
//
// The TotalMemory is changable, MemoryPool uses a lazy policy to shrink the TotalMemory by default
// DataNode is in charge of all the creation, shrink and grow of a memory pool
type MemoryPool struct {
	TotalMemory uint64
	Available   atomic.Uint64

	preFetchGuard sync.Mutex
	PreFetched    map[UniqueID]uint64 // fetchedID and memory size pre-fetched

	FetchID     atomic.Int64
	TargetTotal atomic.Uint64 // For dynamicly changing pool size
}

func NewMemoryPool(total uint64) *MemoryPool {
	p := MemoryPool{
		TotalMemory: total,
		PreFetched:  make(map[UniqueID]uint64),
	}

	p.TargetTotal.Store(total)
	p.Available.Store(total)
	p.FetchID.Store(100)
	return &p
}

func (p *MemoryPool) GetAvailableSize() uint64 {
	return p.Available.Load()
}

// PreFetch tries to prefetch the memory needed in the future.
// returns the fetchID
func (p *MemoryPool) PreFetch(size uint64) (int64, error) {
	if size == 0 {
		return 0, errors.Newf("Invalid input size=%d", size)
	}
	if size > p.GetAvailableSize() {
		return 0, errors.Newf("Insufficient free memory size to prefetch, need %d, available %d", size, p.GetAvailableSize())
	}

	fetchID := p.getID()
	p.preFetchGuard.Lock()
	p.PreFetched[fetchID] = size
	p.preFetchGuard.Unlock()

	p.Available.Sub(size)
	return fetchID, nil
}

func (p *MemoryPool) GetPreFetchSize(fetchID UniqueID) (uint64, error) {
	p.preFetchGuard.Lock()
	defer p.preFetchGuard.Unlock()
	if size, ok := p.PreFetched[fetchID]; ok {
		return size, nil
	}

	return 0, errors.Newf("Unknown fetchID in memory pool: %d", fetchID)
}

// Return returns the memory pre-fetched
func (p *MemoryPool) Return(fetchID UniqueID) error {
	p.preFetchGuard.Lock()
	defer p.preFetchGuard.Unlock()
	if size, ok := p.PreFetched[fetchID]; ok {
		delete(p.PreFetched, fetchID)
		p.Available.Add(size)
		return nil
	}

	// should not happen at all
	return errors.Newf("Unknown fetchID: %d to return", fetchID)
}

// Shrink submits a task to try to shrink the TotalMemory to `totalTarget`
// TODO
func (p *MemoryPool) Shrink(totalTarget uint64) error {
	return nil
}

// Grow submits a task to try to grow the TotalMemory to `totalTarget`
// TODO
func (p *MemoryPool) Grow(totalTarget uint64) error {
	return nil
}

func (p *MemoryPool) lazyShrinkPolicy()  {}
func (p *MemoryPool) forceShrinkPolicy() {}

func (p *MemoryPool) getID() UniqueID {
	p.FetchID.Add(1)
	return p.FetchID.Load()
}

func (p *MemoryPool) check() bool {
	var fetchedMem uint64
	p.preFetchGuard.Lock()
	for _, fetched := range p.PreFetched {
		fetchedMem += fetched
	}
	p.preFetchGuard.Unlock()

	return fetchedMem+p.GetAvailableSize() == p.TotalMemory
}
