// Licensed to the LF AI & Data foundation under one
// or more contributor license agreements. See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership. The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License. You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package compactor

import (
	"context"
	"fmt"
	"io"
	"math"
	"path/filepath"
	"testing"

	"github.com/apache/arrow/go/v17/arrow"
	"github.com/apache/arrow/go/v17/arrow/array"
	"github.com/apache/arrow/go/v17/arrow/memory"

	"github.com/milvus-io/milvus-proto/go-api/v3/commonpb"
	"github.com/milvus-io/milvus-proto/go-api/v3/schemapb"
	"github.com/milvus-io/milvus/internal/allocator"
	"github.com/milvus-io/milvus/internal/compaction"
	"github.com/milvus-io/milvus/internal/storage"
	"github.com/milvus-io/milvus/pkg/v3/common"
	"github.com/milvus-io/milvus/pkg/v3/objectstorage"
	"github.com/milvus-io/milvus/pkg/v3/proto/indexpb"
	"github.com/milvus-io/milvus/pkg/v3/util/conc"
)

const (
	benchmarkPKField        = int64(100)
	benchmarkNamespaceField = int64(101)
	benchmarkVarcharField   = int64(102)
	benchmarkFloatField     = int64(103)
	benchmarkVectorBase     = int64(200)
)

type mixCompactorBenchmarkLayout int

const (
	benchmarkInterleaved mixCompactorBenchmarkLayout = iota
	benchmarkPartialOverlap
	benchmarkDisjoint
)

type mixCompactorBenchmarkKey int

const (
	benchmarkInt64Key mixCompactorBenchmarkKey = iota
	benchmarkVarcharKey
	benchmarkNamespaceKey
)

type mixCompactorBenchmarkCase struct {
	name          string
	readers       int
	rowsPerReader int
	layout        mixCompactorBenchmarkLayout
	key           mixCompactorBenchmarkKey
	scalars       int
	vectors       int
	dim           int
	filterPercent int
	missingField  bool
}

// The retained suite is representative rather than Cartesian. Together these
// cases cover reader counts 1/2/8/30, disjoint/partial/interleaved ranges,
// int64/varchar/namespace keys, PK-only/wide/vector-heavy schemas, and
// 0/10/90-percent filtering. Each merge case is also driven through the real
// V2 and V3 MultiSegmentWriter path by BenchmarkMixCompactorRealWriter.
var mixCompactorBenchmarkCases = []mixCompactorBenchmarkCase{
	{name: "pk_only_one_reader", readers: 1, rowsPerReader: 8192, layout: benchmarkDisjoint, key: benchmarkInt64Key},
	{name: "wide_interleaved_two", readers: 2, rowsPerReader: 8192, layout: benchmarkInterleaved, key: benchmarkInt64Key, scalars: 16, vectors: 2, dim: 32},
	{name: "varchar_partial_eight_filter10", readers: 8, rowsPerReader: 2048, layout: benchmarkPartialOverlap, key: benchmarkVarcharKey, scalars: 4, vectors: 1, dim: 32, filterPercent: 10},
	{name: "namespace_disjoint_thirty", readers: 30, rowsPerReader: 512, layout: benchmarkDisjoint, key: benchmarkNamespaceKey, scalars: 2},
	{name: "vector_heavy_interleaved_eight", readers: 8, rowsPerReader: 1024, layout: benchmarkInterleaved, key: benchmarkInt64Key, scalars: 2, vectors: 6, dim: 64},
	{name: "missing_field_filter90", readers: 2, rowsPerReader: 8192, layout: benchmarkPartialOverlap, key: benchmarkInt64Key, scalars: 6, vectors: 1, dim: 32, filterPercent: 90, missingField: true},
}

type benchmarkRecordReader struct {
	record storage.Record
	done   bool
	active bool
}

func (r *benchmarkRecordReader) Next() (storage.Record, error) {
	if r.done {
		return nil, io.EOF
	}
	r.done = true
	r.record.Retain()
	r.active = true
	return r.record, nil
}

func (r *benchmarkRecordReader) Close() error {
	if r.active {
		r.record.Release()
		r.active = false
	}
	return nil
}

type benchmarkMaterializedRecord struct {
	base     storage.Record
	fieldID  storage.FieldID
	computed arrow.Array
}

func (r *benchmarkMaterializedRecord) Column(fieldID storage.FieldID) arrow.Array {
	if fieldID == r.fieldID {
		return r.computed
	}
	return r.base.Column(fieldID)
}

