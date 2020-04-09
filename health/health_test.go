package health

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"

	"github.com/atsu/goat/stream/mocks"
	"github.com/stretchr/testify/mock"

	"github.com/atsu/goat/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type healthSuite struct {
	suite.Suite
}

func (s *healthSuite) SetupSuite() {
}

func TestHealthSuite(t *testing.T) {
	suite.Run(t, new(healthSuite))
}

func (s *healthSuite) TestDefaults() {

	s.Equal("green", string(Green))
	s.Equal("blue", string(Blue))
	s.Equal("yellow", string(Yellow))
	s.Equal("red", string(Red))
	s.Equal("gray", string(Gray))

}

func generateFields(msg Event) (map[string]interface{}, error) {
	output, err := json.Marshal(msg)
	if err != nil {
		return nil, err
	}

	var fields map[string]interface{}
	err = json.Unmarshal(output, &fields)
	if err != nil {
		return nil, err
	}

	return fields, nil
}

var requiredFields = [...]string{"hostname", "timestamp", "etype", "event", "service", "version", "state", "msg"}

func (s *healthSuite) TestRequiredMessageFields() {

	fields, err := generateFields(Event{})
	s.Equal(err, nil)

	for _, field := range requiredFields {
		_, ok := fields[field]

		s.Equal(ok, true)
	}

}

var reservedFields = [...]string{"data"}

func (s *healthSuite) TestReservedMessageFields() {
	fields, err := generateFields(Event{})
	s.Equal(err, nil)

	for _, field := range reservedFields {
		_, ok := fields[field]

		s.Equal(ok, false)
	}
}

func (s *healthSuite) TestOptionalDataField() {

	data := make(map[string]interface{})

	data["magic"] = Green

	fields, err := generateFields(Event{Data: data})
	s.Equal(err, nil)

	data = fields["data"].(map[string]interface{})

	s.Equal(data["magic"].(string), string(Green))
}

func (s *healthSuite) TestNewReporter() {
	r := NewReporter("service", "test", "test", func(err error) {})
	m := r.Health()
	h, _ := os.Hostname()
	s.Equal(h, m.Hostname)
	s.NotZero(m.Timestamp)
	s.Equal("health", m.Type)
	s.Equal("status", m.Name)
	s.Equal("service", m.Service)
	s.NotEmpty(m.Version)
	s.Equal(StateDefault, m.State)
	s.Equal("", m.Message)
	s.IsType(map[string]interface{}{}, m.Data)
	s.NoError(r.Stop())
}

func (s *healthSuite) TestSetHealth() {
	r := NewReporter("test", "test", "test", func(err error) {})
	s.Equal(Blue, r.Health().State)
	r.SetHealth(Green, "go green!")

	s.Equal(Green, r.Health().State)
	s.Equal("go green!", r.Health().Message)

	s.NoError(r.Stop())
	s.Equal(Gray, r.Health().State)
}

func (s *healthSuite) TestHealthStats() {
	r := NewReporter("test", "test", "test", func(err error) {})

	// set a stat
	r.AddStat("key", "val")
	data, ok := r.Health().Data.(map[string]interface{})
	if !ok {
		s.Fail("invalid Data field type")
	}
	s.Equal("val", data["key"])

	// persist the same stat, but since it already has a value. It should just get masked
	r.PersistStat("key", "default")
	data, ok = r.Health().Data.(map[string]interface{})
	if !ok {
		s.Fail("invalid Data field type")
	}
	s.Equal("val", data["key"]) // persistent stat is always lower in priority to regular stat

	r.ClearStats() // clears regular stats

	data, ok = r.Health().Data.(map[string]interface{})
	if !ok {
		s.Fail("invalid Data field type")
	}
	s.Equal("default", data["key"]) // verify persistent

	s.NoError(r.Stop())
}

func (s *healthSuite) TestHostnameSuffix() {
	r := NewReporter("test", "test", "test", func(err error) {})

	suffix := "test.suffix"

	hn := r.Health().Hostname
	r.AddHostnameSuffix(suffix)
	nhn := r.Health().Hostname

	s.True(strings.HasSuffix(nhn, fmt.Sprintf("-%s", suffix)))

	s.Equal(nhn, fmt.Sprintf("%s-%s", hn, suffix))

	s.NoError(r.Stop())
}

func (s *healthSuite) TestStats() {
	r := NewReporter("test", "test", "test", func(err error) {})

	h := r.Health()
	s.Empty(h.Data)

	r.AddStat("one", "s")
	r.AddStat("two", "s")
	s.Equal("s", r.GetStat("one"))

	r.ClearStat("one")
	s.Nil(r.GetStat("one"))
	s.NotNil(r.GetStat("two"))

	r.ClearStats()
	s.Nil(r.GetStat("two"))

	s.NoError(r.Stop())
}

