// Code generated by mockery. DO NOT EDIT.

package mock

import (
	mock "github.com/stretchr/testify/mock"

	types100 "github.com/containernetworking/cni/pkg/types/100"
)

// MockPtpnatHandler is an autogenerated mock type for the Handler type
type MockPtpnatHandler struct {
	mock.Mock
}

type MockPtpnatHandler_Expecter struct {
	mock *mock.Mock
}

func (_m *MockPtpnatHandler) EXPECT() *MockPtpnatHandler_Expecter {
	return &MockPtpnatHandler_Expecter{mock: &_m.Mock}
}

// AddDefaultRoute provides a mock function with given fields: nsPath
func (_m *MockPtpnatHandler) AddDefaultRoute(nsPath string) error {
	ret := _m.Called(nsPath)

	if len(ret) == 0 {
		panic("no return value specified for AddDefaultRoute")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(nsPath)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockPtpnatHandler_AddDefaultRoute_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'AddDefaultRoute'
type MockPtpnatHandler_AddDefaultRoute_Call struct {
	*mock.Call
}

// AddDefaultRoute is a helper method to define mock.On call
//   - nsPath string
func (_e *MockPtpnatHandler_Expecter) AddDefaultRoute(nsPath interface{}) *MockPtpnatHandler_AddDefaultRoute_Call {
	return &MockPtpnatHandler_AddDefaultRoute_Call{Call: _e.mock.On("AddDefaultRoute", nsPath)}
}

func (_c *MockPtpnatHandler_AddDefaultRoute_Call) Run(run func(nsPath string)) *MockPtpnatHandler_AddDefaultRoute_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *MockPtpnatHandler_AddDefaultRoute_Call) Return(_a0 error) *MockPtpnatHandler_AddDefaultRoute_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockPtpnatHandler_AddDefaultRoute_Call) RunAndReturn(run func(string) error) *MockPtpnatHandler_AddDefaultRoute_Call {
	_c.Call.Return(run)
	return _c
}

// AllocIPs provides a mock function with given fields: plugin, stdinData
func (_m *MockPtpnatHandler) AllocIPs(plugin string, stdinData []byte) ([]*types100.IPConfig, error) {
	ret := _m.Called(plugin, stdinData)

	if len(ret) == 0 {
		panic("no return value specified for AllocIPs")
	}

	var r0 []*types100.IPConfig
	var r1 error
	if rf, ok := ret.Get(0).(func(string, []byte) ([]*types100.IPConfig, error)); ok {
		return rf(plugin, stdinData)
	}
	if rf, ok := ret.Get(0).(func(string, []byte) []*types100.IPConfig); ok {
		r0 = rf(plugin, stdinData)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*types100.IPConfig)
		}
	}

	if rf, ok := ret.Get(1).(func(string, []byte) error); ok {
		r1 = rf(plugin, stdinData)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockPtpnatHandler_AllocIPs_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'AllocIPs'
type MockPtpnatHandler_AllocIPs_Call struct {
	*mock.Call
}

// AllocIPs is a helper method to define mock.On call
//   - plugin string
//   - stdinData []byte
func (_e *MockPtpnatHandler_Expecter) AllocIPs(plugin interface{}, stdinData interface{}) *MockPtpnatHandler_AllocIPs_Call {
	return &MockPtpnatHandler_AllocIPs_Call{Call: _e.mock.On("AllocIPs", plugin, stdinData)}
}

func (_c *MockPtpnatHandler_AllocIPs_Call) Run(run func(plugin string, stdinData []byte)) *MockPtpnatHandler_AllocIPs_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string), args[1].([]byte))
	})
	return _c
}

func (_c *MockPtpnatHandler_AllocIPs_Call) Return(_a0 []*types100.IPConfig, _a1 error) *MockPtpnatHandler_AllocIPs_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockPtpnatHandler_AllocIPs_Call) RunAndReturn(run func(string, []byte) ([]*types100.IPConfig, error)) *MockPtpnatHandler_AllocIPs_Call {
	_c.Call.Return(run)
	return _c
}