func (r *benchmarkMaterializedRecord) Len() int { return r.base.Len() }

func (r *benchmarkMaterializedRecord) Retain() {
	r.base.Retain()
	r.computed.Retain()
}

func (r *benchmarkMaterializedRecord) Release() {
	r.base.Release()
	r.computed.Release()
}

type benchmarkCountingWriter struct {
	rows int
}

func (w *benchmarkCountingWriter) Write(r storage.Record) error {
	w.rows += r.Len()
	return nil
}

func (w *benchmarkCountingWriter) GetWrittenUncompressed() uint64 { return 0 }

func (w *benchmarkCountingWriter) Close() error { return nil }

type benchmarkLocalBinlogIO struct {
	storage.ChunkManager
	pool *conc.Pool[any]
}

func newBenchmarkLocalBinlogIO(root string) *benchmarkLocalBinlogIO {
	return &benchmarkLocalBinlogIO{
		ChunkManager: storage.NewLocalChunkManager(objectstorage.RootPath(root)),
		pool:         conc.NewPool[any](4),
	}
}

func (b *benchmarkLocalBinlogIO) Close() {
	b.pool.Release()
}

func (b *benchmarkLocalBinlogIO) Download(ctx context.Context, paths []string) ([][]byte, error) {
	values := make([][]byte, len(paths))
	for i, path := range paths {
		value, err := b.Read(ctx, path)
		if err != nil {
			return nil, err
		}
		values[i] = value
	}
	return values, nil
}

func (b *benchmarkLocalBinlogIO) AsyncDownload(ctx context.Context, paths []string) []*conc.Future[any] {
	values := make([]*conc.Future[any], 0, len(paths))
	for _, p := range paths {
		path := p
		values = append(values, b.pool.Submit(func() (any, error) { return b.Read(ctx, path) }))
	}
	return values
}

func (b *benchmarkLocalBinlogIO) Upload(ctx context.Context, kvs map[string][]byte) error {
	for path, value := range kvs {
		if err := b.Write(ctx, path, value); err != nil {
			return err
		}
	}
	return nil
}

func (b *benchmarkLocalBinlogIO) AsyncUpload(ctx context.Context, kvs map[string][]byte) []*conc.Future[any] {
	values := make([]*conc.Future[any], 0, len(kvs))
	for p, v := range kvs {
		path, value := p, v
		values = append(values, b.pool.Submit(func() (any, error) {
			return struct{}{}, b.Write(ctx, path, value)
		}))
	}
	return values
}

func warmBenchmarkV2Writer(b *testing.B, schema *schemapb.CollectionSchema, record storage.Record, params compaction.Params) {
	b.Helper()
	warmRoot := filepath.Join(b.TempDir(), "v2-warmup")
	params.StorageConfig = &indexpb.StorageConfig{StorageType: "local", RootPath: warmRoot}
	binlogIO := newBenchmarkLocalBinlogIO(warmRoot)
	defer binlogIO.Close()
	writer, err := NewMultiSegmentWriter(context.Background(), binlogIO,
		NewCompactionAllocator(allocator.NewLocalAllocator(1, math.MaxInt64), allocator.NewLocalAllocator(1000, math.MaxInt64)),
		256*1024*1024, schema, params, int64(record.Len()), 1, 1, "benchmark-warmup", 4096,
		storage.WithStorageConfig(params.StorageConfig), storage.WithVersion(storage.StorageV2))
	if err == nil {
		err = writer.Write(record)
	}
	if err == nil {
		err = writer.Close()
	}
	if err != nil {
		b.Fatal(err)
	}
}

type benchmarkSelectionMaterializer struct {
	base              storage.RecordReader
	selectionSchema   *schemapb.CollectionSchema
	filterPercent     int
	missingFieldID    int64
	materializeBefore bool
	current           storage.Record
}

