package bucket

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"time"

	"github.com/atsu/goat/config"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/kelseyhightower/envconfig"
)

const DefaultStorageClass = "STANDARD"

// S3 provides S3-related configuration
type S3 struct {
	Pipeline string `default:"pipeline.unset" json:"pipeline"`
	Owner    string `default:"unset" json:"owner"`
	Bucket   string `default:"atsu.unset" json:"bucket"`
	Suffix   string `default:"json.gz" json:"suffix"`
	Name     string `json:"name"`

	Objects int `json:"objects"`
	Bytes   int `json:"bytes"`

	session *session.Session
}

// String returns JSON representation
func (s *S3) String() string {
	b, _ := json.Marshal(&s)

	return string(b)
}

func (s *S3) GetName() string {
	return s.Name
}

func (s *S3) SetName(o string) {
	s.Name = o
}

func (s *S3) GetOwner() string {
	return s.Owner
}

func (s *S3) GetPipeline() string {
	return s.Pipeline
}

func (s *S3) GetObject(key string) (io.ReadCloser, error) {
	svc := s3.New(s.GetSession())

	out, err := svc.GetObject(&s3.GetObjectInput{Bucket: &s.Bucket,
		Key: &key})

	return out.Body, err
}

func (s *S3) GetObjectIfMatch(key string, match string) (io.ReadCloser, error) {
	svc := s3.New(s.GetSession())

	out, err := svc.GetObject(&s3.GetObjectInput{Bucket: &s.Bucket,
		Key: &key, IfMatch: &match})

	if err == nil {
		s.Bytes += int(*out.ContentLength)
		s.Objects += 1
	}

	return out.Body, err
}

func (s *S3) ListObjects(path string, latest string, iterator BucketIterator) error {
	svc := s3.New(s.GetSession())

	helper := func(page *s3.ListObjectsV2Output, _ bool) bool {
		for _, c := range page.Contents {
			if !iterator(NewPipelineObject(*c.Key, *c.ETag)) {
				return false
			}
		}

		return true
	}

	err := svc.ListObjectsV2Pages(&s3.ListObjectsV2Input{Bucket: &s.Bucket,
		Prefix: &path, StartAfter: &latest}, helper)

	return err
}

// SetFlags to install various command-line flag(s)
func (s *S3) SetFlags() {
	envconfig.Process(config.AtsuConfigEnvPrefix, s)

	flag.StringVar(&s.Name, "name", s.Name, fmt.Sprintf("Name (%s)", s.Name))
	flag.StringVar(&s.Bucket, "bucket", s.Bucket, fmt.Sprintf("Bucket (%s)", s.Bucket))
	flag.StringVar(&s.Owner, "owner", s.Owner, fmt.Sprintf("Owner (%s)", s.Owner))
	flag.StringVar(&s.Pipeline, "pipeline", s.Pipeline, fmt.Sprintf("Pipeline (%s)", s.Pipeline))
	flag.StringVar(&s.Suffix, "suffix", s.Suffix, fmt.Sprintf("Suffix (%s)", s.Suffix))
}

// Path generates the S3 prefix, excluding time-based prefix
func (s *S3) Path() string {
	return fmt.Sprintf("%s/%s/%s", s.Owner, s.Pipeline, s.Name)
}

// KeyPrefix generates the S3 prefix, including time-based prefix
func (s *S3) KeyPrefix(t time.Time) string {
	hash := fmt.Sprintf("%d/%03d/%02d", t.Year(), t.YearDay(), t.Hour())

	return fmt.Sprintf("%s/%s", s.Path(), hash)
}

// Key returns the full S3 key
func (s *S3) Key(t time.Time) string {
	return fmt.Sprintf("%s/%s-%d.%s", s.KeyPrefix(t), s.Name,
		t.UnixNano(), s.Suffix)
}

func (s *S3) GetCredentials() *credentials.Credentials {
	return credentials.NewChainCredentials(
		[]credentials.Provider{
			&credentials.EnvProvider{},
			&credentials.SharedCredentialsProvider{Filename: "", Profile: "atsu.io"},
			/*
				XXX

				Need recursive session creation to support this.

				&ec2rolecreds.EC2RoleProvider{
					Client: ec2metadata.New(sess),
				},
			*/
			&credentials.SharedCredentialsProvider{Filename: "", Profile: ""},
		})
}

func (s *S3) sessionDefaults() *aws.Config {
	return &aws.Config{
		Region:      aws.String("us-west-2"),
		Credentials: s.GetCredentials(),
	}
}

// NewSession creates a new AWS session using "atsu.io" credentials
func (s *S3) NewSession() error {
	ac := s.sessionDefaults()

	s.session = session.Must(session.NewSession(ac))

	return nil
}

func (s *S3) GetSession() *session.Session {
	return s.session
}

// Upload uploads the provided buffer via S3
func (s *S3) Upload(b *bytes.Buffer) error {
	key := s.Key(time.Now())

	return s.UploadKey(key, b)
}

func (s *S3) UploadCallback(b *bytes.Buffer, intf interface{}, cb UploadCallback) error {
	key := s.Key(time.Now())

	return s.UploadKeyCallback(key, b, intf, cb)
}

func (s *S3) UploadKey(key string, b *bytes.Buffer) error {
	return s.UploadKeyCallback(key, b, nil, func(_ string, _ *bytes.Buffer, _ interface{}, e error) error {
		return e
	})
}
func (s *S3) UploadKeyCallback(key string, b *bytes.Buffer, intf interface{}, cb UploadCallback) error {
	size := b.Len()

	// XXX Using Standard for now.
	uploader := s3manager.NewUploader(s.session)
	_, err := uploader.Upload(&s3manager.UploadInput{
		Bucket:       aws.String(s.Bucket),
		Key:          aws.String(key),
		Body:         b,
		StorageClass: aws.String(DefaultStorageClass)})

	if err == nil {
		s.Objects += 1
		s.Bytes += size
	}
	return cb(key, b, intf, err)
}
