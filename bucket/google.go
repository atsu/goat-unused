package bucket

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"time"

	"github.com/atsu/goat/config"
	"github.com/kelseyhightower/envconfig"
	"google.golang.org/api/iterator"

	"cloud.google.com/go/storage"
)

const GoogleStorageClass = "STANDARD"

// uprovides GCS-related configuration
type GCS struct {
	Pipeline string `default:"pipeline.unset" json:"pipeline"`
	Owner    string `default:"data" json:"owner"`
	Bucket   string `default:"atsu.unset" json:"bucket"`
	Suffix   string `default:"json.gz" json:"suffix"`
	Name     string `json:"name"`

	Objects int `json:"objects"`
	Bytes   int `json:"bytes"`

	session *storage.Client
}

// String returns JSON representation
func (g *GCS) String() string {
	b, _ := json.Marshal(&g)

	return string(b)
}

func (g *GCS) GetName() string {
	return g.Name
}

func (g *GCS) SetName(o string) {
	g.Name = o
}

func (g *GCS) GetOwner() string {
	return g.Owner
}

func (g *GCS) GetPipeline() string {
	return g.Pipeline
}

func (g *GCS) GetObject(key string) (io.ReadCloser, error) {
	ctx := context.TODO()

	obj := g.session.Bucket(g.Bucket).Object(key)

	return obj.NewReader(ctx)
}

func (g *GCS) GetObjectIfMatch(key string, match string) (io.ReadCloser, error) {
	return g.GetObject(key)
}

func (g *GCS) ListObjects(path string, latest string, helper BucketIterator) error {
	ctx := context.TODO()

	it := g.session.Bucket(g.Bucket).Objects(ctx, &storage.Query{
		Prefix: path})

	var err error
	for {
		oa, ierr := it.Next()
		if ierr != nil {
			err = ierr

			break
		}
		if !helper(NewPipelineObject(oa.Name, oa.Etag)) {
			break
		}
	}

	if err == iterator.Done {
		return nil
	}

	return err
}

// SetFlags to install various command-line flag(s)
func (g *GCS) SetFlags() {
	envconfig.Process(config.AtsuConfigEnvPrefix, g)

	flag.StringVar(&g.Name, "name", g.Name, fmt.Sprintf("Name (%s)", g.Name))
	flag.StringVar(&g.Bucket, "bucket", g.Bucket, fmt.Sprintf("Bucket (%s)", g.Bucket))
	flag.StringVar(&g.Owner, "owner", g.Owner, fmt.Sprintf("Owner (%s)", g.Owner))
	flag.StringVar(&g.Pipeline, "pipeline", g.Pipeline, fmt.Sprintf("Pipeline (%s)", g.Pipeline))
	flag.StringVar(&g.Suffix, "suffix", g.Suffix, fmt.Sprintf("Suffix (%s)", g.Suffix))
}

// Path generates the GCS prefix, excluding time-based prefix
func (g *GCS) Path() string {
	return fmt.Sprintf("%s/%s/%s", g.Owner, g.Pipeline, g.Name)
}

// KeyPrefix generates the GCS prefix, including time-based prefix
func (g *GCS) KeyPrefix(t time.Time) string {
	hash := fmt.Sprintf("%d/%03d/%02d", t.Year(), t.YearDay(), t.Hour())

	return fmt.Sprintf("%s/%s", g.Path(), hash)
}

// Key returns the full GCS key
func (g *GCS) Key(t time.Time) string {
	return fmt.Sprintf("%s/%s-%d.%s", g.KeyPrefix(t), g.Name,
		t.UnixNano(), g.Suffix)
}

// NewSession creates a new Google session; requires GOOGLE_APPLICATION_CREDENTIALS
func (g *GCS) NewSession() error {
	ctx := context.TODO()

	var err error

	g.session, err = storage.NewClient(ctx)
	if err != nil {
		return err
	}

	_, err = g.session.Bucket(g.Bucket).Attrs(ctx)

	if err != nil {
		return err
	}

	return err
}

func (g *GCS) GetSession() *storage.Client {
	return g.session
}

// Upload uploads the provided buffer via GCS
func (g *GCS) Upload(b *bytes.Buffer) error {
	key := g.Key(time.Now())

	return g.UploadKey(key, b)
}

func (g *GCS) UploadCallback(b *bytes.Buffer, intf interface{}, cb UploadCallback) error {
	key := g.Key(time.Now())

	return g.UploadKeyCallback(key, b, intf, cb)
}

func (g *GCS) UploadKey(key string, b *bytes.Buffer) error {
	return g.UploadKeyCallback(key, b, nil, func(_ string, _ *bytes.Buffer, _ interface{}, e error) error {
		return e
	})
}
func (g *GCS) UploadKeyCallback(key string, b *bytes.Buffer, intf interface{}, cb UploadCallback) error {
	size := b.Len()

	ctx := context.TODO()

	w := g.session.Bucket(g.Bucket).Object(key).NewWriter(ctx)
	_, err := b.WriteTo(w)

	if err == nil {
		err = w.Close()

		if err == nil {
			g.Objects += 1
			g.Bytes += size
		}
	}

	return cb(key, b, intf, err)
}
