package health

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/atsu/goat/build"
	"github.com/atsu/goat/stream"
	"github.com/atsu/goat/util"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/mem"
)

const MaxKafkaHealthCheckInterval = time.Minute * 5
const DefaultFlushTimeout = 1000 // ms

// State represents a visual indicator of the health
type State string

const (
	Blue   = State("blue")   // init or shut-down
	Green  = State("green")  // A-OK
	Yellow = State("yellow") // may need assistance
	Red    = State("red")    // HELP!
	Gray   = State("gray")   // maintenance mode
)

// Set compiles with the Flag.Value interface
func (s *State) Set(h string) error {
	switch h {
	case string(Blue):
		*s = Blue
	case string(Green):
		*s = Green
	case string(Yellow):
		*s = Yellow
	case string(Red):
		*s = Blue
	case string(Gray):
		*s = Gray
	default:
		return errors.New(fmt.Sprintf("unknown State: %s", h))
	}

	return nil
}

// Set compiles with the Flag.Value interface (and Stringer)
func (s State) String() string {
	return string(s)
}

// StateDefault health is Blue
const StateDefault = Blue

// EventType value is "health"
const EventType = "health"

// Event is the basic format for all JSON-encoded Health
// Example of a json encoded health event
//	{"hostname":"host123", "timestamp":1559761560, "etype":"health", "event":"status",
//	 "service":"gather", "version":"v0.0.1", "state":"green", "msg":"a-ok",
//	 "data": { "testmode": "off", "kernel": true }}
type Event struct {
	Hostname  string `json:"hostname"`  // Emitting host
	Timestamp int64  `json:"timestamp"` // Timestamp of event
	Type      string `json:"etype"`     // Type of event ( "health")
	Name      string `json:"event"`     // Name of event (e.g. status, startup, shutdown)
	Service   string `json:"service"`   // Operating service (e.g. gather)
	Version   string `json:"version"`   // Service version
	State     State  `json:"state"`     // Enumeration of health
	Message   string `json:"msg"`       // User actionable message

	Data interface{} `json:"data,omitempty"` // Service-specific data
}

type VersionMsg struct {
	Version            string `json:"version"`
	Build              string `json:"build"`
	Hostname           string `json:"hostname"`
	Kernel             string `json:"kernel"`
	ModuleLoaded       bool   `json:"loaded"`
	Timestamp          int64  `json:"timestamp"`
	LastEventTimestamp int64  `json:"last"`
	Count              uint64 `json:"count"`
	EventsSkipped      int    `json:"skipped"`
	Health             State  `json:"health"`
}

type HealthMsg struct {
	Hostname  string `json:"hostname"`
	Timestamp int64  `json:"timestamp"`
	Etype     string `json:"etype"`
	Event     string `json:"event"`

	Build string `json:"build,omitempty"`
}

//var HealthTopicSuffix = "system.health"

func GetHostHealthBytes(now time.Time) []byte {
	v, _ := mem.VirtualMemory()

	parts, err := disk.Partitions(false)
	check(err)

	usage := make(map[string]*disk.UsageStat)

	for _, part := range parts {
		u, err := disk.Usage(part.Mountpoint)
		check(err)
		usage[part.Mountpoint] = u
	}
	hostname, _ := os.Hostname()

	// cpu - get CPU number of cores and speed
	//cpuStat, err := cpu.Info()
	//check(err)

	times, err := cpu.Times(true)
	check(err)

	// host or machine kernel, uptime, platform Info
	hostStat, err := host.Info()
	check(err)

	// get interfaces MAC/hardware address
	//interfStat, err := net.Interfaces()
	//check(err)

	out, _ := json.Marshal(struct {
		Hostname  string `json:"hostname"`
		Timestamp int64  `json:"timestamp"`
		Etype     string `json:"etype"`
		Event     string `json:"event"`

		Host *host.InfoStat
		//Cpu   []cpu.InfoStat
		Times []cpu.TimesStat
		Vm    *mem.VirtualMemoryStat
		Usg   map[string]*disk.UsageStat
		//Net       []net.Interface
	}{hostname, now.Unix(), "health", "host", hostStat, times, v, usage})

	return out
}

