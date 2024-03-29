// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import health "github.com/atsu/goat/health"
import http "net/http"
import io "io"
import mock "github.com/stretchr/testify/mock"
import time "time"

// IReporter is an autogenerated mock type for the IReporter type
type IReporter struct {
	mock.Mock
}

// AddHostnameSuffix provides a mock function with given fields: _a0
func (_m *IReporter) AddHostnameSuffix(_a0 string) {
	_m.Called(_a0)
}

// AddStat provides a mock function with given fields: key, val
func (_m *IReporter) AddStat(key string, val interface{}) {
	_m.Called(key, val)
}

// ClearAll provides a mock function with given fields:
func (_m *IReporter) ClearAll() {
	_m.Called()
}

// ClearStat provides a mock function with given fields: key
func (_m *IReporter) ClearStat(key string) {
	_m.Called(key)
}

// ClearStatFns provides a mock function with given fields:
func (_m *IReporter) ClearStatFns() {
	_m.Called()
}

// ClearStats provides a mock function with given fields:
func (_m *IReporter) ClearStats() {
	_m.Called()
}

// GetKafkaLogWriter provides a mock function with given fields:
func (_m *IReporter) GetKafkaLogWriter() io.Writer {
	ret := _m.Called()

	var r0 io.Writer
	if rf, ok := ret.Get(0).(func() io.Writer); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(io.Writer)
		}
	}

	return r0
}

// GetStat provides a mock function with given fields: key
func (_m *IReporter) GetStat(key string) interface{} {
	ret := _m.Called(key)

	var r0 interface{}
	if rf, ok := ret.Get(0).(func(string) interface{}); ok {
		r0 = rf(key)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(interface{})
		}
	}

	return r0
}

// Health provides a mock function with given fields:
func (_m *IReporter) Health() health.Event {
	ret := _m.Called()

	var r0 health.Event
	if rf, ok := ret.Get(0).(func() health.Event); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(health.Event)
	}

	return r0
}

// HealthHandler provides a mock function with given fields: w, req
func (_m *IReporter) HealthHandler(w http.ResponseWriter, req *http.Request) {
	_m.Called(w, req)
}

// Initialize provides a mock function with given fields:
func (_m *IReporter) Initialize() (bool, error) {
	ret := _m.Called()

	var r0 bool
	if rf, ok := ret.Get(0).(func() bool); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(bool)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// KafkaHealthy provides a mock function with given fields:
func (_m *IReporter) KafkaHealthy() (bool, error) {
	ret := _m.Called()

	var r0 bool
	if rf, ok := ret.Get(0).(func() bool); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(bool)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// PersistStat provides a mock function with given fields: key, val
func (_m *IReporter) PersistStat(key string, val interface{}) {
	_m.Called(key, val)
}

// RegisterStatFn provides a mock function with given fields: name, fn
func (_m *IReporter) RegisterStatFn(name string, fn func(health.IReporter)) {
	_m.Called(name, fn)
}

// ReportHealth provides a mock function with given fields:
func (_m *IReporter) ReportHealth() {
	_m.Called()
}

// SetErrFn provides a mock function with given fields: efn
func (_m *IReporter) SetErrFn(efn func(error)) {
	_m.Called(efn)
}

// SetHealth provides a mock function with given fields: state, message
func (_m *IReporter) SetHealth(state health.State, message string) {
	_m.Called(state, message)
}

// SetStdOutFallback provides a mock function with given fields: b
func (_m *IReporter) SetStdOutFallback(b bool) {
	_m.Called(b)
}

// SetTopic provides a mock function with given fields: topic
func (_m *IReporter) SetTopic(topic string) {
	_m.Called(topic)
}

// StartIntervalReporting provides a mock function with given fields: interval
func (_m *IReporter) StartIntervalReporting(interval time.Duration) {
	_m.Called(interval)
}

// Stop provides a mock function with given fields:
func (_m *IReporter) Stop() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// StopWithFinalState provides a mock function with given fields: final, msg
func (_m *IReporter) StopWithFinalState(final health.State, msg string) error {
	ret := _m.Called(final, msg)

	var r0 error
	if rf, ok := ret.Get(0).(func(health.State, string) error); ok {
		r0 = rf(final, msg)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
