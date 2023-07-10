// Code generated by mockery v2.16.0. DO NOT EDIT.

package mocks

import (
	context "context"

	commonpb "github.com/milvus-io/milvus-proto/go-api/v2/commonpb"
	clientv3 "go.etcd.io/etcd/client/v3"

	datapb "github.com/milvus-io/milvus/internal/proto/datapb"

	internalpb "github.com/milvus-io/milvus/internal/proto/internalpb"

	milvuspb "github.com/milvus-io/milvus-proto/go-api/v2/milvuspb"

	mock "github.com/stretchr/testify/mock"

	types "github.com/milvus-io/milvus/internal/types"
)

// MockDataNode is an autogenerated mock type for the DataNodeComponent type
type MockDataNode struct {
	mock.Mock
}

type MockDataNode_Expecter struct {
	mock *mock.Mock
}

func (_m *MockDataNode) EXPECT() *MockDataNode_Expecter {
	return &MockDataNode_Expecter{mock: &_m.Mock}
}

// AddImportSegment provides a mock function with given fields: ctx, req
func (_m *MockDataNode) AddImportSegment(ctx context.Context, req *datapb.AddImportSegmentRequest) (*datapb.AddImportSegmentResponse, error) {
	ret := _m.Called(ctx, req)

	var r0 *datapb.AddImportSegmentResponse
	if rf, ok := ret.Get(0).(func(context.Context, *datapb.AddImportSegmentRequest) *datapb.AddImportSegmentResponse); ok {
		r0 = rf(ctx, req)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*datapb.AddImportSegmentResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *datapb.AddImportSegmentRequest) error); ok {
		r1 = rf(ctx, req)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockDataNode_AddImportSegment_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'AddImportSegment'
type MockDataNode_AddImportSegment_Call struct {
	*mock.Call
}

// AddImportSegment is a helper method to define mock.On call
//  - ctx context.Context
//  - req *datapb.AddImportSegmentRequest
func (_e *MockDataNode_Expecter) AddImportSegment(ctx interface{}, req interface{}) *MockDataNode_AddImportSegment_Call {
	return &MockDataNode_AddImportSegment_Call{Call: _e.mock.On("AddImportSegment", ctx, req)}
}

func (_c *MockDataNode_AddImportSegment_Call) Run(run func(ctx context.Context, req *datapb.AddImportSegmentRequest)) *MockDataNode_AddImportSegment_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*datapb.AddImportSegmentRequest))
	})
	return _c
}

func (_c *MockDataNode_AddImportSegment_Call) Return(_a0 *datapb.AddImportSegmentResponse, _a1 error) *MockDataNode_AddImportSegment_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

// CheckChannelOperationProgress provides a mock function with given fields: ctx, req
func (_m *MockDataNode) CheckChannelOperationProgress(ctx context.Context, req *datapb.ChannelWatchInfo) (*datapb.ChannelOperationProgressResponse, error) {
	ret := _m.Called(ctx, req)

	var r0 *datapb.ChannelOperationProgressResponse
	if rf, ok := ret.Get(0).(func(context.Context, *datapb.ChannelWatchInfo) *datapb.ChannelOperationProgressResponse); ok {
		r0 = rf(ctx, req)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*datapb.ChannelOperationProgressResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *datapb.ChannelWatchInfo) error); ok {
		r1 = rf(ctx, req)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockDataNode_CheckChannelOperationProgress_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'CheckChannelOperationProgress'
type MockDataNode_CheckChannelOperationProgress_Call struct {
	*mock.Call
}

// CheckChannelOperationProgress is a helper method to define mock.On call
//  - ctx context.Context
//  - req *datapb.ChannelWatchInfo
func (_e *MockDataNode_Expecter) CheckChannelOperationProgress(ctx interface{}, req interface{}) *MockDataNode_CheckChannelOperationProgress_Call {
	return &MockDataNode_CheckChannelOperationProgress_Call{Call: _e.mock.On("CheckChannelOperationProgress", ctx, req)}
}

func (_c *MockDataNode_CheckChannelOperationProgress_Call) Run(run func(ctx context.Context, req *datapb.ChannelWatchInfo)) *MockDataNode_CheckChannelOperationProgress_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*datapb.ChannelWatchInfo))
	})
	return _c
}