// Reporter is the front door for reporting
type Reporter struct {
	// Errfn is the callback function that fires whenever there is an error.
	Errfn func(err error)

	hostname        string
	service         string
	version         string
	state           State
	message         string
	stats           sync.Map
	persistentStats sync.Map
	stdOutFallback  bool

	statFns             sync.Map
	healthCheckInterval time.Duration
	maxCheckInterval    time.Duration
	mux                 sync.Mutex
	kafkaHealthy        bool
	kafkaErr            error
	doneCh              chan struct{}
	reportTicker        *time.Ticker
	sc                  stream.KafkaStreamConfig
	topic               string
}

var _ IReporter = &Reporter{}

// NewReporter create a new reporter
func NewReporter(service, prefix, brokers string, errfn func(err error)) *Reporter {
	hn, _ := os.Hostname()
	info := build.GetInfo(service)

	return &Reporter{
		service:             service,
		hostname:            hn,
		version:             info.Version,
		state:               StateDefault,
		healthCheckInterval: time.Minute,
		maxCheckInterval:    MaxKafkaHealthCheckInterval,
		doneCh:              make(chan struct{}),
		reportTicker:        &time.Ticker{},
		sc:                  &stream.StreamConfig{Brokers: brokers, Prefix: prefix},
		Errfn:               errfn}
}

// GetKafkaLogWriter returns an io.Writer which can be used to produce to the service log stream,
// *Note* if you use pass this io.Writer to log.SetOutput(), be careful not to call any log functions
//   in the error callback, otherwise you may cause a deadlock.
// this stream is always `<prefix>.service.log`
func (r *Reporter) GetKafkaLogWriter() io.Writer {
	svcTopic := fmt.Sprintf("%s.log", r.service)
	return newLogWriter(svcTopic, r)
}

// SetStdOutFallback defaults to false. When enabled, health messages will be written to std out if kafka is unhealthy
// or if no producer is set
func (r *Reporter) SetStdOutFallback(b bool) {
	r.stdOutFallback = b
}

// SetTopic provides the ability to override the default topic.
// if never called, Initialize() will default the topic to `health.<service name>`
func (r *Reporter) SetTopic(topic string) {
	r.topic = topic
}

func (r *Reporter) SetErrFn(efn func(err error)) {
	r.Errfn = efn
}

func (r *Reporter) AddHostnameSuffix(suffix string) {
	if suffix == "" {
		return
	}

	r.hostname = fmt.Sprintf("%s-%s", r.hostname, suffix)
}

func (r *Reporter) validate() error {
	if r.sc == nil {
		return errors.New("stream config cannot be nil")
	}
	return nil
}

// Initialize must be called to create a new producer
// If Initialize fails with a fatal, non-recoverable error, it returns false and the non-nil error.
// when false is returned, kafka monitoring is disabled because no producer could be created or found
// Otherwise it returns true and and a non-nil error only if kafka is unhealthy
func (r *Reporter) Initialize() (bool, error) {
	if err := r.validate(); err != nil {
		return false, err
	}

	var p *kafka.Producer
	if r.sc.GetProducer() == nil {
		if r.topic == "" {
			r.topic = fmt.Sprintf("health.%s", r.service)
		}
		pd := r.sc.ProducerDefaults()
		err := pd.SetKey("go.delivery.reports", true)
		if err != nil {
			return false, err
		}

		r.mux.Lock()
		p, r.kafkaErr = r.sc.NewProducer(pd)
		if r.kafkaErr == nil {
			r.kafkaHealthy = true
		}
		r.mux.Unlock()
	}

	if r.kafkaErr == nil {
		r.ReportHealth()
	}
	r.monitorKafkaErrors()
	r.startHealthCheck(r.healthCheckInterval)
	return p != nil, r.kafkaErr
}

// Health returns the current health as a populated Event object
func (r *Reporter) Health() Event {
	r.mux.Lock()
	defer r.mux.Unlock()
	stats := make(map[string]interface{})
	fn := func(k interface{}, v interface{}) bool {
		if key, ok := k.(string); ok { // all keys should be strings, but just in case
			stats[key] = v
		} else {
			fmt.Printf("[warn] stat key of type (%T) is not string and will be ignored.\n", k)
		}
		return true
	}
	// persistent stats first, so we override them.
	r.persistentStats.Range(fn)
	r.stats.Range(fn)
	return Event{
		Hostname:  r.hostname,
		Timestamp: time.Now().Unix(),
		Type:      "health",
		Name:      "status",
		Service:   r.service,
		Version:   r.version,
		State:     r.state,
		Message:   r.message,
		Data:      stats,
	}
}