func (r *benchmarkSelectionMaterializer) Next() (storage.Record, error) {
	if r.current != nil {
		r.current.Release()
		r.current = nil
	}
	base, err := r.base.Next()
	if err != nil {
		return nil, err
	}
	if r.filterPercent == 0 && r.missingFieldID == 0 {
		return base, nil
	}
	if r.materializeBefore && r.missingFieldID != 0 {
		base = materializeBenchmarkMissingField(base, r.missingFieldID, false)
	}
	if r.filterPercent > 0 {
		builder := storage.NewRecordBuilder(r.selectionSchema)
		for i := 0; i < base.Len(); i++ {
			if benchmarkRowKept(i, r.filterPercent) {
				if err := builder.Append(base, i, i+1); err != nil {
					builder.Release()
					if _, ok := base.(*benchmarkMaterializedRecord); ok {
						base.Release()
					}
					return nil, err
				}
			}
		}
		selected := builder.Build()
		builder.Release()
		if _, ok := base.(*benchmarkMaterializedRecord); ok {
			base.Release()
		}
		base = selected
	}
	if !r.materializeBefore && r.missingFieldID != 0 {
		base = materializeBenchmarkMissingField(base, r.missingFieldID, true)
	}
	r.current = base
	return base, nil
}

func (r *benchmarkSelectionMaterializer) Close() error {
	if r.current != nil {
		r.current.Release()
		r.current = nil
	}
	return r.base.Close()
}

func benchmarkRowKept(row, filterPercent int) bool {
	return row%100 >= filterPercent
}

func materializeBenchmarkMissingField(base storage.Record, fieldID int64, ownBase bool) storage.Record {
	if !ownBase {
		// The source reader owns its current reference. Keep one extra reference
		// for the wrapper so the borrowed source remains valid until cleanup.
		base.Retain()
	}
	builder := array.NewStringBuilder(memory.DefaultAllocator)
	builder.Reserve(base.Len())
	for i := 0; i < base.Len(); i++ {
		builder.Append(fmt.Sprintf("computed-%08d", i))
	}
	computed := builder.NewArray()
	builder.Release()
	return &benchmarkMaterializedRecord{base: base, fieldID: fieldID, computed: computed}
}

func benchmarkMergeKeys(tc mixCompactorBenchmarkCase) []int64 {
	switch tc.key {
	case benchmarkNamespaceKey:
		return []int64{benchmarkNamespaceField, benchmarkPKField}
	case benchmarkVarcharKey:
		return []int64{benchmarkVarcharField}
	default:
		return []int64{benchmarkPKField}
	}
}

func benchmarkKeyValue(tc mixCompactorBenchmarkCase, reader, row int) int64 {
	switch tc.layout {
	case benchmarkDisjoint:
		return int64(reader*tc.rowsPerReader + row)
	case benchmarkPartialOverlap:
		return int64(reader*(tc.rowsPerReader/2) + row)
	default:
		return int64(row*tc.readers + reader)
	}
}

func benchmarkSchema(tc mixCompactorBenchmarkCase) *schemapb.CollectionSchema {
	fields := []*schemapb.FieldSchema{
		{FieldID: common.RowIDField, Name: "row_id", DataType: schemapb.DataType_Int64},
		{FieldID: common.TimeStampField, Name: "timestamp", DataType: schemapb.DataType_Int64},
		{FieldID: benchmarkPKField, Name: "pk", DataType: schemapb.DataType_Int64, IsPrimaryKey: true},
	}
	if tc.key == benchmarkNamespaceKey {
		fields = append(fields, &schemapb.FieldSchema{FieldID: benchmarkNamespaceField, Name: "namespace", DataType: schemapb.DataType_Int64, IsPartitionKey: true})
	}
	if tc.key == benchmarkVarcharKey {
		fields = append(fields, &schemapb.FieldSchema{
			FieldID: benchmarkVarcharField, Name: "varchar_key", DataType: schemapb.DataType_VarChar,
			TypeParams: []*commonpb.KeyValuePair{{Key: common.MaxLengthKey, Value: "64"}},
		})
	}
	for i := 0; i < tc.scalars; i++ {
		fields = append(fields, &schemapb.FieldSchema{FieldID: benchmarkFloatField + int64(i), Name: fmt.Sprintf("scalar_%02d", i), DataType: schemapb.DataType_Double})
	}
	for i := 0; i < tc.vectors; i++ {
		fields = append(fields, &schemapb.FieldSchema{
			FieldID: benchmarkVectorBase + int64(i), Name: fmt.Sprintf("vector_%02d", i), DataType: schemapb.DataType_FloatVector,
			TypeParams: []*commonpb.KeyValuePair{{Key: common.DimKey, Value: fmt.Sprint(tc.dim)}},
		})
	}
	if tc.missingField {
		fields = append(fields, &schemapb.FieldSchema{
			FieldID: benchmarkVectorBase + 100, Name: "materialized", DataType: schemapb.DataType_VarChar,
			TypeParams: []*commonpb.KeyValuePair{{Key: common.MaxLengthKey, Value: "64"}},
		})
	}
	return &schemapb.CollectionSchema{Name: tc.name, EnableNamespace: tc.key == benchmarkNamespaceKey, Fields: fields}
}

