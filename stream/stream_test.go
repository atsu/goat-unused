package stream

import (
	"encoding/json"
	"testing"

	"flag"
	"os"

	"github.com/atsu/goat/config"
	"github.com/stretchr/testify/suite"
	"gopkg.in/yaml.v2"
)

type streamSuite struct {
	suite.Suite
}

// testStreamConfig is the default configuration.
var testStreamConfig = &StreamConfig{
	Brokers:  "kafka-atsu-prod-01:9092,kafka-atsu-prod-02:9092,kafka-atsu-prod-03:9092",
	Prefix:   "atsu",
	Topic:    "unset",
	Messages: 0,
	Bytes:    0,
	Offset:   "latest",
	GroupId:  "atsu-unset-group-id",
	Codec:    "none"}

func (s *streamSuite) SetupSuite() {
}

func TestStreamSuite(t *testing.T) {
	suite.Run(t, new(streamSuite))
}

func (s *streamSuite) TestDefaults() {
	sc := &StreamConfig{}

	sc.SetFlags()
	s.Equal(sc, testStreamConfig)

	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
}

func (s *streamSuite) TestEnvironment() {
	sc := &StreamConfig{}

	envkey := config.AtsuConfigEnvPrefix + "_TOPIC"
	os.Setenv(envkey, "testing")
	envkey = config.AtsuConfigEnvPrefix + "_GROUP_ID"
	os.Setenv(envkey, "surfers")

	sc.SetFlags()
	s.NotEqual(sc, testStreamConfig)
	s.Equal("testing", sc.Topic)
	s.Equal("surfers", sc.GroupId)

	os.Unsetenv(envkey)
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
}

func (s *streamSuite) TestJsonMarshal() {
	data := `{"brokers":"testbroker","prefix":"test.prefix","topic":"sometopic","messages":123,"bytes":100,"offset":"sdfasdf1","group_id":"id123","glob":true,"reports":true,"codec":"codectest"}`

	var sc StreamConfig
	err := json.Unmarshal([]byte(data), &sc)
	if err != nil {
		s.Fail(err.Error())
	}

	out, err := json.Marshal(sc)
	if err != nil {
		s.Fail(err.Error())
	}

	s.Equal(string(data), string(out))
}

func (s *streamSuite) TestYamlMarshal() {
	data := `brokers: testbroker
prefix: test.prefix
topic: sometopic
messages: 123
bytes: 100
timeout: 0s
interval: 0s
offset: sdfasdf1
group_id: id123
glob: true
reports: true
codec: codectest
`

	var sc StreamConfig
	err := yaml.Unmarshal([]byte(data), &sc)
	if err != nil {
		s.Fail(err.Error())
	}

	out, err := yaml.Marshal(sc)
	if err != nil {
		s.Fail(err.Error())
	}

	s.Equal(string(data), string(out))
}
