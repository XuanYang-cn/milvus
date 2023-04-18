// Licensed to the LF AI & Data foundation under one
// or more contributor license agreements. See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership. The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License. You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package meta

import (
	"context"
	"testing"

	"github.com/milvus-io/milvus/internal/proto/etcdpb"
	"github.com/milvus-io/milvus/internal/proto/rootcoordpb"
	"github.com/milvus-io/milvus/internal/types"

	"github.com/cockroachdb/errors"
	"github.com/milvus-io/milvus-proto/go-api/commonpb"
	"github.com/milvus-io/milvus-proto/go-api/milvuspb"
	"github.com/milvus-io/milvus-proto/go-api/schemapb"
	"github.com/stretchr/testify/assert"
)

const (
	collectionID0   = UniqueID(2)
	collectionID1   = UniqueID(1)
	collectionName0 = "collection_0"
	collectionName1 = "collection_1"
)

func TestMetaService_All(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mFactory := &RootCoordFactory{
		pkType: schemapb.DataType_Int64,
	}
	mFactory.setCollectionID(collectionID0)
	mFactory.setCollectionName(collectionName0)
	ms := newMetaService(mFactory, collectionID0)

	t.Run("Test getCollectionSchema", func(t *testing.T) {

		sch, err := ms.getCollectionSchema(ctx, collectionID0, 0)
		assert.NoError(t, err)
		assert.NotNil(t, sch)
		assert.Equal(t, sch.Name, collectionName0)
	})

	t.Run("Test printCollectionStruct", func(t *testing.T) {
		mf := &MetaFactory{}
		collectionMeta := mf.GetCollectionMeta(collectionID0, collectionName0, schemapb.DataType_Int64)
		printCollectionStruct(collectionMeta)
	})
}

// RootCoordFails1 root coord mock for failure
type RootCoordFails1 struct {
	RootCoordFactory
}

// DescribeCollectionInternal override method that will fails
func (rc *RootCoordFails1) DescribeCollectionInternal(ctx context.Context, req *milvuspb.DescribeCollectionRequest) (*milvuspb.DescribeCollectionResponse, error) {
	return nil, errors.New("always fail")
}

// RootCoordFails2 root coord mock for failure
type RootCoordFails2 struct {
	RootCoordFactory
}

// DescribeCollectionInternal override method that will fails
func (rc *RootCoordFails2) DescribeCollectionInternal(ctx context.Context, req *milvuspb.DescribeCollectionRequest) (*milvuspb.DescribeCollectionResponse, error) {
	return &milvuspb.DescribeCollectionResponse{
		Status: &commonpb.Status{ErrorCode: commonpb.ErrorCode_UnexpectedError},
	}, nil
}

func TestMetaServiceRootCoodFails(t *testing.T) {

	t.Run("Test Describe with error", func(t *testing.T) {
		rc := &RootCoordFails1{}
		rc.setCollectionID(collectionID0)
		rc.setCollectionName(collectionName0)

		ms := newMetaService(rc, collectionID0)
		_, err := ms.getCollectionSchema(context.Background(), collectionID1, 0)
		assert.NotNil(t, err)
	})

	t.Run("Test Describe wit nil response", func(t *testing.T) {
		rc := &RootCoordFails2{}
		rc.setCollectionID(collectionID0)
		rc.setCollectionName(collectionName0)

		ms := newMetaService(rc, collectionID0)
		_, err := ms.getCollectionSchema(context.Background(), collectionID1, 0)
		assert.NotNil(t, err)
	})
}

type RootCoordFactory struct {
	types.RootCoord
	ID             UniqueID
	collectionName string
	collectionID   UniqueID
	pkType         schemapb.DataType

	ReportImportErr        bool
	ReportImportNotSuccess bool
}

// If id == 0, AllocID will return not successful status
// If id == -1, AllocID will return err
func (m *RootCoordFactory) setID(id UniqueID) {
	m.ID = id // GOOSE TODO: random ID generator
}

func (m *RootCoordFactory) setCollectionID(id UniqueID) {
	m.collectionID = id
}