// SetHealth sets the current state and message
func (r *Reporter) SetHealth(state State, message string) {
	r.mux.Lock()
	defer r.mux.Unlock()
	r.state = state
	r.message = message
}

// RegisterStatFn registers a function to be called before each call of ReportHealth.
func (r *Reporter) RegisterStatFn(name string, fn func(reporter IReporter)) {
	r.statFns.Store(name, fn)
}

// ClearStatFns clears the reporter functions
func (r *Reporter) ClearStatFns() {
	r.statFns = sync.Map{}
}

// PersistStat is akin to storing a stat, but a persistent stat is a stat that does not
// clear when Clear is called.
func (r *Reporter) PersistStat(key string, val interface{}) {
	r.persistentStats.Store(key, val)
}

// AddStat adds a keyed entry to be included in reporting
func (r *Reporter) AddStat(key string, val interface{}) {
	r.stats.Store(key, val)
}

// GetStat retrieve a previously entered stat or persistent stat.
// persistent stats will act as default stats.
func (r *Reporter) GetStat(key string) interface{} {
	if val, ok := r.stats.Load(key); ok {
		return val
	}
	if val, ok := r.persistentStats.Load(key); ok {
		return val
	}
	return nil
}

// ClearStat clears an individual stat by key, does not clear persistent stats
func (r *Reporter) ClearStat(key string) {
	r.stats.Delete(key)
}

// ClearStats clears all non persistent stats
func (r *Reporter) ClearStats() {
	r.stats = sync.Map{}
}

// ClearAll clears all stats and persistent stats
func (r *Reporter) ClearAll() {
	r.ClearStats()
	r.persistentStats = sync.Map{}
}

// JsonStats returns the current set of stats as a json
func (r *Reporter) JsonStats() []byte {
	b := safeMarshal(r.Health().Data)
	return b
}

// ReportHealth sends the current health to kafka
func (r *Reporter) ReportHealth() {
	r.statFns.Range(func(_, fun interface{}) bool {
		if fn, ok := fun.(func(reporter IReporter)); ok {
			fn(r)
		}
		return true
	})
	h := r.Health()
	b := safeMarshal(h)
	_ = r.produce(r.stdOutFallback, r.topic, b) // we don't care about this error because it will get reported via callback
}

// StartIntervalReporting starts automatic reporting over the given interval.
func (r *Reporter) StartIntervalReporting(interval time.Duration) {
	r.reportTicker.Stop() // just in case
	r.reportTicker = time.NewTicker(interval)
}

func (r *Reporter) startHealthCheck(interval time.Duration) {
	healthTicker := time.NewTicker(interval)

	nextCheck := time.Time{}
	backOff := time.Duration(0)
	retry := 0
	go func() {
		for {
			select {
			case <-r.doneCh:
				healthTicker.Stop()
				return
			case <-healthTicker.C:
				if time.Now().After(nextCheck) {
					if r.checkKafkaHealth() {
						retry = 0 // reset count when health
						backOff = time.Duration(0)
					} else {
						retry++
						backOff = backOff + (interval * time.Duration(retry)) // Super simple backoff logic,
						//log.Println("Backoff:", backOff)
						if backOff > r.maxCheckInterval {
							backOff = r.maxCheckInterval
						}
						nextCheck = time.Now().Add(backOff)
					}
				}
			case <-r.reportTicker.C:
				r.ReportHealth()
			}
		}
	}()
}

// Stop interval reporting and close the kafka producer, after calling stop, the health reporter should no
// longer be used. To re-start, you should create a new reporter
// calling this function will result in emitting a default shutdown state of Gray with an empty message
// if you want to set the final state or message, use StopWithFinalState
func (r *Reporter) Stop() error {
	return r.StopWithFinalState(Gray, "")
}

// StopWithFinalState is the same as calling Stop() but takes in a final state and message,
// that will be emitted before shutdown.
func (r *Reporter) StopWithFinalState(final State, msg string) error {
	r.SetHealth(final, msg)
	r.ReportHealth()
	if r.reportTicker != nil {
		r.reportTicker.Stop()
	}
	if r.doneCh != nil {
		close(r.doneCh)
	}
	if r.sc.GetProducer() != nil {
		r.sc.Flush(DefaultFlushTimeout)
	}
	return r.sc.Close()
}

