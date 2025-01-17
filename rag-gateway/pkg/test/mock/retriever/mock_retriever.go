// Code generated by MockGen. DO NOT EDIT.
// Source: ./pkg/retriever/retriever.go
//
// Generated by this command:
//
//	mockgen -source=./pkg/retriever/retriever.go -destination=pkg/test/mock/retriever/mock_retriever.go Retriever
//

// Package mock_retriever is a generated GoMock package.
package mock_retriever

import (
	reflect "reflect"

	schema "github.com/tmc/langchaingo/schema"
	gomock "go.uber.org/mock/gomock"
)

// MockRetriever is a mock of Retriever interface.
type MockRetriever struct {
	ctrl     *gomock.Controller
	recorder *MockRetrieverMockRecorder
}

// MockRetrieverMockRecorder is the mock recorder for MockRetriever.
type MockRetrieverMockRecorder struct {
	mock *MockRetriever
}

// NewMockRetriever creates a new mock instance.
func NewMockRetriever(ctrl *gomock.Controller) *MockRetriever {
	mock := &MockRetriever{ctrl: ctrl}
	mock.recorder = &MockRetrieverMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockRetriever) EXPECT() *MockRetrieverMockRecorder {
	return m.recorder
}

// AsRetriever mocks base method.
func (m *MockRetriever) AsRetriever() schema.Retriever {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AsRetriever")
	ret0, _ := ret[0].(schema.Retriever)
	return ret0
}

// AsRetriever indicates an expected call of AsRetriever.
func (mr *MockRetrieverMockRecorder) AsRetriever() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AsRetriever", reflect.TypeOf((*MockRetriever)(nil).AsRetriever))
}

// Name mocks base method.
func (m *MockRetriever) Name() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Name")
	ret0, _ := ret[0].(string)
	return ret0
}

// Name indicates an expected call of Name.
func (mr *MockRetrieverMockRecorder) Name() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Name", reflect.TypeOf((*MockRetriever)(nil).Name))
}

// Priority mocks base method.
func (m *MockRetriever) Priority() int {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Priority")
	ret0, _ := ret[0].(int)
	return ret0
}

// Priority indicates an expected call of Priority.
func (mr *MockRetrieverMockRecorder) Priority() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Priority", reflect.TypeOf((*MockRetriever)(nil).Priority))
}

// Weight mocks base method.
func (m *MockRetriever) Weight() float64 {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Weight")
	ret0, _ := ret[0].(float64)
	return ret0
}

// Weight indicates an expected call of Weight.
func (mr *MockRetrieverMockRecorder) Weight() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Weight", reflect.TypeOf((*MockRetriever)(nil).Weight))
}