func (s *healthSuite) TestHealthHandler() {
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		s.Error(err)
	}

	rr := httptest.NewRecorder()

	r := NewReporter("test", "test", "test", func(err error) {})

	handler := http.HandlerFunc(r.HealthHandler)

	// keeping these two calls as close as possible to prevent timestamp changes
	want := r.Health()
	handler.ServeHTTP(rr, req)

	var got Event
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		s.Error(err)
	}

	s.Equal(want, got)

	s.NoError(r.Stop())
}

func (s *healthSuite) TestHealthHandlerPretty() {
	req, err := http.NewRequest("GET", "/?pretty=true", nil)
	if err != nil {
		s.Error(err)
	}

	rr := httptest.NewRecorder()

	r := NewReporter("test", "test", "test", func(err error) {})

	handler := http.HandlerFunc(r.HealthHandler)

	// keeping these two calls as close as possible to prevent timestamp changes
	h := r.Health()
	handler.ServeHTTP(rr, req)

	// Marshal and prettify
	b, err := json.Marshal(h)
	if err != nil {
		s.Error(err)
	}
	want := util.JsonPrettyPrint(b)

	// got should already be marshaled and pretty because 'pretty=true' query param
	got := rr.Body.Bytes()

	s.Equal(want, got)

	s.NoError(r.Stop())
}

func TestReporter_Stop(t *testing.T) {
	r := NewReporter("test", "test", "0.0.0.0:9092", func(error) {})
	kMock := new(mocks.KafkaStreamConfig)
	kMock.On("Close").Return(nil)
	kMock.On("Flush", mock.AnythingOfType("int")).Return(1)
	kMock.On("FullTopic", mock.AnythingOfType("string")).Return("test")
	kMock.On("GetProducer").Return(&kafka.Producer{})
	kMock.On("Produce", mock.AnythingOfType("*string"), mock.MatchedBy(func(b []uint8) bool {
		evt := unmarshalEvent(t, b)

		// verify message and state match the default state
		return evt.Message == "" && evt.State == Gray
	})).Return(nil)
	r.sc = kMock // replace stream config with the mock

	assert.NoError(t, r.Stop())
}

func TestReporter_StopWithFinalState(t *testing.T) {
	tests := []struct {
		state State
		msg   string
	}{
		{Blue, util.RandomString(10)},
		{Green, util.RandomString(10)},
		{Yellow, util.RandomString(10)},
		{Red, util.RandomString(10)},
		{Gray, util.RandomString(10)},
	}
	for _, test := range tests {
		t.Run(string(test.state), func(t *testing.T) {
			r := NewReporter("test", "test", "0.0.0.0:9092", func(error) {})
			kMock := new(mocks.KafkaStreamConfig)
			kMock.On("Close").Return(nil)
			kMock.On("Flush", mock.AnythingOfType("int")).Return(1)
			kMock.On("FullTopic", mock.AnythingOfType("string")).Return("test")
			kMock.On("GetProducer").Return(&kafka.Producer{})
			kMock.On("Produce", mock.AnythingOfType("*string"), mock.MatchedBy(func(b []uint8) bool {
				evt := unmarshalEvent(t, b)
				// verify message and state get matched in the produced event
				return evt.Message == test.msg && evt.State == test.state
			})).Return(nil)

			r.sc = kMock // replace stream config with the mock

			assert.NoError(t, r.StopWithFinalState(test.state, test.msg))
		})
	}
}

func TestPersistentStats(t *testing.T) {
	r := NewReporter("test", "test", "0.0.0.0:9092", func(error) {})
	// *note* no need to initialize because we aren't touching kafka

	assert.Empty(t, r.GetStat("key"))

	r.PersistStat("key", "default")

	assert.Equal(t, "default", r.GetStat("key"))

	r.AddStat("key", "value")

	assert.Equal(t, "value", r.GetStat("key"))

	r.ClearStat("key")

	assert.Equal(t, "default", r.GetStat("key"))

	r.AddStat("key", "value2")

	r.ClearStats()

	assert.Equal(t, "default", r.GetStat("key"))

	r.ClearAll()

	assert.Empty(t, r.GetStat("key"))

	assert.NoError(t, r.Stop())
}

func TestReporter_KafkaHealthy(t *testing.T) {
	r := NewReporter("test", "test", "0.0.0.0:9092", func(error) {})
	scMock := new(mocks.KafkaStreamConfig)
	scMock.On("FullTopic", mock.AnythingOfType("string")).Return("test")
	scMock.On("GetProducer").Return(nil)
	scMock.On("ProducerDefaults").Return(&kafka.ConfigMap{})
	scMock.On("NewProducer", mock.AnythingOfType("*kafka.ConfigMap")).Return(nil, errors.New("test"))
	scMock.On("Produce", mock.AnythingOfType("*string"), mock.AnythingOfType("[]byte")).Return(errors.New("test"))
	r.sc = scMock

	ok, err := r.Initialize()
	assert.False(t, ok)
	assert.Error(t, err)

	healthy, err := r.KafkaHealthy()
	assert.False(t, healthy)
	assert.Error(t, err)
}