// AttachDNATBPF provides a mock function with given fields: ifaceName
func (_m *MockPtpnatHandler) AttachDNATBPF(ifaceName string) error {
	ret := _m.Called(ifaceName)

	if len(ret) == 0 {
		panic("no return value specified for AttachDNATBPF")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(ifaceName)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockPtpnatHandler_AttachDNATBPF_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'AttachDNATBPF'
type MockPtpnatHandler_AttachDNATBPF_Call struct {
	*mock.Call
}

// AttachDNATBPF is a helper method to define mock.On call
//   - ifaceName string
func (_e *MockPtpnatHandler_Expecter) AttachDNATBPF(ifaceName interface{}) *MockPtpnatHandler_AttachDNATBPF_Call {
	return &MockPtpnatHandler_AttachDNATBPF_Call{Call: _e.mock.On("AttachDNATBPF", ifaceName)}
}

func (_c *MockPtpnatHandler_AttachDNATBPF_Call) Run(run func(ifaceName string)) *MockPtpnatHandler_AttachDNATBPF_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *MockPtpnatHandler_AttachDNATBPF_Call) Return(_a0 error) *MockPtpnatHandler_AttachDNATBPF_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockPtpnatHandler_AttachDNATBPF_Call) RunAndReturn(run func(string) error) *MockPtpnatHandler_AttachDNATBPF_Call {
	_c.Call.Return(run)
	return _c
}

// AttachSNATBPF provides a mock function with given fields: ifaceName
func (_m *MockPtpnatHandler) AttachSNATBPF(ifaceName string) error {
	ret := _m.Called(ifaceName)

	if len(ret) == 0 {
		panic("no return value specified for AttachSNATBPF")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(ifaceName)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockPtpnatHandler_AttachSNATBPF_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'AttachSNATBPF'
type MockPtpnatHandler_AttachSNATBPF_Call struct {
	*mock.Call
}

// AttachSNATBPF is a helper method to define mock.On call
//   - ifaceName string
func (_e *MockPtpnatHandler_Expecter) AttachSNATBPF(ifaceName interface{}) *MockPtpnatHandler_AttachSNATBPF_Call {
	return &MockPtpnatHandler_AttachSNATBPF_Call{Call: _e.mock.On("AttachSNATBPF", ifaceName)}
}

func (_c *MockPtpnatHandler_AttachSNATBPF_Call) Run(run func(ifaceName string)) *MockPtpnatHandler_AttachSNATBPF_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *MockPtpnatHandler_AttachSNATBPF_Call) Return(_a0 error) *MockPtpnatHandler_AttachSNATBPF_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockPtpnatHandler_AttachSNATBPF_Call) RunAndReturn(run func(string) error) *MockPtpnatHandler_AttachSNATBPF_Call {
	_c.Call.Return(run)
	return _c
}

// ConfigureSNAT provides a mock function with given fields: ifaceName
func (_m *MockPtpnatHandler) ConfigureSNAT(ifaceName string) error {
	ret := _m.Called(ifaceName)

	if len(ret) == 0 {
		panic("no return value specified for ConfigureSNAT")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(ifaceName)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockPtpnatHandler_ConfigureSNAT_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ConfigureSNAT'
type MockPtpnatHandler_ConfigureSNAT_Call struct {
	*mock.Call
}

// ConfigureSNAT is a helper method to define mock.On call
//   - ifaceName string
func (_e *MockPtpnatHandler_Expecter) ConfigureSNAT(ifaceName interface{}) *MockPtpnatHandler_ConfigureSNAT_Call {
	return &MockPtpnatHandler_ConfigureSNAT_Call{Call: _e.mock.On("ConfigureSNAT", ifaceName)}
}

func (_c *MockPtpnatHandler_ConfigureSNAT_Call) Run(run func(ifaceName string)) *MockPtpnatHandler_ConfigureSNAT_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *MockPtpnatHandler_ConfigureSNAT_Call) Return(_a0 error) *MockPtpnatHandler_ConfigureSNAT_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockPtpnatHandler_ConfigureSNAT_Call) RunAndReturn(run func(string) error) *MockPtpnatHandler_ConfigureSNAT_Call {
	_c.Call.Return(run)
	return _c
}