func (_c *MockDataNode_CheckChannelOperationProgress_Call) Return(_a0 *datapb.ChannelOperationProgressResponse, _a1 error) *MockDataNode_CheckChannelOperationProgress_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

// Compaction provides a mock function with given fields: ctx, req
func (_m *MockDataNode) Compaction(ctx context.Context, req *datapb.CompactionPlan) (*commonpb.Status, error) {
	ret := _m.Called(ctx, req)

	var r0 *commonpb.Status
	if rf, ok := ret.Get(0).(func(context.Context, *datapb.CompactionPlan) *commonpb.Status); ok {
		r0 = rf(ctx, req)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*commonpb.Status)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *datapb.CompactionPlan) error); ok {
		r1 = rf(ctx, req)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockDataNode_Compaction_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Compaction'
type MockDataNode_Compaction_Call struct {
	*mock.Call
}

// Compaction is a helper method to define mock.On call
//  - ctx context.Context
//  - req *datapb.CompactionPlan
func (_e *MockDataNode_Expecter) Compaction(ctx interface{}, req interface{}) *MockDataNode_Compaction_Call {
	return &MockDataNode_Compaction_Call{Call: _e.mock.On("Compaction", ctx, req)}
}

func (_c *MockDataNode_Compaction_Call) Run(run func(ctx context.Context, req *datapb.CompactionPlan)) *MockDataNode_Compaction_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*datapb.CompactionPlan))
	})
	return _c
}

func (_c *MockDataNode_Compaction_Call) Return(_a0 *commonpb.Status, _a1 error) *MockDataNode_Compaction_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

// FlushSegments provides a mock function with given fields: ctx, req
func (_m *MockDataNode) FlushSegments(ctx context.Context, req *datapb.FlushSegmentsRequest) (*commonpb.Status, error) {
	ret := _m.Called(ctx, req)

	var r0 *commonpb.Status
	if rf, ok := ret.Get(0).(func(context.Context, *datapb.FlushSegmentsRequest) *commonpb.Status); ok {
		r0 = rf(ctx, req)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*commonpb.Status)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *datapb.FlushSegmentsRequest) error); ok {
		r1 = rf(ctx, req)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockDataNode_FlushSegments_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'FlushSegments'
type MockDataNode_FlushSegments_Call struct {
	*mock.Call
}

// FlushSegments is a helper method to define mock.On call
//  - ctx context.Context
//  - req *datapb.FlushSegmentsRequest
func (_e *MockDataNode_Expecter) FlushSegments(ctx interface{}, req interface{}) *MockDataNode_FlushSegments_Call {
	return &MockDataNode_FlushSegments_Call{Call: _e.mock.On("FlushSegments", ctx, req)}
}

func (_c *MockDataNode_FlushSegments_Call) Run(run func(ctx context.Context, req *datapb.FlushSegmentsRequest)) *MockDataNode_FlushSegments_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*datapb.FlushSegmentsRequest))
	})
	return _c
}

func (_c *MockDataNode_FlushSegments_Call) Return(_a0 *commonpb.Status, _a1 error) *MockDataNode_FlushSegments_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

// GetAddress provides a mock function with given fields:
func (_m *MockDataNode) GetAddress() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// MockDataNode_GetAddress_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetAddress'
type MockDataNode_GetAddress_Call struct {
	*mock.Call
}

// GetAddress is a helper method to define mock.On call
func (_e *MockDataNode_Expecter) GetAddress() *MockDataNode_GetAddress_Call {
	return &MockDataNode_GetAddress_Call{Call: _e.mock.On("GetAddress")}
}

func (_c *MockDataNode_GetAddress_Call) Run(run func()) *MockDataNode_GetAddress_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockDataNode_GetAddress_Call) Return(_a0 string) *MockDataNode_GetAddress_Call {
	_c.Call.Return(_a0)
	return _c
}