func TestBackOff(t *testing.T) {
	t.SkipNow()
	r := NewReporter("test", "test", "0.0.0.0:9092", func(err error) { fmt.Println("err:", err) })
	r.maxCheckInterval = time.Second * 5
	r.healthCheckInterval = time.Millisecond * 100
	_, err := r.Initialize()
	assert.Error(t, err)

	writer := r.GetKafkaLogWriter()
	//r.SetStdOutFallback(true)
	for i := 0; i < 2000; i++ {
		//fmt.Println("Calling Write")
		_, err := writer.Write([]byte("testing!"))
		if err != nil {
			fmt.Println(err)
		}
		fmt.Printf("Time:%s\n", time.Now())
		//time.Sleep(time.Second)
	}

	assert.NoError(t, r.Stop())
}

func TestReporter_AddFunc(t *testing.T) {
	r := NewReporter("test", "test", "0.0.0.0:9092", func(err error) { fmt.Println("err:", err) })
	scMock := new(mocks.KafkaStreamConfig)
	scMock.On("FullTopic", mock.AnythingOfType("string")).Return("test")
	scMock.On("GetProducer").Return(&kafka.Producer{})
	scMock.On("ProducerDefaults").Return(&kafka.ConfigMap{})
	scMock.On("NewProducer", mock.AnythingOfType("*kafka.ConfigMap")).Return(&kafka.Producer{}, nil)
	scMock.On("Produce", mock.AnythingOfType("*string"), mock.AnythingOfType("[]uint8")).Return(nil)
	r.sc = scMock
	key := util.RandomString(10)
	val := util.RandomString(10)
	r.RegisterStatFn("test", func(reporter IReporter) {
		r.PersistStat(key, val)
	})
	r.ReportHealth()
	scMock.AssertCalled(t, "Produce", mock.AnythingOfType("*string"), mock.MatchedBy(func(b []byte) bool {
		var evt Event
		if err := json.Unmarshal(b, &evt); err != nil {
			t.Fatal(err)
		}
		if data, ok := evt.Data.(map[string]interface{}); ok {
			return data[key] == val
		}
		return false
	}))
}

func TestRunReporter(t *testing.T) {
	t.SkipNow()
	/*
		This is example usage of the health reporter
		as well as a test that can be pointed to kafka
		and run for a simple manual verification.
	*/

	// Define an error function for notification when an error occurs
	// *Note* since we have `log.SetOutput(r.GetKafkaLogWriter())` below,
	// we cannot use `log.Println` in the error handler.
	eFn := func(err error) {
		fmt.Println("Error Occurred:", err)
	}

	// Create a new health reporter
	r := NewReporter("test", "test", "0.0.0.0:9092", eFn)

	// Initialize is called to create a new producer.
	if ok, err := r.Initialize(); !ok {
		t.Error(err)
	} else if err != nil {
		log.Println("NonFatal Init error:", err)
	}

	// Start interval reporting every second
	r.StartIntervalReporting(time.Second)
	time.Sleep(time.Second)

	// Clear the old stat, and add a new one.
	r.PersistStat("Stat", "a default value")
	time.Sleep(time.Second)

	// Add stats that should show up on the next report
	r.AddStat("Stat", "Woo")
	time.Sleep(time.Second)

	// clearing this stat results back to the default value
	r.ClearStat("Stat")
	r.AddStat("new", "stat")
	time.Sleep(time.Second)

	// clear all but persistent stats
	r.ClearStats()
	time.Sleep(time.Second)

	r.ClearAll()
	time.Sleep(time.Second)

	// If brokers are down, we need to have StdOutFallback enabled to redirect the logWriter to std.Out
	log.Println("testing with fallback!")
	r.SetStdOutFallback(true)
	r.ReportHealth()
	time.Sleep(time.Second)

	// GetLogWriter example usage
	log.SetOutput(r.GetKafkaLogWriter()) // replace the std log output
	log.Println("testing!")              // use log. as normal, results go to <prefix>.<service>.log

	time.Sleep(time.Second)
	// Be sure to stop the the health reporter as a shutdown step.
	assert.NoError(t, r.Stop())
}

func TestSafeMarshal(t *testing.T) {
	goodOut := safeMarshal(struct {
		One   string
		Two   int
		Three float32
	}{
		"one",
		2,
		3.0,
	})
	assert.Equal(t, `{"One":"one","Two":2,"Three":3}`, string(goodOut))

	badOutput := safeMarshal(struct {
		Chan chan int // can't serialize a channel
	}{
		Chan: make(chan int),
	})
	assert.Equal(t, `{"error":"internal","msg":"json: unsupported type: chan int"}`, string(badOutput))
}

func unmarshalEvent(t *testing.T, b []byte) Event {
	t.Helper()
	var evt Event
	err := json.Unmarshal(b, &evt)
	if err != nil {
		t.Error(err)
	}
	return evt
}
