package compaction

import (
	"testing"

	"github.com/milvus-io/milvus/internal/proto/datapb"
	"github.com/stretchr/testify/assert"
)

func TestCalculateInsertBatchesByResources(t *testing.T) {
	binlogs := []*datapb.Binlog{
		&datapb.Binlog{LogPath: "p0", LogSize: 1},
		&datapb.Binlog{LogPath: "p2", LogSize: 1},
	}
	fieldBinlogs := []*datapb.FieldBinlog{
		&datapb.FieldBinlog{FieldID: 0, Binlogs: binlogs},
		&datapb.FieldBinlog{FieldID: 1, Binlogs: binlogs},
	}

	batches := CalculateInsertBatchesByResources(2, fieldBinlogs)
	assert.Equal(t, 2, len(batches))
	assert.True(t, batches[0].notEmpty())

	assert.Equal(t, 2, len(batches[0].paths))
	assert.Equal(t, 2, len(batches[1].paths))
}