// GetCompactionState provides a mock function with given fields: ctx, req
func (_m *MockDataNode) GetCompactionState(ctx context.Context, req *datapb.CompactionStateRequest) (*datapb.CompactionStateResponse, error) {
	ret := _m.Called(ctx, req)

	var r0 *datapb.CompactionStateResponse
	if rf, ok := ret.Get(0).(func(context.Context, *datapb.CompactionStateRequest) *datapb.CompactionStateResponse); ok {
		r0 = rf(ctx, req)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*datapb.CompactionStateResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *datapb.CompactionStateRequest) error); ok {
		r1 = rf(ctx, req)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockDataNode_GetCompactionState_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetCompactionState'
type MockDataNode_GetCompactionState_Call struct {
	*mock.Call
}

// GetCompactionState is a helper method to define mock.On call
//  - ctx context.Context
//  - req *datapb.CompactionStateRequest
func (_e *MockDataNode_Expecter) GetCompactionState(ctx interface{}, req interface{}) *MockDataNode_GetCompactionState_Call {
	return &MockDataNode_GetCompactionState_Call{Call: _e.mock.On("GetCompactionState", ctx, req)}
}

func (_c *MockDataNode_GetCompactionState_Call) Run(run func(ctx context.Context, req *datapb.CompactionStateRequest)) *MockDataNode_GetCompactionState_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*datapb.CompactionStateRequest))
	})
	return _c
}

func (_c *MockDataNode_GetCompactionState_Call) Return(_a0 *datapb.CompactionStateResponse, _a1 error) *MockDataNode_GetCompactionState_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

// GetComponentStates provides a mock function with given fields: ctx
func (_m *MockDataNode) GetComponentStates(ctx context.Context) (*milvuspb.ComponentStates, error) {
	ret := _m.Called(ctx)

	var r0 *milvuspb.ComponentStates
	if rf, ok := ret.Get(0).(func(context.Context) *milvuspb.ComponentStates); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*milvuspb.ComponentStates)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockDataNode_GetComponentStates_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetComponentStates'
type MockDataNode_GetComponentStates_Call struct {
	*mock.Call
}

// GetComponentStates is a helper method to define mock.On call
//  - ctx context.Context
func (_e *MockDataNode_Expecter) GetComponentStates(ctx interface{}) *MockDataNode_GetComponentStates_Call {
	return &MockDataNode_GetComponentStates_Call{Call: _e.mock.On("GetComponentStates", ctx)}
}

func (_c *MockDataNode_GetComponentStates_Call) Run(run func(ctx context.Context)) *MockDataNode_GetComponentStates_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context))
	})
	return _c
}

func (_c *MockDataNode_GetComponentStates_Call) Return(_a0 *milvuspb.ComponentStates, _a1 error) *MockDataNode_GetComponentStates_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

// GetMetrics provides a mock function with given fields: ctx, req
func (_m *MockDataNode) GetMetrics(ctx context.Context, req *milvuspb.GetMetricsRequest) (*milvuspb.GetMetricsResponse, error) {
	ret := _m.Called(ctx, req)

	var r0 *milvuspb.GetMetricsResponse
	if rf, ok := ret.Get(0).(func(context.Context, *milvuspb.GetMetricsRequest) *milvuspb.GetMetricsResponse); ok {
		r0 = rf(ctx, req)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*milvuspb.GetMetricsResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *milvuspb.GetMetricsRequest) error); ok {
		r1 = rf(ctx, req)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockDataNode_GetMetrics_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetMetrics'
type MockDataNode_GetMetrics_Call struct {
	*mock.Call
}

// GetMetrics is a helper method to define mock.On call
//  - ctx context.Context
//  - req *milvuspb.GetMetricsRequest
func (_e *MockDataNode_Expecter) GetMetrics(ctx interface{}, req interface{}) *MockDataNode_GetMetrics_Call {
	return &MockDataNode_GetMetrics_Call{Call: _e.mock.On("GetMetrics", ctx, req)}
}

func (_c *MockDataNode_GetMetrics_Call) Run(run func(ctx context.Context, req *milvuspb.GetMetricsRequest)) *MockDataNode_GetMetrics_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*milvuspb.GetMetricsRequest))
	})
	return _c
}