func (m *RootCoordFactory) setCollectionName(name string) {
	m.collectionName = name
}

func (m *RootCoordFactory) AllocID(ctx context.Context, in *rootcoordpb.AllocIDRequest) (*rootcoordpb.AllocIDResponse, error) {
	resp := &rootcoordpb.AllocIDResponse{
		Status: &commonpb.Status{
			ErrorCode: commonpb.ErrorCode_UnexpectedError,
		}}

	if in.Count == 12 {
		resp.Status.ErrorCode = commonpb.ErrorCode_Success
		resp.ID = 1
		resp.Count = 12
	}

	if m.ID == 0 {
		resp.Status.Reason = "Zero ID"
		return resp, nil
	}

	if m.ID == -1 {
		return nil, errors.New(resp.Status.GetReason())
	}

	resp.ID = m.ID
	resp.Count = in.GetCount()
	resp.Status.ErrorCode = commonpb.ErrorCode_Success
	return resp, nil
}

func (m *RootCoordFactory) AllocTimestamp(ctx context.Context, in *rootcoordpb.AllocTimestampRequest) (*rootcoordpb.AllocTimestampResponse, error) {
	resp := &rootcoordpb.AllocTimestampResponse{
		Status:    &commonpb.Status{},
		Timestamp: 1000,
	}

	return resp, nil
}

func (m *RootCoordFactory) ShowCollections(ctx context.Context, in *milvuspb.ShowCollectionsRequest) (*milvuspb.ShowCollectionsResponse, error) {
	resp := &milvuspb.ShowCollectionsResponse{
		Status:          &commonpb.Status{},
		CollectionNames: []string{m.collectionName},
	}
	return resp, nil

}

func (m *RootCoordFactory) DescribeCollectionInternal(ctx context.Context, in *milvuspb.DescribeCollectionRequest) (*milvuspb.DescribeCollectionResponse, error) {
	f := MetaFactory{}
	meta := f.GetCollectionMeta(m.collectionID, m.collectionName, m.pkType)
	resp := &milvuspb.DescribeCollectionResponse{
		Status: &commonpb.Status{
			ErrorCode: commonpb.ErrorCode_UnexpectedError,
		},
	}

	if m.collectionID == -2 {
		resp.Status.Reason = "Status not success"
		return resp, nil
	}

	if m.collectionID == -1 {
		resp.Status.ErrorCode = commonpb.ErrorCode_Success
		return resp, errors.New(resp.Status.GetReason())
	}

	resp.CollectionID = m.collectionID
	resp.Schema = meta.Schema
	resp.ShardsNum = 2
	resp.Status.ErrorCode = commonpb.ErrorCode_Success
	return resp, nil
}

func (m *RootCoordFactory) GetComponentStates(ctx context.Context) (*milvuspb.ComponentStates, error) {
	return &milvuspb.ComponentStates{
		State:              &milvuspb.ComponentInfo{},
		SubcomponentStates: make([]*milvuspb.ComponentInfo, 0),
		Status: &commonpb.Status{
			ErrorCode: commonpb.ErrorCode_Success,
		},
	}, nil
}

func (m *RootCoordFactory) ReportImport(ctx context.Context, req *rootcoordpb.ImportResult) (*commonpb.Status, error) {
	return &commonpb.Status{
		ErrorCode: commonpb.ErrorCode_Success,
	}, nil
}

type MetaFactory struct {
}

func (mf *MetaFactory) GetCollectionMeta(collectionID UniqueID, collectionName string, pkDataType schemapb.DataType) *etcdpb.CollectionMeta {
	sch := schemapb.CollectionSchema{
		Name:        collectionName,
		Description: "test collection by meta factory",
		AutoID:      true,
	}
	sch.Fields = mf.GetFieldSchema()
	for _, field := range sch.Fields {
		if field.GetDataType() == pkDataType && field.FieldID >= 100 {
			field.IsPrimaryKey = true
		}
	}

	return &etcdpb.CollectionMeta{
		ID:           collectionID,
		Schema:       &sch,
		CreateTime:   Timestamp(1),
		SegmentIDs:   make([]UniqueID, 0),
		PartitionIDs: []UniqueID{0},
	}
}

