# Compaction design docs

See benchmarks at [BENCH.md](./BENCH.md)

## Usage
```golang
import "github.com/milvus-io/milvus/internal/datanode/compaction"
factory := compaction.NewCompactionFactory(binlogIO, resources, allocator)
go factory.Start(ctx)

err := factory.CreateWorkLine(plan, meta, partitionID)
```

## About memory control

We knows the system's free memory, used memory, used cpu, but those're not enough
to schedule and execute compaction tasks, especially when we want dynamics.

**Proposing a memory pool for compaction**

1. DataNode controls the memory pool size according to the **load** (How to define the load of a DN?)
    - The first policy: static (50, 50)
    - The second policy: dynamic (20, 80) according to configured time range
2. Compaction factory controls the usage of the pool
    - Estimate the memory usage of a compaction task
    - Access the memory from the memory pool
    - Executing the Compaction
    - Return the memory from the pool

A memory pool for flowgraphs?????

```golang
type MemoryPool struct{}

func (p *MemoryPool) PreFetch(size uint64) (int64, error)
func (p *MemoryPool) Return(ID int64) error
func (p *MemoryPool) Shrink(totalTarget uint64) error
func (p *MemoryPool) Grow(totalTarget uint64) error
```

## Inventory

```golang
// Provide Workers that can download binlogs according to resources and
// compaction types
type Inventory struct {
    resources   *resource.GlobalResources
    binlogIO    *io.BinlogIO
}

func (i *Inventory) ProvideWorker(plan *datapb.CompactionPlan) Worker {
    i.available()

    worker := NewWorker(plan.GetType())
    return worker
}
func (i *Inventory) ReleaseWorker() {}

func (i *Inventory) available(){ }
```

### Worker
```golang
type Worker interface {
    Work()
    Close()
}

type StatsLogWorker struct {}
type InsertLogWorker struct {}
type DeltaLogWorker struct {}

// MixedWorker can download deltalogs and binlogs
type MixedWorker struct {
    cartGuard    sync.Mutex
    segmentCarts map[UniqueID]*SegmentCart
}
func (w *MixedWorker) Work() {
    for _, cart := range carts {
        fillDeltaLogs(cart)
        fillBinlogs(cart)
    }
}

func NewWorker() Worker {
    if planType == MixedCompaction {
        return &MixWorker{}
    }
}
```

## CompactionWorkLine

```golang
type CompactionWorkLine struct{
    inWorker   Worker
    packer     Packer
    Dispatcher Dispatcher
}

func (l *CompactionWorkLine) Assemble(inWorker Worker, packer Packer)
func (l *CompactionWorkLine) Start() {
    go l.Dispactcher.Dispatch()
    go l.inWorker.Work()
    go l.packer.Pack()
}
```

## Depository
```golang
type Depository struct{}

func (d *Depository) ProvidePacker(lin *CompactionWorkLine) *Packer { }
```

### Packer
```golang
type Packer struct { }

func (p *Packer) Pack()
func (p *Packer) Close()
```