func (_c *MockDataNode_GetMetrics_Call) Return(_a0 *milvuspb.GetMetricsResponse, _a1 error) *MockDataNode_GetMetrics_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

// GetStateCode provides a mock function with given fields:
func (_m *MockDataNode) GetStateCode() commonpb.StateCode {
	ret := _m.Called()

	var r0 commonpb.StateCode
	if rf, ok := ret.Get(0).(func() commonpb.StateCode); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(commonpb.StateCode)
	}

	return r0
}

// MockDataNode_GetStateCode_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetStateCode'
type MockDataNode_GetStateCode_Call struct {
	*mock.Call
}

// GetStateCode is a helper method to define mock.On call
func (_e *MockDataNode_Expecter) GetStateCode() *MockDataNode_GetStateCode_Call {
	return &MockDataNode_GetStateCode_Call{Call: _e.mock.On("GetStateCode")}
}

func (_c *MockDataNode_GetStateCode_Call) Run(run func()) *MockDataNode_GetStateCode_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockDataNode_GetStateCode_Call) Return(_a0 commonpb.StateCode) *MockDataNode_GetStateCode_Call {
	_c.Call.Return(_a0)
	return _c
}

// GetStatisticsChannel provides a mock function with given fields: ctx
func (_m *MockDataNode) GetStatisticsChannel(ctx context.Context) (*milvuspb.StringResponse, error) {
	ret := _m.Called(ctx)

	var r0 *milvuspb.StringResponse
	if rf, ok := ret.Get(0).(func(context.Context) *milvuspb.StringResponse); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*milvuspb.StringResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockDataNode_GetStatisticsChannel_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetStatisticsChannel'
type MockDataNode_GetStatisticsChannel_Call struct {
	*mock.Call
}

// GetStatisticsChannel is a helper method to define mock.On call
//  - ctx context.Context
func (_e *MockDataNode_Expecter) GetStatisticsChannel(ctx interface{}) *MockDataNode_GetStatisticsChannel_Call {
	return &MockDataNode_GetStatisticsChannel_Call{Call: _e.mock.On("GetStatisticsChannel", ctx)}
}

func (_c *MockDataNode_GetStatisticsChannel_Call) Run(run func(ctx context.Context)) *MockDataNode_GetStatisticsChannel_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context))
	})
	return _c
}

func (_c *MockDataNode_GetStatisticsChannel_Call) Return(_a0 *milvuspb.StringResponse, _a1 error) *MockDataNode_GetStatisticsChannel_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

// Import provides a mock function with given fields: ctx, req
func (_m *MockDataNode) Import(ctx context.Context, req *datapb.ImportTaskRequest) (*commonpb.Status, error) {
	ret := _m.Called(ctx, req)

	var r0 *commonpb.Status
	if rf, ok := ret.Get(0).(func(context.Context, *datapb.ImportTaskRequest) *commonpb.Status); ok {
		r0 = rf(ctx, req)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*commonpb.Status)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *datapb.ImportTaskRequest) error); ok {
		r1 = rf(ctx, req)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockDataNode_Import_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Import'
type MockDataNode_Import_Call struct {
	*mock.Call
}

// Import is a helper method to define mock.On call
//  - ctx context.Context
//  - req *datapb.ImportTaskRequest
func (_e *MockDataNode_Expecter) Import(ctx interface{}, req interface{}) *MockDataNode_Import_Call {
	return &MockDataNode_Import_Call{Call: _e.mock.On("Import", ctx, req)}
}

func (_c *MockDataNode_Import_Call) Run(run func(ctx context.Context, req *datapb.ImportTaskRequest)) *MockDataNode_Import_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*datapb.ImportTaskRequest))
	})
	return _c
}

func (_c *MockDataNode_Import_Call) Return(_a0 *commonpb.Status, _a1 error) *MockDataNode_Import_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

// Init provides a mock function with given fields:
func (_m *MockDataNode) Init() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockDataNode_Init_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Init'
type MockDataNode_Init_Call struct {
	*mock.Call
}

