// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/celestiaorg/celestia-node/nodebuilder/blob (interfaces: Module)

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	blob "github.com/celestiaorg/celestia-node/blob"
	namespace "github.com/celestiaorg/nmt/namespace"
	gomock "github.com/golang/mock/gomock"
)

// MockModule is a mock of Module interface.
type MockModule struct {
	ctrl     *gomock.Controller
	recorder *MockModuleMockRecorder
}

// MockModuleMockRecorder is the mock recorder for MockModule.
type MockModuleMockRecorder struct {
	mock *MockModule
}

// NewMockModule creates a new mock instance.
func NewMockModule(ctrl *gomock.Controller) *MockModule {
	mock := &MockModule{ctrl: ctrl}
	mock.recorder = &MockModuleMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockModule) EXPECT() *MockModuleMockRecorder {
	return m.recorder
}

// Get mocks base method.
func (m *MockModule) Get(arg0 context.Context, arg1 uint64, arg2 namespace.ID, arg3 blob.Commitment) (*blob.Blob, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Get", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].(*blob.Blob)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Get indicates an expected call of Get.
func (mr *MockModuleMockRecorder) Get(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockModule)(nil).Get), arg0, arg1, arg2, arg3)
}

// GetAll mocks base method.
func (m *MockModule) GetAll(arg0 context.Context, arg1 uint64, arg2 ...namespace.ID) ([]*blob.Blob, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "GetAll", varargs...)
	ret0, _ := ret[0].([]*blob.Blob)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAll indicates an expected call of GetAll.
func (mr *MockModuleMockRecorder) GetAll(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAll", reflect.TypeOf((*MockModule)(nil).GetAll), varargs...)
}

// GetProof mocks base method.
func (m *MockModule) GetProof(arg0 context.Context, arg1 uint64, arg2 namespace.ID, arg3 blob.Commitment) (*blob.Proof, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetProof", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].(*blob.Proof)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetProof indicates an expected call of GetProof.
func (mr *MockModuleMockRecorder) GetProof(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetProof", reflect.TypeOf((*MockModule)(nil).GetProof), arg0, arg1, arg2, arg3)
}

// Included mocks base method.
func (m *MockModule) Included(arg0 context.Context, arg1 uint64, arg2 namespace.ID, arg3 *blob.Proof, arg4 blob.Commitment) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Included", arg0, arg1, arg2, arg3, arg4)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Included indicates an expected call of Included.
func (mr *MockModuleMockRecorder) Included(arg0, arg1, arg2, arg3, arg4 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Included", reflect.TypeOf((*MockModule)(nil).Included), arg0, arg1, arg2, arg3, arg4)
}

// Submit mocks base method.
func (m *MockModule) Submit(arg0 context.Context, arg1 ...*blob.Blob) (uint64, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0}
	for _, a := range arg1 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Submit", varargs...)
	ret0, _ := ret[0].(uint64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Submit indicates an expected call of Submit.
func (mr *MockModuleMockRecorder) Submit(arg0 interface{}, arg1 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0}, arg1...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Submit", reflect.TypeOf((*MockModule)(nil).Submit), varargs...)
}