func benchmarkRecords(b *testing.B, tc mixCompactorBenchmarkCase, schema *schemapb.CollectionSchema) []storage.Record {
	b.Helper()
	inputSchema := schema
	if tc.missingField {
		inputSchema = &schemapb.CollectionSchema{Name: schema.GetName(), EnableNamespace: schema.GetEnableNamespace(), Fields: schema.GetFields()[:len(schema.GetFields())-1]}
	}
	arrowSchema, err := storage.ConvertToArrowSchema(inputSchema, false)
	if err != nil {
		b.Fatal(err)
	}
	inputFields := inputSchema.GetFields()
	records := make([]storage.Record, tc.readers)
	for reader := 0; reader < tc.readers; reader++ {
		builders := make([]array.Builder, len(inputFields))
		for i, field := range inputFields {
			switch field.GetDataType() {
			case schemapb.DataType_Int64:
				builders[i] = array.NewInt64Builder(memory.DefaultAllocator)
			case schemapb.DataType_VarChar:
				builders[i] = array.NewStringBuilder(memory.DefaultAllocator)
			case schemapb.DataType_Double:
				builders[i] = array.NewFloat64Builder(memory.DefaultAllocator)
			case schemapb.DataType_FloatVector:
				builders[i] = array.NewFixedSizeBinaryBuilder(memory.DefaultAllocator, &arrow.FixedSizeBinaryType{ByteWidth: tc.dim * 4})
			default:
				b.Fatalf("unsupported benchmark field type %s", field.GetDataType())
			}
			builders[i].Reserve(tc.rowsPerReader)
		}
		vectorValue := make([]byte, tc.dim*4)
		for row := 0; row < tc.rowsPerReader; row++ {
			key := benchmarkKeyValue(tc, reader, row)
			for i, field := range inputFields {
				switch field.GetFieldID() {
				case common.RowIDField:
					builders[i].(*array.Int64Builder).Append(key)
				case common.TimeStampField:
					builders[i].(*array.Int64Builder).Append(1)
				case benchmarkPKField:
					builders[i].(*array.Int64Builder).Append(key)
				case benchmarkNamespaceField:
					builders[i].(*array.Int64Builder).Append(int64(reader / 2))
				case benchmarkVarcharField:
					builders[i].(*array.StringBuilder).Append(fmt.Sprintf("%032d", key))
				default:
					switch builder := builders[i].(type) {
					case *array.Float64Builder:
						builder.Append(float64(key) + float64(field.GetFieldID())/1000)
					case *array.FixedSizeBinaryBuilder:
						vectorValue[0] = byte(key)
						vectorValue[len(vectorValue)-1] = byte(key >> 8)
						builder.Append(vectorValue)
					}
				}
			}
		}
		arrays := make([]arrow.Array, len(builders))
		field2Col := make(map[storage.FieldID]int, len(builders))
		for i, builder := range builders {
			arrays[i] = builder.NewArray()
			builder.Release()
			field2Col[inputFields[i].GetFieldID()] = i
		}
		arrowRecord := array.NewRecord(arrowSchema, arrays, int64(tc.rowsPerReader))
		for _, values := range arrays {
			values.Release()
		}
		records[reader] = storage.NewSimpleArrowRecord(arrowRecord, field2Col)
	}
	return records
}

func benchmarkReaders(tc mixCompactorBenchmarkCase, schema *schemapb.CollectionSchema, records []storage.Record, materializeBefore bool) []storage.RecordReader {
	readers := make([]storage.RecordReader, len(records))
	selectionSchema := schema
	missingFieldID := int64(0)
	if tc.missingField {
		missingFieldID = benchmarkVectorBase + 100
		if !materializeBefore {
			selectionSchema = &schemapb.CollectionSchema{Name: schema.GetName(), EnableNamespace: schema.GetEnableNamespace(), Fields: schema.GetFields()[:len(schema.GetFields())-1]}
		}
	}
	for i, record := range records {
		base := storage.RecordReader(&benchmarkRecordReader{record: record})
		if tc.filterPercent > 0 || missingFieldID != 0 {
			base = &benchmarkSelectionMaterializer{
				base: base, selectionSchema: selectionSchema, filterPercent: tc.filterPercent,
				missingFieldID: missingFieldID, materializeBefore: materializeBefore,
			}
		}
		readers[i] = base
	}
	return readers
}