// Init is a helper method to define mock.On call
func (_e *MockDataNode_Expecter) Init() *MockDataNode_Init_Call {
	return &MockDataNode_Init_Call{Call: _e.mock.On("Init")}
}

func (_c *MockDataNode_Init_Call) Run(run func()) *MockDataNode_Init_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockDataNode_Init_Call) Return(_a0 error) *MockDataNode_Init_Call {
	_c.Call.Return(_a0)
	return _c
}

// NotifyChannelOperation provides a mock function with given fields: ctx, req
func (_m *MockDataNode) NotifyChannelOperation(ctx context.Context, req *datapb.ChannelOperations) (*commonpb.Status, error) {
	ret := _m.Called(ctx, req)

	var r0 *commonpb.Status
	if rf, ok := ret.Get(0).(func(context.Context, *datapb.ChannelOperations) *commonpb.Status); ok {
		r0 = rf(ctx, req)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*commonpb.Status)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *datapb.ChannelOperations) error); ok {
		r1 = rf(ctx, req)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockDataNode_NotifyChannelOperation_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'NotifyChannelOperation'
type MockDataNode_NotifyChannelOperation_Call struct {
	*mock.Call
}

// NotifyChannelOperation is a helper method to define mock.On call
//  - ctx context.Context
//  - req *datapb.ChannelOperations
func (_e *MockDataNode_Expecter) NotifyChannelOperation(ctx interface{}, req interface{}) *MockDataNode_NotifyChannelOperation_Call {
	return &MockDataNode_NotifyChannelOperation_Call{Call: _e.mock.On("NotifyChannelOperation", ctx, req)}
}

func (_c *MockDataNode_NotifyChannelOperation_Call) Run(run func(ctx context.Context, req *datapb.ChannelOperations)) *MockDataNode_NotifyChannelOperation_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*datapb.ChannelOperations))
	})
	return _c
}

func (_c *MockDataNode_NotifyChannelOperation_Call) Return(_a0 *commonpb.Status, _a1 error) *MockDataNode_NotifyChannelOperation_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

// Register provides a mock function with given fields:
func (_m *MockDataNode) Register() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockDataNode_Register_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Register'
type MockDataNode_Register_Call struct {
	*mock.Call
}

// Register is a helper method to define mock.On call
func (_e *MockDataNode_Expecter) Register() *MockDataNode_Register_Call {
	return &MockDataNode_Register_Call{Call: _e.mock.On("Register")}
}

func (_c *MockDataNode_Register_Call) Run(run func()) *MockDataNode_Register_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockDataNode_Register_Call) Return(_a0 error) *MockDataNode_Register_Call {
	_c.Call.Return(_a0)
	return _c
}

// ResendSegmentStats provides a mock function with given fields: ctx, req
func (_m *MockDataNode) ResendSegmentStats(ctx context.Context, req *datapb.ResendSegmentStatsRequest) (*datapb.ResendSegmentStatsResponse, error) {
	ret := _m.Called(ctx, req)

	var r0 *datapb.ResendSegmentStatsResponse
	if rf, ok := ret.Get(0).(func(context.Context, *datapb.ResendSegmentStatsRequest) *datapb.ResendSegmentStatsResponse); ok {
		r0 = rf(ctx, req)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*datapb.ResendSegmentStatsResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *datapb.ResendSegmentStatsRequest) error); ok {
		r1 = rf(ctx, req)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockDataNode_ResendSegmentStats_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ResendSegmentStats'
type MockDataNode_ResendSegmentStats_Call struct {
	*mock.Call
}

// ResendSegmentStats is a helper method to define mock.On call
//  - ctx context.Context
//  - req *datapb.ResendSegmentStatsRequest
func (_e *MockDataNode_Expecter) ResendSegmentStats(ctx interface{}, req interface{}) *MockDataNode_ResendSegmentStats_Call {
	return &MockDataNode_ResendSegmentStats_Call{Call: _e.mock.On("ResendSegmentStats", ctx, req)}
}

func (_c *MockDataNode_ResendSegmentStats_Call) Run(run func(ctx context.Context, req *datapb.ResendSegmentStatsRequest)) *MockDataNode_ResendSegmentStats_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*datapb.ResendSegmentStatsRequest))
	})
	return _c
}

