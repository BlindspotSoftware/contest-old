package hsi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/insomniacslk/xjson"
	"github.com/linuxboot/contest/pkg/event/testevent"
	"github.com/linuxboot/contest/pkg/events"
	"github.com/linuxboot/contest/pkg/target"
	"github.com/linuxboot/contest/pkg/test"
	"github.com/linuxboot/contest/pkg/xcontext"
	"github.com/linuxboot/contest/plugins/teststeps/abstraction/transport"
)

const (
	ssh      = "ssh"
	local    = "local"
	bin      = "fwupdmgr"
	command  = "security"
	jsonFlag = "--json"
)

type TargetRunner struct {
	ts *TestStep
	ev testevent.Emitter
}

type Error struct {
	Msg string `json:"error"`
}

func NewTargetRunner(ts *TestStep, ev testevent.Emitter) *TargetRunner {
	return &TargetRunner{
		ts: ts,
		ev: ev,
	}
}

func (r *TargetRunner) Run(ctx xcontext.Context, target *target.Target) error {
	var outputBuf strings.Builder

	ctx, cancel := xcontext.WithTimeout(ctx, time.Duration(xjson.Duration(defaultTimeout)))
	defer cancel()

	pe := test.NewParamExpander(target)

	r.ts.writeTestStep(&outputBuf)

	transportProto, err := transport.NewTransport(r.ts.transport.Proto, []string{ssh, local}, r.ts.transport.Options, pe)
	if err != nil {
		err := fmt.Errorf("failed to create transport: %w", err)
		outputBuf.WriteString(fmt.Sprintf("%v", err))

		return events.EmitError(ctx, outputBuf.String(), target, r.ev, err)
	}

	hsiOutput, err := r.ts.runHSI(ctx, &outputBuf, transportProto)
	if err != nil {
		return events.EmitError(ctx, outputBuf.String(), target, r.ev, err)
	}

	if err := events.EmitOutput(ctx, "hsi", hsiOutput, target, r.ev); err != nil {
		return events.EmitError(ctx, outputBuf.String(), target, r.ev, err)
	}

	return events.EmitLog(ctx, outputBuf.String(), target, r.ev)
}

type FwupdSecurity struct {
	SecurityAttributes []SecurityInfo `json:"SecurityAttributes"`
	SecurityEvents     []SecurityInfo `json:"SecurityEvents"`
}

type SecurityInfo struct {
	AppstreamId        string   `json:"AppstreamId"`
	Created            int64    `json:"Created"`
	HsiLevel           int      `json:"HsiLevel"`
	HsiResult          string   `json:"HsiResult"`
	HsiResultSuccess   string   `json:"HsiResultSuccess"`
	Name               string   `json:"Name"`
	Summary            string   `json:"Summary"`
	Description        string   `json:"Description"`
	Plugin             string   `json:"Plugin"`
	Uri                string   `json:"Uri"`
	Flags              []string `json:"Flags"`
	Guid               []string `json:"Guid,omitempty"`
	BiosSettingId      string   `json:"BiosSettingId,omitempty"`
	BiosSettingCurrent string   `json:"BiosSettingCurrentValue,omitempty"`
	BiosSettingTarget  string   `json:"BiosSettingTargetValue,omitempty"`
}

type Output struct {
	Attributes []OutputAttr `json:"attributes"`
}

type OutputAttr struct {
	HsiLevel  string `json:"HsiLevel"`
	HsiResult string `json:"HsiResult"`
	// Success differs from HsiResultSuccess, it is a boolean for the success of the attribute.
	// It is the value resulting from comparing HsiResult with HsiResultSuccess
	Success     bool     `json:"Success"`
	Name        string   `json:"Name"`
	Summary     string   `json:"Summary"`
	Description string   `json:"Description"`
	Uri         string   `json:"Uri"`
	Flags       []string `json:"Flags"`
}

func (ts *TestStep) runHSI(ctx xcontext.Context, outputBuf *strings.Builder, transport transport.Transport) (json.RawMessage, error) {
	args := []string{"security", "--json"}

	proc, err := transport.NewProcess(ctx, "fwupdmgr", args, "")
	if err != nil {
		return nil, fmt.Errorf("Failed to create proc: %w", err)
	}

	writeCommand(proc.String(), outputBuf)

	stdout, stderr, err := runProcess(ctx, proc)
	if err != nil {
		return nil, fmt.Errorf("Failed to run process: %w", err)
	}

	if len(stderr) > 0 {
		outputBuf.WriteString(fmt.Sprintf("Stderr:\n%s\n", string(stderr)))
	}

	var hsiData FwupdSecurity

	if err := json.Unmarshal(stdout, &hsiData); err != nil {
		return nil, fmt.Errorf("Error unmarshalling JSON: %v", err)
	}

	hsiOutput := Output{}

	for _, attr := range hsiData.SecurityAttributes {
		outAttr := OutputAttr{
			HsiLevel:    fmt.Sprintf("Level %d", attr.HsiLevel),
			HsiResult:   attr.HsiResult,
			Success:     attr.HsiResult == attr.HsiResultSuccess,
			Name:        attr.Name,
			Summary:     attr.Summary,
			Description: attr.Description,
			Uri:         attr.Uri,
			Flags:       attr.Flags,
		}

		for _, flag := range attr.Flags {
			if flag == "missing-data" {
				outAttr.HsiResult = "unknown"
				break
			}
		}

		for _, flag := range attr.Flags {
			if flag == "runtime-issue" {
				outAttr.HsiLevel = "Runtime Suffix"
				break
			}
		}

		hsiOutput.Attributes = append(hsiOutput.Attributes, outAttr)
	}

	jsonHSI, err := json.Marshal(hsiOutput)
	if err != nil {
		return nil, err
	}

	return jsonHSI, nil
}

func runProcess(ctx xcontext.Context, proc transport.Process) ([]byte, []byte, error) {
	stdoutPipe, err := proc.StdoutPipe()
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to pipe stdout: %v", err)
	}

	stderrPipe, err := proc.StderrPipe()
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to pipe stderr: %v", err)
	}

	var wg sync.WaitGroup
	var stdout, stderr []byte
	var stdoutErr, stderrErr error

	wg.Add(1)
	go func() {
		defer wg.Done()
		stdout, stdoutErr = getOutputFromReader(stdoutPipe)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		stderr, stderrErr = getOutputFromReader(stderrPipe)
	}()

	// Start the process
	if err := proc.Start(ctx); err != nil {
		return nil, nil, fmt.Errorf("Failed to start process: %w", err)
	}

	// Wait for the process to finish
	if err := proc.Wait(ctx); err != nil {
		return nil, nil, fmt.Errorf("Failed to run hsi check: %v", err)
	}

	// Wait for the goroutines to finish reading from the pipes
	wg.Wait()

	if stdoutErr != nil {
		return nil, nil, fmt.Errorf("Failed to read from Stdout buffer: %v", stdoutErr)
	}

	if stderrErr != nil {
		return nil, nil, fmt.Errorf("Failed to read from Stderr buffer: %v", stderrErr)
	}

	return stdout, stderr, nil
}

func getOutputFromReader(reader io.Reader) ([]byte, error) {
	var buf bytes.Buffer

	_, err := io.Copy(&buf, reader)

	return buf.Bytes(), err
}
