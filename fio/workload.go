package fio

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strconv"
)

const nfsTestIO_OutputMatcher = `(?m)\d{2}:\d{2}:\d{2}\.\d+\s+([^:]+):\s+(\d+)($|,)`

type WorkLoadOperationCount struct {
	Open    int64 `json:"OPEN,omitempty"`
	Opendgr int64 `json:"OPENDGR,omitempty"`
	Close   int64 `json:"CLOSE,omitempty"`
	OSync   int64 `json:"OSYNC,omitempty"`
	Read    int64 `json:"READ,omitempty"`
	Write   int64 `json:"WRITE,omitempty"`
	FSync   int64 `json:"FSYNC,omitempty"`
	Rename  int64 `json:"RENAME,omitempty"`
	Remove  int64 `json:"REMOVE,omitempty"`
	Trunc   int64 `json:"TRUNC,omitempty"`
	FTrunc  int64 `json:"FTRUNC,omitempty"`
	Link    int64 `json:"LINK,omitempty"`
	SLink   int64 `json:"SLINK,omitempty"`
	ReadDir int64 `json:"READDIR,omitempty"`
	Lock    int64 `json:"LOCK,omitempty"`
	TLock   int64 `json:"TLOCK,omitempty"`
	UnLock  int64 `json:"UNLOCK,omitempty"`
	Errors  int64 `json:"ERRORS,omitempty"`
}

type WorkLoadConfig struct {
	logLevel   string
	readPct    string
	writePct   string
	rdwrPct    string
	randioPct  string
	ioDelay    string
	direct     bool
	rdwrOnly   bool
	runSeconds string
}

func RunDefaultWorkload(dir string) WorkLoadOperationCount {
	wc := NewDefaultWorkloadConfig()
	return wc.GenerateWorkload(dir)
}

func NewDefaultWorkloadConfig() *WorkLoadConfig {
	return &WorkLoadConfig{
		logLevel:   "info",
		readPct:    "40.0",
		writePct:   "40.0",
		rdwrPct:    "20.0",
		randioPct:  "50.0",
		ioDelay:    "0.0",
		direct:     true,
		runSeconds: "5"}
}

func (w WorkLoadConfig) parseOpts() []string {
	options := []string{
		fmt.Sprintf("--read=%s", w.readPct),
		fmt.Sprintf("--write=%s", w.writePct),
		fmt.Sprintf("--rdwr=%s", w.rdwrPct),
		fmt.Sprintf("--randio=%s", w.randioPct),
		fmt.Sprintf("--iodelay=%s", w.ioDelay),
	}

	if w.direct {
		options = append(options, "--direct")
	}
	if w.rdwrOnly {
		options = append(options, "--rdwronly")
	}

	return options
}

func (w WorkLoadConfig) GenerateWorkload(directory string) WorkLoadOperationCount {
	nfsTestPath, err := exec.LookPath("nfstest_io")
	if err != nil {
		log.Fatal("'nfstest_io' is required and not found! Is PATH set correctly? " +
			"Try executing 'which nfstest_io' on the host machine.")
	}

	params := []string{
		"-d", directory,
		"-v", w.logLevel,
		"-r", w.runSeconds,
	}
	params = append(params, w.parseOpts()...)

	cmd := exec.Command(nfsTestPath, params...)
	cmd.Env = os.Environ()

	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatal(err)
	}

	return parseOutputStats(string(out))
}

func parseOutputStats(output string) WorkLoadOperationCount {
	reg, err := regexp.Compile(nfsTestIO_OutputMatcher)
	if err != nil {
		panic(err)
	}
	results := reg.FindAllStringSubmatch(output, -1)
	buf := bytes.NewBufferString("{")
	for index, r := range results {
		if index > 0 {
			buf.WriteString(",")
		}
		intVal, err := strconv.ParseInt(r[2], 10, 64)
		if err != nil {
			panic(err)
		}
		buf.WriteString(fmt.Sprintf("%q:%d", r[1], intVal))
	}
	buf.WriteString("}")

	var opCounts WorkLoadOperationCount
	if err := json.Unmarshal(buf.Bytes(), &opCounts); err != nil {
		log.Fatal(err)
	}

	return opCounts
}
