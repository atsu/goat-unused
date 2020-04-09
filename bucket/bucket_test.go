package bucket

import (
	"flag"
	"os"
	"testing"

	"github.com/atsu/goat/config"
	"github.com/stretchr/testify/suite"
)

type bucketSuite struct {
	suite.Suite
}

// testBucketConfig is the default configuration.
var testBucketConfig = &S3{
	Pipeline: "pipeline.unset",
	Owner:    "unset",
	Bucket:   "atsu.unset",
	Suffix:   "json.gz",
	Name:     "",
	Objects:  0,
	Bytes:    0,
}

func (s *bucketSuite) SetupSuite() {
}

func TestBucketSuite(t *testing.T) {
	suite.Run(t, new(bucketSuite))
}

func (s *bucketSuite) TestDefaults() {
	sc := &S3{}

	sc.SetFlags()
	s.Equal(sc, testBucketConfig)
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
}

func (s *bucketSuite) TestEnvironment() {
	sc := &S3{}

	envkey := config.AtsuConfigEnvPrefix + "_NAME"
	os.Setenv(envkey, "testing")

	sc.SetFlags()
	s.NotEqual(sc, testBucketConfig)
	s.Equal("testing", sc.Name)
	os.Unsetenv(envkey)
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
}

func (s *bucketSuite) TestNewBucket() {
	b, e := NewBucket("s3://atsu.corp")

	// Verify the underlying struct type
	switch b.(type) {
	case *S3:
		s.Equal(b.GetName(), "atsu.corp")
		s.Nil(e)
	default:
		s.Equal(true, false)
	}

	b, e = NewBucket("gs://atsu-corp")
	// Verify the underlying struct type
	switch b.(type) {
	case *GCS:
		s.Nil(e)
		s.Equal(b.GetName(), "atsu-corp")
	default:
		s.Equal(true, false)
	}

	// Period is an illegal character in GCS
	b, e = NewBucket("gs://atsu.corp")
	s.NotNil(e)
}
