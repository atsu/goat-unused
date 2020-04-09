package bucket

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"time"
)

type UploadCallback func(string, *bytes.Buffer, interface{}, error) error
type BucketIterator func(*PipelineObject) bool

type Bucket interface {
	GetName() string
	SetName(string)
	GetOwner() string
	GetPipeline() string

	NewSession() error
	GetObject(string) (io.ReadCloser, error)
	GetObjectIfMatch(string, string) (io.ReadCloser, error)

	Path() string
	KeyPrefix(time.Time) string
	Upload(*bytes.Buffer) error
	ListObjects(string, string, BucketIterator) error

	SetFlags()
}

// NewBucket takes s3://foo or gs://bar and returns either
// S3{} or GCS{}, respectively.
//
// Note: prefix (s3:// or gs://) is stripped.
//
// The default is S3{}
//
// NewBucket returns error if illegal characters are present.
func NewBucket(name string) (Bucket, error) {
	var bucket Bucket

	// Examine name[0:5]
	switch strings.ToLower(name)[0:5] {
	default:
		fallthrough
	case "s3://":
		bucket = &S3{}
	case "gs://":
		if strings.ContainsAny(name, ".") {
			return nil, errors.New("illegal dot (.) character in name")
		}

		bucket = &GCS{}
	}

	bucket.SetName(name[5:])

	return bucket, nil
}
