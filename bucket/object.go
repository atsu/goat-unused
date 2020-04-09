package bucket

import (
	"strconv"
	"strings"
	"time"

	"github.com/atsu/goat/stream"
)

// PipelineObject represents an individual object stored in a Bucket
type PipelineObject struct {
	Owner    string
	Pipeline string
	Name     string
	Date     time.Time
	File     string

	Key  string
	Etag string
}

// objectPathCount owner/pipeline/prefix/Y/DOY/H/topic-time_in_ns.suffix
const objectPathCount = 7
const objectStrNanoLength = 19
const objectStrSuffixLength = 8

// NewPipelineObject parses an S3 key  and returns an PipelineObject
func NewPipelineObject(key string, etag string) *PipelineObject {
	p := strings.Split(key, "/")

	if len(p) != objectPathCount {
		return nil
	}

	file := p[6]

	// Unix Nano is at position ..../<topic>-<nano>.json.gz
	start := len(file) - (objectStrNanoLength + objectStrSuffixLength) // <nano>.json.gz
	end := len(file) - objectStrSuffixLength                           //.json.gz

	nano, _ := strconv.ParseInt(p[6][start:end], 10, 64)

	// time in nano seconds
	when := time.Unix(0, nano)

	obj := PipelineObject{
		Key:      key,
		Etag:     etag,
		Owner:    p[0],
		Pipeline: p[1],
		Name:     p[2],
		Date:     when,
		File:     file}

	return &obj
}

// Ensure this object belongs to us
func (oc PipelineObject) MatchesBucket(bc Bucket) bool {
	return oc.Pipeline == bc.GetPipeline() &&
		oc.Owner == bc.GetOwner() &&
		oc.Name == bc.GetName()
}

// Deprecated: Use MatchesBucket instead
func (oc PipelineObject) MatchesStream(sc *stream.StreamConfig) bool {
	return oc.Name == sc.Topic
}
