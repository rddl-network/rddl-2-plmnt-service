// Code generated by MockGen. DO NOT EDIT.
// Source: service/elements_client.go

// Package testutil is a generated GoMock package.
package testutil

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	types "github.com/rddl-network/elements-rpc/types"
)

// MockIElementsClient is a mock of IElementsClient interface.
type MockIElementsClient struct {
	ctrl     *gomock.Controller
	recorder *MockIElementsClientMockRecorder
}

// MockIElementsClientMockRecorder is the mock recorder for MockIElementsClient.
type MockIElementsClientMockRecorder struct {
	mock *MockIElementsClient
}

// NewMockIElementsClient creates a new mock instance.
func NewMockIElementsClient(ctrl *gomock.Controller) *MockIElementsClient {
	mock := &MockIElementsClient{ctrl: ctrl}
	mock.recorder = &MockIElementsClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockIElementsClient) EXPECT() *MockIElementsClientMockRecorder {
	return m.recorder
}

// GetNewAddress mocks base method.
func (m *MockIElementsClient) GetNewAddress(url string, params []string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetNewAddress", url, params)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetNewAddress indicates an expected call of GetNewAddress.
func (mr *MockIElementsClientMockRecorder) GetNewAddress(url, params interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetNewAddress", reflect.TypeOf((*MockIElementsClient)(nil).GetNewAddress), url, params)
}

// ListReceivedByAddress mocks base method.
func (m *MockIElementsClient) ListReceivedByAddress(url string, params []string) ([]types.ListReceivedByAddressResult, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListReceivedByAddress", url, params)
	ret0, _ := ret[0].([]types.ListReceivedByAddressResult)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListReceivedByAddress indicates an expected call of ListReceivedByAddress.
func (mr *MockIElementsClientMockRecorder) ListReceivedByAddress(url, params interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListReceivedByAddress", reflect.TypeOf((*MockIElementsClient)(nil).ListReceivedByAddress), url, params)
}
