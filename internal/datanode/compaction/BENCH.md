# Benchmarks for executing compaciton

> This is an early prototype benchmark to testify if the compaction algorithm is working as desgin

## Goals

- Compaction works correctly and fast
- Compaction can scale vertically
- Compaction works on limited resources

**IMPORTANT branches: select_policy with partition distribution, select_policy_2.2**

## Before Tests

Several works before starting benchmarks

- [x] [ManualCompactionEnhanceCommit]Manual Compaction can provide target segments, no limits
    - Milvus: `xuanyang-cn/milvus@enhance-manual-compaction`
    - PyMilvus: `xuanyang-cn/pymilvus@enhance-manual-compaction`
    - milvus-proto: `git@github.com/xuanyang-cn/milvus-proto.git@enhance-manual-compaction`
- [] Test cases outline

## Test Cases

1. Test environment

```
Architecture:       x86_64
Processor:          Intel(R) Core(TM) i7-8700 CPU @ 3.20GHz
GPU:                GeForce GTX 1060 6GB * 2
Memory:             64GB
Disk:               256GB
```

2. Common Configs:
```
dataCoord.segment.maxSize: 512 # MB
dataCoord.segment.sealProportion: 2
dataCoord.compaction.enableAutoCompaction: false
```

```
# collection schema
NumShards: 1
Dimension: 128
Index: NONE
Partition: default
Fields:
    - pk:           Int64         auto_id=True, primary_key=True
    - embeddings:   FloatVector
```

### 1. Baseline tests, default behavior of current compaction@4.26.2023


- milvus: https://github.com/XuanYang-cn/milvus/tree/bench-baseline
- pymilvus: https://github.com/XuanYang-cn/pymilvus/tree/enhance-manual-compaction
- milvus-proto: https://github.com/XuanYang-cn/milvus-proto/tree/enhance-manual-compaction



#### 1.1 One compaction plan
|Distribution before compaction/MB|Distribution after compaction/MB|tr(execute compaction)/ms|
|:-------------------------------|:------------------------------:|:------------------------:|
|(512)|(512)||
|(256, 256)|(512)||
|(128, 128, 128, 128)|(512)||
|(64, 64, 64, 64, 64, 64, 64, 64)|(512)||
|(2048)|(2048)||
|(1024, 1024)|(2048)||
|(512, 512, 512, 512)|(2048)||
|(2048, 2048, 2048, 2048)|(8192)||

### 2. Enough memory for multiple Compactions
#### 2.1 n compactions, MemSumOf(n) <= TotalMem, linear execution, (tr, success)
#### 2.2 n compactions, MemSumOf(n) <= TotalMem, parallel execution, (tr, success)
#### 2.3 n compactions, MemSumOf(n) > TotalMem, parallel execution, (tr, success)
### 3. Not enough memory for ONE compaction
#### 3.1 one compaction, (tr, success)
#### 3.2 ten compaction, (tr, success)
## Test Results
