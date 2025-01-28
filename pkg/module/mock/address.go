// Code generated by MockGen. DO NOT EDIT.
// Source: address.go
//
// Generated by this command:
//
//	mockgen -source=address.go -destination=mock/address.go -package=mock -typed
//

// Package mock is a generated GoMock package.
package mock

import (
	context "context"
	reflect "reflect"

	gomock "go.uber.org/mock/gomock"
)

// MockIdByHash is a mock of IdByHash interface.
type MockIdByHash struct {
	ctrl     *gomock.Controller
	recorder *MockIdByHashMockRecorder
}

// MockIdByHashMockRecorder is the mock recorder for MockIdByHash.
type MockIdByHashMockRecorder struct {
	mock *MockIdByHash
}

// NewMockIdByHash creates a new mock instance.
func NewMockIdByHash(ctrl *gomock.Controller) *MockIdByHash {
	mock := &MockIdByHash{ctrl: ctrl}
	mock.recorder = &MockIdByHashMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockIdByHash) EXPECT() *MockIdByHashMockRecorder {
	return m.recorder
}

// IdByHash mocks base method.
func (m *MockIdByHash) IdByHash(ctx context.Context, hash ...[]byte) ([]uint64, error) {
	m.ctrl.T.Helper()
	varargs := []any{ctx}
	for _, a := range hash {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "IdByHash", varargs...)
	ret0, _ := ret[0].([]uint64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// IdByHash indicates an expected call of IdByHash.
func (mr *MockIdByHashMockRecorder) IdByHash(ctx any, hash ...any) *MockIdByHashIdByHashCall {
	mr.mock.ctrl.T.Helper()
	varargs := append([]any{ctx}, hash...)
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IdByHash", reflect.TypeOf((*MockIdByHash)(nil).IdByHash), varargs...)
	return &MockIdByHashIdByHashCall{Call: call}
}

// MockIdByHashIdByHashCall wrap *gomock.Call
type MockIdByHashIdByHashCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockIdByHashIdByHashCall) Return(arg0 []uint64, arg1 error) *MockIdByHashIdByHashCall {
	c.Call = c.Call.Return(arg0, arg1)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockIdByHashIdByHashCall) Do(f func(context.Context, ...[]byte) ([]uint64, error)) *MockIdByHashIdByHashCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockIdByHashIdByHashCall) DoAndReturn(f func(context.Context, ...[]byte) ([]uint64, error)) *MockIdByHashIdByHashCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}
