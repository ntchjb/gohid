// Code generated by MockGen. DO NOT EDIT.
// Source: ./usb/usb.go
//
// Generated by this command:
//
//	mockgen -source=./usb/usb.go -destination=./usb/mock_usb.go -package=usb
//

// Package usb is a generated GoMock package.
package usb

import (
	context "context"
	reflect "reflect"

	gousb "github.com/google/gousb"
	gomock "go.uber.org/mock/gomock"
)

// MockStreamWriter is a mock of StreamWriter interface.
type MockStreamWriter struct {
	ctrl     *gomock.Controller
	recorder *MockStreamWriterMockRecorder
}

// MockStreamWriterMockRecorder is the mock recorder for MockStreamWriter.
type MockStreamWriterMockRecorder struct {
	mock *MockStreamWriter
}

// NewMockStreamWriter creates a new mock instance.
func NewMockStreamWriter(ctrl *gomock.Controller) *MockStreamWriter {
	mock := &MockStreamWriter{ctrl: ctrl}
	mock.recorder = &MockStreamWriterMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockStreamWriter) EXPECT() *MockStreamWriterMockRecorder {
	return m.recorder
}

// Close mocks base method.
func (m *MockStreamWriter) Close() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Close")
	ret0, _ := ret[0].(error)
	return ret0
}

// Close indicates an expected call of Close.
func (mr *MockStreamWriterMockRecorder) Close() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*MockStreamWriter)(nil).Close))
}

// WriteContext mocks base method.
func (m *MockStreamWriter) WriteContext(ctx context.Context, data []byte) (int, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "WriteContext", ctx, data)
	ret0, _ := ret[0].(int)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// WriteContext indicates an expected call of WriteContext.
func (mr *MockStreamWriterMockRecorder) WriteContext(ctx, data any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WriteContext", reflect.TypeOf((*MockStreamWriter)(nil).WriteContext), ctx, data)
}

// MockStreamReader is a mock of StreamReader interface.
type MockStreamReader struct {
	ctrl     *gomock.Controller
	recorder *MockStreamReaderMockRecorder
}

// MockStreamReaderMockRecorder is the mock recorder for MockStreamReader.
type MockStreamReaderMockRecorder struct {
	mock *MockStreamReader
}

// NewMockStreamReader creates a new mock instance.
func NewMockStreamReader(ctrl *gomock.Controller) *MockStreamReader {
	mock := &MockStreamReader{ctrl: ctrl}
	mock.recorder = &MockStreamReaderMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockStreamReader) EXPECT() *MockStreamReaderMockRecorder {
	return m.recorder
}

// Close mocks base method.
func (m *MockStreamReader) Close() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Close")
	ret0, _ := ret[0].(error)
	return ret0
}

// Close indicates an expected call of Close.
func (mr *MockStreamReaderMockRecorder) Close() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*MockStreamReader)(nil).Close))
}

// ReadContext mocks base method.
func (m *MockStreamReader) ReadContext(ctx context.Context, data []byte) (int, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ReadContext", ctx, data)
	ret0, _ := ret[0].(int)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ReadContext indicates an expected call of ReadContext.
func (mr *MockStreamReaderMockRecorder) ReadContext(ctx, data any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ReadContext", reflect.TypeOf((*MockStreamReader)(nil).ReadContext), ctx, data)
}

// MockOutEndpoint is a mock of OutEndpoint interface.
type MockOutEndpoint struct {
	ctrl     *gomock.Controller
	recorder *MockOutEndpointMockRecorder
}

// MockOutEndpointMockRecorder is the mock recorder for MockOutEndpoint.
type MockOutEndpointMockRecorder struct {
	mock *MockOutEndpoint
}

// NewMockOutEndpoint creates a new mock instance.
func NewMockOutEndpoint(ctrl *gomock.Controller) *MockOutEndpoint {
	mock := &MockOutEndpoint{ctrl: ctrl}
	mock.recorder = &MockOutEndpointMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockOutEndpoint) EXPECT() *MockOutEndpointMockRecorder {
	return m.recorder
}

// Descriptor mocks base method.
func (m *MockOutEndpoint) Descriptor() gousb.EndpointDesc {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Descriptor")
	ret0, _ := ret[0].(gousb.EndpointDesc)
	return ret0
}

// Descriptor indicates an expected call of Descriptor.
func (mr *MockOutEndpointMockRecorder) Descriptor() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Descriptor", reflect.TypeOf((*MockOutEndpoint)(nil).Descriptor))
}

// NewStream mocks base method.
func (m *MockOutEndpoint) NewStream(count int) (StreamWriter, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "NewStream", count)
	ret0, _ := ret[0].(StreamWriter)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// NewStream indicates an expected call of NewStream.
func (mr *MockOutEndpointMockRecorder) NewStream(count any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "NewStream", reflect.TypeOf((*MockOutEndpoint)(nil).NewStream), count)
}

// MockInEndpoint is a mock of InEndpoint interface.
type MockInEndpoint struct {
	ctrl     *gomock.Controller
	recorder *MockInEndpointMockRecorder
}