// CreateAndConfigureVethPair provides a mock function with given fields: netNS, ips
func (_m *MockPtpnatHandler) CreateAndConfigureVethPair(netNS string, ips []*types100.IPConfig) (string, string, error) {
	ret := _m.Called(netNS, ips)

	if len(ret) == 0 {
		panic("no return value specified for CreateAndConfigureVethPair")
	}

	var r0 string
	var r1 string
	var r2 error
	if rf, ok := ret.Get(0).(func(string, []*types100.IPConfig) (string, string, error)); ok {
		return rf(netNS, ips)
	}
	if rf, ok := ret.Get(0).(func(string, []*types100.IPConfig) string); ok {
		r0 = rf(netNS, ips)
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func(string, []*types100.IPConfig) string); ok {
		r1 = rf(netNS, ips)
	} else {
		r1 = ret.Get(1).(string)
	}

	if rf, ok := ret.Get(2).(func(string, []*types100.IPConfig) error); ok {
		r2 = rf(netNS, ips)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// MockPtpnatHandler_CreateAndConfigureVethPair_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'CreateAndConfigureVethPair'
type MockPtpnatHandler_CreateAndConfigureVethPair_Call struct {
	*mock.Call
}

// CreateAndConfigureVethPair is a helper method to define mock.On call
//   - netNS string
//   - ips []*types100.IPConfig
func (_e *MockPtpnatHandler_Expecter) CreateAndConfigureVethPair(netNS interface{}, ips interface{}) *MockPtpnatHandler_CreateAndConfigureVethPair_Call {
	return &MockPtpnatHandler_CreateAndConfigureVethPair_Call{Call: _e.mock.On("CreateAndConfigureVethPair", netNS, ips)}
}

func (_c *MockPtpnatHandler_CreateAndConfigureVethPair_Call) Run(run func(netNS string, ips []*types100.IPConfig)) *MockPtpnatHandler_CreateAndConfigureVethPair_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string), args[1].([]*types100.IPConfig))
	})
	return _c
}

func (_c *MockPtpnatHandler_CreateAndConfigureVethPair_Call) Return(_a0 string, _a1 string, _a2 error) *MockPtpnatHandler_CreateAndConfigureVethPair_Call {
	_c.Call.Return(_a0, _a1, _a2)
	return _c
}

func (_c *MockPtpnatHandler_CreateAndConfigureVethPair_Call) RunAndReturn(run func(string, []*types100.IPConfig) (string, string, error)) *MockPtpnatHandler_CreateAndConfigureVethPair_Call {
	_c.Call.Return(run)
	return _c
}

// DeallocIPs provides a mock function with given fields: plugin, stdinData
func (_m *MockPtpnatHandler) DeallocIPs(plugin string, stdinData []byte) error {
	ret := _m.Called(plugin, stdinData)

	if len(ret) == 0 {
		panic("no return value specified for DeallocIPs")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(string, []byte) error); ok {
		r0 = rf(plugin, stdinData)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockPtpnatHandler_DeallocIPs_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'DeallocIPs'
type MockPtpnatHandler_DeallocIPs_Call struct {
	*mock.Call
}

// DeallocIPs is a helper method to define mock.On call
//   - plugin string
//   - stdinData []byte
func (_e *MockPtpnatHandler_Expecter) DeallocIPs(plugin interface{}, stdinData interface{}) *MockPtpnatHandler_DeallocIPs_Call {
	return &MockPtpnatHandler_DeallocIPs_Call{Call: _e.mock.On("DeallocIPs", plugin, stdinData)}
}

func (_c *MockPtpnatHandler_DeallocIPs_Call) Run(run func(plugin string, stdinData []byte)) *MockPtpnatHandler_DeallocIPs_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string), args[1].([]byte))
	})
	return _c
}

func (_c *MockPtpnatHandler_DeallocIPs_Call) Return(_a0 error) *MockPtpnatHandler_DeallocIPs_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockPtpnatHandler_DeallocIPs_Call) RunAndReturn(run func(string, []byte) error) *MockPtpnatHandler_DeallocIPs_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockPtpnatHandler creates a new instance of MockPtpnatHandler. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockPtpnatHandler(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockPtpnatHandler {
	mock := &MockPtpnatHandler{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