func (mf *MetaFactory) GetFieldSchema() []*schemapb.FieldSchema {
	fields := []*schemapb.FieldSchema{
		{
			FieldID:     0,
			Name:        "RowID",
			Description: "RowID field",
			DataType:    schemapb.DataType_Int64,
			TypeParams: []*commonpb.KeyValuePair{
				{
					Key:   "f0_tk1",
					Value: "f0_tv1",
				},
			},
		},
		{
			FieldID:     1,
			Name:        "Timestamp",
			Description: "Timestamp field",
			DataType:    schemapb.DataType_Int64,
			TypeParams: []*commonpb.KeyValuePair{
				{
					Key:   "f1_tk1",
					Value: "f1_tv1",
				},
			},
		},
		{
			FieldID:     100,
			Name:        "float_vector_field",
			Description: "field 100",
			DataType:    schemapb.DataType_FloatVector,
			TypeParams: []*commonpb.KeyValuePair{
				{
					Key:   "dim",
					Value: "2",
				},
			},
			IndexParams: []*commonpb.KeyValuePair{
				{
					Key:   "indexkey",
					Value: "indexvalue",
				},
			},
		},
		{
			FieldID:     101,
			Name:        "binary_vector_field",
			Description: "field 101",
			DataType:    schemapb.DataType_BinaryVector,
			TypeParams: []*commonpb.KeyValuePair{
				{
					Key:   "dim",
					Value: "32",
				},
			},
			IndexParams: []*commonpb.KeyValuePair{
				{
					Key:   "indexkey",
					Value: "indexvalue",
				},
			},
		},
		{
			FieldID:     102,
			Name:        "bool_field",
			Description: "field 102",
			DataType:    schemapb.DataType_Bool,
			TypeParams:  []*commonpb.KeyValuePair{},
			IndexParams: []*commonpb.KeyValuePair{},
		},
		{
			FieldID:     103,
			Name:        "int8_field",
			Description: "field 103",
			DataType:    schemapb.DataType_Int8,
			TypeParams:  []*commonpb.KeyValuePair{},
			IndexParams: []*commonpb.KeyValuePair{},
		},
		{
			FieldID:     104,
			Name:        "int16_field",
			Description: "field 104",
			DataType:    schemapb.DataType_Int16,
			TypeParams:  []*commonpb.KeyValuePair{},
			IndexParams: []*commonpb.KeyValuePair{},
		},
		{
			FieldID:     105,
			Name:        "int32_field",
			Description: "field 105",
			DataType:    schemapb.DataType_Int32,
			TypeParams:  []*commonpb.KeyValuePair{},
			IndexParams: []*commonpb.KeyValuePair{},
		},
		{
			FieldID:     106,
			Name:        "int64_field",
			Description: "field 106",
			DataType:    schemapb.DataType_Int64,
			TypeParams:  []*commonpb.KeyValuePair{},
			IndexParams: []*commonpb.KeyValuePair{},
		},
		{
			FieldID:     107,
			Name:        "float32_field",
			Description: "field 107",
			DataType:    schemapb.DataType_Float,
			TypeParams:  []*commonpb.KeyValuePair{},
			IndexParams: []*commonpb.KeyValuePair{},
		},
		{
			FieldID:     108,
			Name:        "float64_field",
			Description: "field 108",
			DataType:    schemapb.DataType_Double,
			TypeParams:  []*commonpb.KeyValuePair{},
			IndexParams: []*commonpb.KeyValuePair{},
		},
		{
			FieldID:     109,
			Name:        "varChar_field",
			Description: "field 109",
			DataType:    schemapb.DataType_VarChar,
			TypeParams: []*commonpb.KeyValuePair{
				{
					Key:   "max_length",
					Value: "100",
				},
			},
			IndexParams: []*commonpb.KeyValuePair{},
		},
	}

	return fields
}