// MockInEndpointMockRecorder is the mock recorder for MockInEndpoint.
type MockInEndpointMockRecorder struct {
	mock *MockInEndpoint
}

// NewMockInEndpoint creates a new mock instance.
func NewMockInEndpoint(ctrl *gomock.Controller) *MockInEndpoint {
	mock := &MockInEndpoint{ctrl: ctrl}
	mock.recorder = &MockInEndpointMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockInEndpoint) EXPECT() *MockInEndpointMockRecorder {
	return m.recorder
}

// Descriptor mocks base method.
func (m *MockInEndpoint) Descriptor() gousb.EndpointDesc {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Descriptor")
	ret0, _ := ret[0].(gousb.EndpointDesc)
	return ret0
}

// Descriptor indicates an expected call of Descriptor.
func (mr *MockInEndpointMockRecorder) Descriptor() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Descriptor", reflect.TypeOf((*MockInEndpoint)(nil).Descriptor))
}

// NewStream mocks base method.
func (m *MockInEndpoint) NewStream(count int) (StreamReader, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "NewStream", count)
	ret0, _ := ret[0].(StreamReader)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// NewStream indicates an expected call of NewStream.
func (mr *MockInEndpointMockRecorder) NewStream(count any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "NewStream", reflect.TypeOf((*MockInEndpoint)(nil).NewStream), count)
}

// MockInterface is a mock of Interface interface.
type MockInterface struct {
	ctrl     *gomock.Controller
	recorder *MockInterfaceMockRecorder
}

// MockInterfaceMockRecorder is the mock recorder for MockInterface.
type MockInterfaceMockRecorder struct {
	mock *MockInterface
}

// NewMockInterface creates a new mock instance.
func NewMockInterface(ctrl *gomock.Controller) *MockInterface {
	mock := &MockInterface{ctrl: ctrl}
	mock.recorder = &MockInterfaceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockInterface) EXPECT() *MockInterfaceMockRecorder {
	return m.recorder
}

// Close mocks base method.
func (m *MockInterface) Close() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Close")
	ret0, _ := ret[0].(error)
	return ret0
}

// Close indicates an expected call of Close.
func (mr *MockInterfaceMockRecorder) Close() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*MockInterface)(nil).Close))
}

// InEndpoint mocks base method.
func (m *MockInterface) InEndpoint(num int) (InEndpoint, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "InEndpoint", num)
	ret0, _ := ret[0].(InEndpoint)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// InEndpoint indicates an expected call of InEndpoint.
func (mr *MockInterfaceMockRecorder) InEndpoint(num any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "InEndpoint", reflect.TypeOf((*MockInterface)(nil).InEndpoint), num)
}

// OutEndpoint mocks base method.
func (m *MockInterface) OutEndpoint(num int) (OutEndpoint, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "OutEndpoint", num)
	ret0, _ := ret[0].(OutEndpoint)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// OutEndpoint indicates an expected call of OutEndpoint.
func (mr *MockInterfaceMockRecorder) OutEndpoint(num any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "OutEndpoint", reflect.TypeOf((*MockInterface)(nil).OutEndpoint), num)
}

// MockConfig is a mock of Config interface.
type MockConfig struct {
	ctrl     *gomock.Controller
	recorder *MockConfigMockRecorder
}

// MockConfigMockRecorder is the mock recorder for MockConfig.
type MockConfigMockRecorder struct {
	mock *MockConfig
}

// NewMockConfig creates a new mock instance.
func NewMockConfig(ctrl *gomock.Controller) *MockConfig {
	mock := &MockConfig{ctrl: ctrl}
	mock.recorder = &MockConfigMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockConfig) EXPECT() *MockConfigMockRecorder {
	return m.recorder
}

// Close mocks base method.
func (m *MockConfig) Close() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Close")
	ret0, _ := ret[0].(error)
	return ret0
}

// Close indicates an expected call of Close.
func (mr *MockConfigMockRecorder) Close() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*MockConfig)(nil).Close))
}

// Interface mocks base method.
func (m *MockConfig) Interface(num, alt int) (Interface, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Interface", num, alt)
	ret0, _ := ret[0].(Interface)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Interface indicates an expected call of Interface.
func (mr *MockConfigMockRecorder) Interface(num, alt any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Interface", reflect.TypeOf((*MockConfig)(nil).Interface), num, alt)
}

// MockDevice is a mock of Device interface.
type MockDevice struct {
	ctrl     *gomock.Controller
	recorder *MockDeviceMockRecorder
}

// MockDeviceMockRecorder is the mock recorder for MockDevice.
type MockDeviceMockRecorder struct {
	mock *MockDevice
}

// NewMockDevice creates a new mock instance.
func NewMockDevice(ctrl *gomock.Controller) *MockDevice {
	mock := &MockDevice{ctrl: ctrl}
	mock.recorder = &MockDeviceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockDevice) EXPECT() *MockDeviceMockRecorder {
	return m.recorder
}

// Close mocks base method.
func (m *MockDevice) Close() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Close")
	ret0, _ := ret[0].(error)
	return ret0
}