func benchmarkSourceBytes(records []storage.Record, schema *schemapb.CollectionSchema) int64 {
	var total uint64
	for _, record := range records {
		for _, field := range schema.GetFields() {
			col := record.Column(field.GetFieldID())
			if col != nil {
				total += storage.ActualSizeInBytes(col.Data())
			}
		}
	}
	return int64(total)
}

func closeBenchmarkReaders(readers []storage.RecordReader) {
	for _, reader := range readers {
		_ = reader.Close()
	}
}

func benchmarkPhaseRows(tc mixCompactorBenchmarkCase) int {
	return tc.readers * tc.rowsPerReader * (100 - tc.filterPercent) / 100
}

func BenchmarkMixCompactorPhases(b *testing.B) {
	tc := mixCompactorBenchmarkCases[1]
	schema := benchmarkSchema(tc)
	records := benchmarkRecords(b, tc, schema)
	defer func() {
		for _, record := range records {
			record.Release()
		}
	}()
	expectedRows := benchmarkPhaseRows(tc)

	b.Run("reader_decode", func(b *testing.B) {
		b.ReportAllocs()
		b.SetBytes(benchmarkSourceBytes(records, schema))
		for i := 0; i < b.N; i++ {
			readers := benchmarkReaders(tc, schema, records, false)
			rows := 0
			for _, reader := range readers {
				record, err := reader.Next()
				if err != nil {
					b.Fatal(err)
				}
				rows += record.Len()
			}
			closeBenchmarkReaders(readers)
			if rows != expectedRows {
				b.Fatalf("decoded rows=%d expected=%d", rows, expectedRows)
			}
		}
	})

	b.Run("predicate_selection", func(b *testing.B) {
		filtered := mixCompactorBenchmarkCases[len(mixCompactorBenchmarkCases)-1]
		filteredSchema := benchmarkSchema(filtered)
		filteredRecords := benchmarkRecords(b, filtered, filteredSchema)
		defer func() {
			for _, record := range filteredRecords {
				record.Release()
			}
		}()
		expected := benchmarkPhaseRows(filtered)
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			readers := benchmarkReaders(filtered, filteredSchema, filteredRecords, false)
			rows := 0
			for _, reader := range readers {
				record, err := reader.Next()
				if err != nil {
					b.Fatal(err)
				}
				rows += record.Len()
			}
			closeBenchmarkReaders(readers)
			if rows != expected {
				b.Fatalf("selected rows=%d expected=%d", rows, expected)
			}
		}
	})

	b.Run("merge_output", func(b *testing.B) {
		b.ReportAllocs()
		b.SetBytes(benchmarkSourceBytes(records, schema))
		for i := 0; i < b.N; i++ {
			readers := benchmarkReaders(tc, schema, records, false)
			writer := &benchmarkCountingWriter{}
			rows, err := storage.MergeSort(64*1024*1024, schema, readers, writer,
				func(storage.Record, int, int) bool { return true }, benchmarkMergeKeys(tc))
			closeBenchmarkReaders(readers)
			if err != nil {
				b.Fatal(err)
			}
			if rows != expectedRows || writer.rows != expectedRows {
				b.Fatalf("row mismatch: merge=%d writer=%d expected=%d", rows, writer.rows, expectedRows)
			}
		}
	})
}

func BenchmarkMixCompactorMergeCore(b *testing.B) {
	for _, tc := range mixCompactorBenchmarkCases {
		tc := tc
		b.Run(tc.name, func(b *testing.B) {
			schema := benchmarkSchema(tc)
			records := benchmarkRecords(b, tc, schema)
			defer func() {
				for _, record := range records {
					record.Release()
				}
			}()
			b.ReportAllocs()
			b.SetBytes(benchmarkSourceBytes(records, schema))
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				readers := benchmarkReaders(tc, schema, records, false)
				writer := &benchmarkCountingWriter{}
				rows, err := storage.MergeSort(64*1024*1024, schema, readers, writer,
					func(storage.Record, int, int) bool { return true }, benchmarkMergeKeys(tc))
				closeBenchmarkReaders(readers)
				if err != nil {
					b.Fatal(err)
				}
				expectedRows := benchmarkPhaseRows(tc)
				if rows != expectedRows || writer.rows != expectedRows {
					b.Fatalf("row mismatch: merge=%d writer=%d expected=%d", rows, writer.rows, expectedRows)
				}
			}
		})
	}
}

