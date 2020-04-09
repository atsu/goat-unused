package health

import (
	"io"
	"net/http"
	"time"
)

type IReporter interface {
	Initialize() (bool, error)
	Health() Event
	SetStdOutFallback(b bool)
	SetErrFn(efn func(err error))
	SetHealth(state State, message string)
	SetTopic(topic string)
	PersistStat(key string, val interface{})
	RegisterStatFn(name string, fn func(reporter IReporter))
	AddStat(key string, val interface{})
	GetStat(key string) interface{}
	ClearStat(key string)
	ClearStats()
	ClearStatFns()
	ClearAll()
	GetKafkaLogWriter() io.Writer
	KafkaHealthy() (bool, error)
	ReportHealth()
	AddHostnameSuffix(string)
	StartIntervalReporting(interval time.Duration)
	HealthHandler(w http.ResponseWriter, req *http.Request)
	Stop() error
	StopWithFinalState(final State, msg string) error
}