func (r *Reporter) monitorKafkaErrors() {
	go func() {
		producer := r.sc.GetProducer()
		if producer == nil {
			return
		}
		for e := range producer.Events() {
			switch ev := e.(type) {
			case kafka.Error:
				if ev.Code() == kafka.ErrAllBrokersDown {
					r.mux.Lock()
					r.kafkaHealthy = false
					r.mux.Unlock()
					r.errorHandler("kafka error", fmt.Errorf("code = %s, error = %s", ev.Code(), ev.Error()))
				}
			}
		}
	}()
}

// KafkaHealthy returns true if kafka is healthy, false if it isn't, and the last error associated with
// the unhealthy state
func (r *Reporter) KafkaHealthy() (bool, error) {
	return r.kafkaHealthy, r.kafkaErr
}

func (r *Reporter) checkKafkaHealth() bool {
	p := r.sc.GetProducer()
	if p != nil {
		r.mux.Lock()
		defer r.mux.Unlock()
		_, r.kafkaErr = p.GetMetadata(nil, true, stream.SessionTimeoutDefault)
		r.kafkaHealthy = r.kafkaErr == nil
		if r.kafkaErr != nil {
			r.errorHandler("kafka unhealthy", r.kafkaErr)
		}
	}
	// if we don't have a producer what does kafkaHealth even mean?
	// this is edge case, so for now we are just hands off.

	return r.kafkaHealthy
}

func (r *Reporter) produce(fallback bool, topic string, message []byte) error {
	fullTopic := r.sc.FullTopic(topic)
	producer := r.sc.GetProducer()

	healthy := r.kafkaHealthy

	var produceErr error
	if producer != nil { // producer can be nil if Initialize has errors
		produceErr = r.sc.Produce(&fullTopic, message)
	}

	if !healthy || produceErr != nil || producer == nil {

		if fallback {
			// Fallback write to std out
			fmt.Println(string(message))
		}

		errorMessage := ""

		switch {
		case producer == nil:
			errorMessage = "producer not set"
		case produceErr != nil:
			errorMessage = fmt.Sprintf("could not produce: %s", produceErr.Error())
		case !healthy:
			errorMessage = "kafka not available"
		}
		// todo only call the errorHandler on a legitimate produceErr != nil ?
		r.errorHandler("produce error", fmt.Errorf("[%s] %s", errorMessage, fullTopic))
	}

	return produceErr
}

// HealthHandler return the current health on demand
func (r *Reporter) HealthHandler(w http.ResponseWriter, req *http.Request) {
	p := req.URL.Query().Get("pretty")
	out := safeMarshal(r.Health())
	if pretty, _ := strconv.ParseBool(p); pretty {
		out = util.JsonPrettyPrint(out)
	}
	_, err := w.Write(out)
	// todo why call errorHandler here?
	r.errorHandler("could not write status response", err)
}

// todo it might be better to call this only when an error happens the first time,
// it's also possible we don't need this callback functionality
func (r *Reporter) errorHandler(message string, err error) {
	go func() {
		if err != nil && r.Errfn != nil {
			r.Errfn(fmt.Errorf("%s: %v", message, err.Error()))
		}
	}()
}

func safeMarshal(msg interface{}) []byte {
	b, err := json.Marshal(msg)
	if err != nil {
		b, err = json.Marshal(struct {
			Error string `json:"error"` // *note* error doesn't marshal automatically
			Msg   string `json:"msg"`
		}{
			Error: "internal",
			Msg:   err.Error(),
		})
	}
	return b
}

func check(err error) {
	if err != nil {
		log.Println(err.Error())
	}
}

type logWriter struct {
	topic    string
	reporter *Reporter
}

var _ io.Writer = &logWriter{}

func newLogWriter(topic string, reporter *Reporter) *logWriter {
	return &logWriter{
		topic:    topic,
		reporter: reporter}
}

func (k logWriter) Write(p []byte) (n int, err error) {
	out, err := json.Marshal(struct {
		App       string `json:"app"`
		Timestamp int64  `json:"timestamp"`
		Msg       string `json:"msg"`
	}{
		App:       k.reporter.service,
		Timestamp: time.Now().Unix(),
		Msg:       string(p),
	})
	if err != nil {
		return 0, err
	}
	return len(p), k.reporter.produce(false, k.topic, out)
}
