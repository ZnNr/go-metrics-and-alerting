// Code generated by mockery. DO NOT EDIT.

package server

import (
	net "net"

	mock "github.com/stretchr/testify/mock"
)

// mockGrpcServer is an autogenerated mock type for the grpcServer type
type mockGrpcServer struct {
	mock.Mock
}

// GracefulStop provides a mock function with given fields:
func (_m *mockGrpcServer) GracefulStop() {
	_m.Called()
}

// Serve provides a mock function with given fields: lis
func (_m *mockGrpcServer) Serve(lis net.Listener) error {
	ret := _m.Called(lis)

	if len(ret) == 0 {
		panic("no return value specified for Serve")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(net.Listener) error); ok {
		r0 = rf(lis)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// newMockGrpcServer creates a new instance of mockGrpcServer. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func newMockGrpcServer(t interface {
	mock.TestingT
	Cleanup(func())
}) *mockGrpcServer {
	mock := &mockGrpcServer{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