func (_c *MockDataNode_ResendSegmentStats_Call) Return(_a0 *datapb.ResendSegmentStatsResponse, _a1 error) *MockDataNode_ResendSegmentStats_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

// SetAddress provides a mock function with given fields: address
func (_m *MockDataNode) SetAddress(address string) {
	_m.Called(address)
}

// MockDataNode_SetAddress_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'SetAddress'
type MockDataNode_SetAddress_Call struct {
	*mock.Call
}

// SetAddress is a helper method to define mock.On call
//  - address string
func (_e *MockDataNode_Expecter) SetAddress(address interface{}) *MockDataNode_SetAddress_Call {
	return &MockDataNode_SetAddress_Call{Call: _e.mock.On("SetAddress", address)}
}

func (_c *MockDataNode_SetAddress_Call) Run(run func(address string)) *MockDataNode_SetAddress_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *MockDataNode_SetAddress_Call) Return() *MockDataNode_SetAddress_Call {
	_c.Call.Return()
	return _c
}

// SetDataCoord provides a mock function with given fields: dataCoord
func (_m *MockDataNode) SetDataCoord(dataCoord types.DataCoord) error {
	ret := _m.Called(dataCoord)

	var r0 error
	if rf, ok := ret.Get(0).(func(types.DataCoord) error); ok {
		r0 = rf(dataCoord)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockDataNode_SetDataCoord_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'SetDataCoord'
type MockDataNode_SetDataCoord_Call struct {
	*mock.Call
}

// SetDataCoord is a helper method to define mock.On call
//  - dataCoord types.DataCoord
func (_e *MockDataNode_Expecter) SetDataCoord(dataCoord interface{}) *MockDataNode_SetDataCoord_Call {
	return &MockDataNode_SetDataCoord_Call{Call: _e.mock.On("SetDataCoord", dataCoord)}
}

func (_c *MockDataNode_SetDataCoord_Call) Run(run func(dataCoord types.DataCoord)) *MockDataNode_SetDataCoord_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(types.DataCoord))
	})
	return _c
}

func (_c *MockDataNode_SetDataCoord_Call) Return(_a0 error) *MockDataNode_SetDataCoord_Call {
	_c.Call.Return(_a0)
	return _c
}

// SetEtcdClient provides a mock function with given fields: etcdClient
func (_m *MockDataNode) SetEtcdClient(etcdClient *clientv3.Client) {
	_m.Called(etcdClient)
}

// MockDataNode_SetEtcdClient_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'SetEtcdClient'
type MockDataNode_SetEtcdClient_Call struct {
	*mock.Call
}

// SetEtcdClient is a helper method to define mock.On call
//  - etcdClient *clientv3.Client
func (_e *MockDataNode_Expecter) SetEtcdClient(etcdClient interface{}) *MockDataNode_SetEtcdClient_Call {
	return &MockDataNode_SetEtcdClient_Call{Call: _e.mock.On("SetEtcdClient", etcdClient)}
}

func (_c *MockDataNode_SetEtcdClient_Call) Run(run func(etcdClient *clientv3.Client)) *MockDataNode_SetEtcdClient_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(*clientv3.Client))
	})
	return _c
}

func (_c *MockDataNode_SetEtcdClient_Call) Return() *MockDataNode_SetEtcdClient_Call {
	_c.Call.Return()
	return _c
}

// SetRootCoord provides a mock function with given fields: rootCoord
func (_m *MockDataNode) SetRootCoord(rootCoord types.RootCoord) error {
	ret := _m.Called(rootCoord)

	var r0 error
	if rf, ok := ret.Get(0).(func(types.RootCoord) error); ok {
		r0 = rf(rootCoord)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockDataNode_SetRootCoord_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'SetRootCoord'
type MockDataNode_SetRootCoord_Call struct {
	*mock.Call
}

// SetRootCoord is a helper method to define mock.On call
//  - rootCoord types.RootCoord
func (_e *MockDataNode_Expecter) SetRootCoord(rootCoord interface{}) *MockDataNode_SetRootCoord_Call {
	return &MockDataNode_SetRootCoord_Call{Call: _e.mock.On("SetRootCoord", rootCoord)}
}

func (_c *MockDataNode_SetRootCoord_Call) Run(run func(rootCoord types.RootCoord)) *MockDataNode_SetRootCoord_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(types.RootCoord))
	})
	return _c
}