func BenchmarkMixCompactorSelectionBeforeMaterialization(b *testing.B) {
	tc := mixCompactorBenchmarkCases[len(mixCompactorBenchmarkCases)-1]
	schema := benchmarkSchema(tc)
	records := benchmarkRecords(b, tc, schema)
	defer func() {
		for _, record := range records {
			record.Release()
		}
	}()
	for _, materializeBefore := range []bool{true, false} {
		name := "select_first"
		if materializeBefore {
			name = "materialize_first"
		}
		b.Run(name, func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				readers := benchmarkReaders(tc, schema, records, materializeBefore)
				writer := &benchmarkCountingWriter{}
				rows, err := storage.MergeSort(64*1024*1024, schema, readers, writer,
					func(storage.Record, int, int) bool { return true }, benchmarkMergeKeys(tc))
				closeBenchmarkReaders(readers)
				if err != nil {
					b.Fatal(err)
				}
				expectedRows := benchmarkPhaseRows(tc)
				if rows != expectedRows || writer.rows != expectedRows {
					b.Fatalf("row mismatch: merge=%d writer=%d expected=%d", rows, writer.rows, expectedRows)
				}
			}
		})
	}
}

func BenchmarkMixCompactorRealWriter(b *testing.B) {
	cases := []mixCompactorBenchmarkCase{
		mixCompactorBenchmarkCases[0],
		mixCompactorBenchmarkCases[1],
		mixCompactorBenchmarkCases[2],
		mixCompactorBenchmarkCases[3],
		mixCompactorBenchmarkCases[4],
		mixCompactorBenchmarkCases[5],
	}
	for _, version := range []int64{storage.StorageV2, storage.StorageV3} {
		versionName := fmt.Sprintf("v%d", version)
		for _, tc := range cases {
			tc := tc
			b.Run(versionName+"/"+tc.name, func(b *testing.B) {
				schema := benchmarkSchema(tc)
				records := benchmarkRecords(b, tc, schema)
				defer func() {
					for _, record := range records {
						record.Release()
					}
				}()
				root := b.TempDir()
				params := compaction.Params{
					StorageVersion: version,
					BinLogMaxSize:  64 * 1024 * 1024,
					StorageConfig:  &indexpb.StorageConfig{StorageType: "local", RootPath: root},
				}
				if version == storage.StorageV2 {
					warmBenchmarkV2Writer(b, schema, records[0], params)
				}
				b.ReportAllocs()
				b.SetBytes(benchmarkSourceBytes(records, schema))
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					iterationRoot := filepath.Join(root, fmt.Sprintf("iteration-%d", i))
					params.StorageConfig.RootPath = iterationRoot
					binlogIO := newBenchmarkLocalBinlogIO(iterationRoot)
					readers := benchmarkReaders(tc, schema, records, false)
					segmentAllocator := allocator.NewLocalAllocator(int64(i+1)*1_000_000, math.MaxInt64)
					logAllocator := allocator.NewLocalAllocator(int64(i+1)*10_000_000, math.MaxInt64)
					writer, err := NewMultiSegmentWriter(context.Background(), binlogIO,
						NewCompactionAllocator(segmentAllocator, logAllocator), 256*1024*1024,
						schema, params, int64(tc.readers*tc.rowsPerReader), 1, 1, "benchmark", 4096,
						storage.WithStorageConfig(params.StorageConfig), storage.WithVersion(version))
					rows := 0
					if err == nil {
						rows, err = storage.MergeSort(params.BinLogMaxSize, schema, readers, writer,
							func(storage.Record, int, int) bool { return true }, benchmarkMergeKeys(tc))
					}
					if writer != nil {
						if closeErr := writer.Close(); err == nil {
							err = closeErr
						}
					}
					closeBenchmarkReaders(readers)
					binlogIO.Close()
					if err != nil {
						b.Fatal(err)
					}
					expectedRows := benchmarkPhaseRows(tc)
					writtenRows := int64(0)
					for _, segment := range writer.GetCompactionSegments() {
						writtenRows += segment.GetNumOfRows()
					}
					if rows != expectedRows || writtenRows != int64(expectedRows) {
						b.Fatalf("row mismatch: merge=%d writer=%d expected=%d", rows, writtenRows, expectedRows)
					}
				}
			})
		}
	}
}