// Close indicates an expected call of Close.
func (mr *MockDeviceMockRecorder) Close() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*MockDevice)(nil).Close))
}

// Config mocks base method.
func (m *MockDevice) Config(configNumber int) (Config, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Config", configNumber)
	ret0, _ := ret[0].(Config)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Config indicates an expected call of Config.
func (mr *MockDeviceMockRecorder) Config(configNumber any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Config", reflect.TypeOf((*MockDevice)(nil).Config), configNumber)
}

// Control mocks base method.
func (m *MockDevice) Control(bmRequestType, bRequest uint8, wValue, wIndex uint16, data []byte) (int, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Control", bmRequestType, bRequest, wValue, wIndex, data)
	ret0, _ := ret[0].(int)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Control indicates an expected call of Control.
func (mr *MockDeviceMockRecorder) Control(bmRequestType, bRequest, wValue, wIndex, data any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Control", reflect.TypeOf((*MockDevice)(nil).Control), bmRequestType, bRequest, wValue, wIndex, data)
}

// Descriptor mocks base method.
func (m *MockDevice) Descriptor() *gousb.DeviceDesc {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Descriptor")
	ret0, _ := ret[0].(*gousb.DeviceDesc)
	return ret0
}

// Descriptor indicates an expected call of Descriptor.
func (mr *MockDeviceMockRecorder) Descriptor() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Descriptor", reflect.TypeOf((*MockDevice)(nil).Descriptor))
}

// GetStringDescriptor mocks base method.
func (m *MockDevice) GetStringDescriptor(index int) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetStringDescriptor", index)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetStringDescriptor indicates an expected call of GetStringDescriptor.
func (mr *MockDeviceMockRecorder) GetStringDescriptor(index any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetStringDescriptor", reflect.TypeOf((*MockDevice)(nil).GetStringDescriptor), index)
}

// Manufacturer mocks base method.
func (m *MockDevice) Manufacturer() (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Manufacturer")
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Manufacturer indicates an expected call of Manufacturer.
func (mr *MockDeviceMockRecorder) Manufacturer() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Manufacturer", reflect.TypeOf((*MockDevice)(nil).Manufacturer))
}

// Product mocks base method.
func (m *MockDevice) Product() (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Product")
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Product indicates an expected call of Product.
func (mr *MockDeviceMockRecorder) Product() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Product", reflect.TypeOf((*MockDevice)(nil).Product))
}

// SerialNumber mocks base method.
func (m *MockDevice) SerialNumber() (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SerialNumber")
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// SerialNumber indicates an expected call of SerialNumber.
func (mr *MockDeviceMockRecorder) SerialNumber() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SerialNumber", reflect.TypeOf((*MockDevice)(nil).SerialNumber))
}

// SetAutoDetach mocks base method.
func (m *MockDevice) SetAutoDetach(autoDetach bool) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetAutoDetach", autoDetach)
	ret0, _ := ret[0].(error)
	return ret0
}

// SetAutoDetach indicates an expected call of SetAutoDetach.
func (mr *MockDeviceMockRecorder) SetAutoDetach(autoDetach any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetAutoDetach", reflect.TypeOf((*MockDevice)(nil).SetAutoDetach), autoDetach)
}

// MockContext is a mock of Context interface.
type MockContext struct {
	ctrl     *gomock.Controller
	recorder *MockContextMockRecorder
}

// MockContextMockRecorder is the mock recorder for MockContext.
type MockContextMockRecorder struct {
	mock *MockContext
}

// NewMockContext creates a new mock instance.
func NewMockContext(ctrl *gomock.Controller) *MockContext {
	mock := &MockContext{ctrl: ctrl}
	mock.recorder = &MockContextMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockContext) EXPECT() *MockContextMockRecorder {
	return m.recorder
}

// Close mocks base method.
func (m *MockContext) Close() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Close")
	ret0, _ := ret[0].(error)
	return ret0
}

// Close indicates an expected call of Close.
func (mr *MockContextMockRecorder) Close() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*MockContext)(nil).Close))
}

// IterateDevices mocks base method.
func (m *MockContext) IterateDevices(reader func(*gousb.DeviceDesc)) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IterateDevices", reader)
	ret0, _ := ret[0].(error)
	return ret0
}

// IterateDevices indicates an expected call of IterateDevices.
func (mr *MockContextMockRecorder) IterateDevices(reader any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IterateDevices", reflect.TypeOf((*MockContext)(nil).IterateDevices), reader)
}

// OpenDevice mocks base method.
func (m *MockContext) OpenDevice(vid, pid gousb.ID) (Device, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "OpenDevice", vid, pid)
	ret0, _ := ret[0].(Device)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// OpenDevice indicates an expected call of OpenDevice.
func (mr *MockContextMockRecorder) OpenDevice(vid, pid any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "OpenDevice", reflect.TypeOf((*MockContext)(nil).OpenDevice), vid, pid)
}