func (_c *MockDataNode_SetRootCoord_Call) Return(_a0 error) *MockDataNode_SetRootCoord_Call {
	_c.Call.Return(_a0)
	return _c
}

// ShowConfigurations provides a mock function with given fields: ctx, req
func (_m *MockDataNode) ShowConfigurations(ctx context.Context, req *internalpb.ShowConfigurationsRequest) (*internalpb.ShowConfigurationsResponse, error) {
	ret := _m.Called(ctx, req)

	var r0 *internalpb.ShowConfigurationsResponse
	if rf, ok := ret.Get(0).(func(context.Context, *internalpb.ShowConfigurationsRequest) *internalpb.ShowConfigurationsResponse); ok {
		r0 = rf(ctx, req)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*internalpb.ShowConfigurationsResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *internalpb.ShowConfigurationsRequest) error); ok {
		r1 = rf(ctx, req)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockDataNode_ShowConfigurations_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ShowConfigurations'
type MockDataNode_ShowConfigurations_Call struct {
	*mock.Call
}

// ShowConfigurations is a helper method to define mock.On call
//  - ctx context.Context
//  - req *internalpb.ShowConfigurationsRequest
func (_e *MockDataNode_Expecter) ShowConfigurations(ctx interface{}, req interface{}) *MockDataNode_ShowConfigurations_Call {
	return &MockDataNode_ShowConfigurations_Call{Call: _e.mock.On("ShowConfigurations", ctx, req)}
}

func (_c *MockDataNode_ShowConfigurations_Call) Run(run func(ctx context.Context, req *internalpb.ShowConfigurationsRequest)) *MockDataNode_ShowConfigurations_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*internalpb.ShowConfigurationsRequest))
	})
	return _c
}

func (_c *MockDataNode_ShowConfigurations_Call) Return(_a0 *internalpb.ShowConfigurationsResponse, _a1 error) *MockDataNode_ShowConfigurations_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

// Start provides a mock function with given fields:
func (_m *MockDataNode) Start() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockDataNode_Start_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Start'
type MockDataNode_Start_Call struct {
	*mock.Call
}

// Start is a helper method to define mock.On call
func (_e *MockDataNode_Expecter) Start() *MockDataNode_Start_Call {
	return &MockDataNode_Start_Call{Call: _e.mock.On("Start")}
}

func (_c *MockDataNode_Start_Call) Run(run func()) *MockDataNode_Start_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockDataNode_Start_Call) Return(_a0 error) *MockDataNode_Start_Call {
	_c.Call.Return(_a0)
	return _c
}

// Stop provides a mock function with given fields:
func (_m *MockDataNode) Stop() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockDataNode_Stop_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Stop'
type MockDataNode_Stop_Call struct {
	*mock.Call
}

// Stop is a helper method to define mock.On call
func (_e *MockDataNode_Expecter) Stop() *MockDataNode_Stop_Call {
	return &MockDataNode_Stop_Call{Call: _e.mock.On("Stop")}
}

func (_c *MockDataNode_Stop_Call) Run(run func()) *MockDataNode_Stop_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockDataNode_Stop_Call) Return(_a0 error) *MockDataNode_Stop_Call {
	_c.Call.Return(_a0)
	return _c
}

// SyncSegments provides a mock function with given fields: ctx, req
func (_m *MockDataNode) SyncSegments(ctx context.Context, req *datapb.SyncSegmentsRequest) (*commonpb.Status, error) {
	ret := _m.Called(ctx, req)

	var r0 *commonpb.Status
	if rf, ok := ret.Get(0).(func(context.Context, *datapb.SyncSegmentsRequest) *commonpb.Status); ok {
		r0 = rf(ctx, req)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*commonpb.Status)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *datapb.SyncSegmentsRequest) error); ok {
		r1 = rf(ctx, req)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockDataNode_SyncSegments_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'SyncSegments'
type MockDataNode_SyncSegments_Call struct {
	*mock.Call
}

// SyncSegments is a helper method to define mock.On call
//  - ctx context.Context
//  - req *datapb.SyncSegmentsRequest
func (_e *MockDataNode_Expecter) SyncSegments(ctx interface{}, req interface{}) *MockDataNode_SyncSegments_Call {
	return &MockDataNode_SyncSegments_Call{Call: _e.mock.On("SyncSegments", ctx, req)}
}

func (_c *MockDataNode_SyncSegments_Call) Run(run func(ctx context.Context, req *datapb.SyncSegmentsRequest)) *MockDataNode_SyncSegments_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*datapb.SyncSegmentsRequest))
	})
	return _c
}

func (_c *MockDataNode_SyncSegments_Call) Return(_a0 *commonpb.Status, _a1 error) *MockDataNode_SyncSegments_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

// UpdateStateCode provides a mock function with given fields: stateCode
func (_m *MockDataNode) UpdateStateCode(stateCode commonpb.StateCode) {
	_m.Called(stateCode)
}

// MockDataNode_UpdateStateCode_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'UpdateStateCode'
type MockDataNode_UpdateStateCode_Call struct {
	*mock.Call
}

// UpdateStateCode is a helper method to define mock.On call
//  - stateCode commonpb.StateCode
func (_e *MockDataNode_Expecter) UpdateStateCode(stateCode interface{}) *MockDataNode_UpdateStateCode_Call {
	return &MockDataNode_UpdateStateCode_Call{Call: _e.mock.On("UpdateStateCode", stateCode)}
}

func (_c *MockDataNode_UpdateStateCode_Call) Run(run func(stateCode commonpb.StateCode)) *MockDataNode_UpdateStateCode_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(commonpb.StateCode))
	})
	return _c
}

func (_c *MockDataNode_UpdateStateCode_Call) Return() *MockDataNode_UpdateStateCode_Call {
	_c.Call.Return()
	return _c
}

// WatchDmChannels provides a mock function with given fields: ctx, req
func (_m *MockDataNode) WatchDmChannels(ctx context.Context, req *datapb.WatchDmChannelsRequest) (*commonpb.Status, error) {
	ret := _m.Called(ctx, req)

	var r0 *commonpb.Status
	if rf, ok := ret.Get(0).(func(context.Context, *datapb.WatchDmChannelsRequest) *commonpb.Status); ok {
		r0 = rf(ctx, req)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*commonpb.Status)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *datapb.WatchDmChannelsRequest) error); ok {
		r1 = rf(ctx, req)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockDataNode_WatchDmChannels_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'WatchDmChannels'
type MockDataNode_WatchDmChannels_Call struct {
	*mock.Call
}

// WatchDmChannels is a helper method to define mock.On call
//  - ctx context.Context
//  - req *datapb.WatchDmChannelsRequest
func (_e *MockDataNode_Expecter) WatchDmChannels(ctx interface{}, req interface{}) *MockDataNode_WatchDmChannels_Call {
	return &MockDataNode_WatchDmChannels_Call{Call: _e.mock.On("WatchDmChannels", ctx, req)}
}

func (_c *MockDataNode_WatchDmChannels_Call) Run(run func(ctx context.Context, req *datapb.WatchDmChannelsRequest)) *MockDataNode_WatchDmChannels_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*datapb.WatchDmChannelsRequest))
	})
	return _c
}

func (_c *MockDataNode_WatchDmChannels_Call) Return(_a0 *commonpb.Status, _a1 error) *MockDataNode_WatchDmChannels_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

type mockConstructorTestingTNewMockDataNode interface {
	mock.TestingT
	Cleanup(func())
}

// NewMockDataNode creates a new instance of MockDataNode. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewMockDataNode(t mockConstructorTestingTNewMockDataNode) *MockDataNode {
	mock := &MockDataNode{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
