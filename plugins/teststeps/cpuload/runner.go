package cpuload

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/linuxboot/contest/pkg/event/testevent"
	"github.com/linuxboot/contest/pkg/events"
	"github.com/linuxboot/contest/pkg/target"
	"github.com/linuxboot/contest/pkg/test"
	"github.com/linuxboot/contest/pkg/xcontext"
	"github.com/linuxboot/contest/plugins/teststeps/abstraction/options"
	"github.com/linuxboot/contest/plugins/teststeps/abstraction/transport"
	"github.com/linuxboot/contest/plugins/teststeps/cpu"
)

const (
	supportedProto = "ssh"
	privileged     = "sudo"
	cmd            = "cpu"
	argument       = "stats"
	jsonFlag       = "--json"
)

type TargetRunner struct {
	ts *TestStep
	ev testevent.Emitter
}

func NewTargetRunner(ts *TestStep, ev testevent.Emitter) *TargetRunner {
	return &TargetRunner{
		ts: ts,
		ev: ev,
	}
}

func (r *TargetRunner) Run(ctx xcontext.Context, target *target.Target) error {
	var outputBuf strings.Builder

	ctx, cancel := options.NewOptions(ctx, defaultTimeout, r.ts.options.Timeout)
	defer cancel()

	pe := test.NewParamExpander(target)

	r.ts.writeTestStep(&outputBuf)

	transportProto, err := transport.NewTransport(r.ts.transport.Proto, []string{supportedProto}, r.ts.transport.Options, pe)
	if err != nil {
		err := fmt.Errorf("failed to create transport: %w", err)
		outputBuf.WriteString(fmt.Sprintf("%v", err))

		return events.EmitError(ctx, outputBuf.String(), target, r.ev, err)
	}

	if r.ts.Duration != "" {
		if _, err := time.ParseDuration(r.ts.Duration); err != nil {
			return fmt.Errorf("wrong interval statement, valid units are ns, us, ms, s, m and h")
		}
	}

	if err := r.ts.runLoad(ctx, &outputBuf, transportProto); err != nil {
		return events.EmitError(ctx, outputBuf.String(), target, r.ev, err)
	}

	return events.EmitLog(ctx, outputBuf.String(), target, r.ev)
}

func (ts *TestStep) runLoad(ctx xcontext.Context, outputBuf *strings.Builder, transport transport.Transport,
) error {
	var args []string

	if len(ts.Args) > 0 {
		args = append([]string{"stress-ng", "--cpu $(nproc)"}, ts.Args...)
		args = append(args, fmt.Sprintf("--timeout %s", ts.Duration), "&")
	} else {
		if len(ts.CPUs) == 0 {
			args = []string{"stress-ng", "--cpu $(nproc)", "--cpu-load 100", fmt.Sprintf("--timeout %s", ts.Duration), "&"}
		} else {
			for _, core := range ts.CPUs {
				args = append(args, []string{
					fmt.Sprintf("taskset -c %d", core), "stress-ng", "--cpu 1", "--cpu-load 100",
					fmt.Sprintf("--timeout %s", ts.Duration), "&",
				}...)
			}
		}
	}

	proc, err := transport.NewProcess(ctx, privileged, args, "")
	if err != nil {
		return fmt.Errorf("Failed to create proc: %w", err)
	}

	writeCommand(proc.String(), outputBuf)

	stdoutPipe, err := proc.StdoutPipe()
	if err != nil {
		return fmt.Errorf("Failed to pipe stdout: %v", err)
	}

	stderrPipe, err := proc.StderrPipe()
	if err != nil {
		return fmt.Errorf("Failed to pipe stderr: %v", err)
	}

	// try to start the process, if that succeeds then the outcome is the result of
	// waiting on the process for its result; this way there's a semantic difference
	// between "an error occured while launching" and "this was the outcome of the execution"
	outcome := proc.Start(ctx)
	if outcome != nil {
		return fmt.Errorf("Failed to run load test: %v.", outcome)
	}

	if len(ts.Expect.Individual) > 0 || len(ts.Expect.General) > 0 {
		if err := ts.parseStats(ctx, outputBuf, transport); err != nil {
			return fmt.Errorf("Failed to parse cpu stats: %v.", err)
		}
	}

	if outcome == nil {
		outcome = proc.Wait(ctx)
	}

	_, stderr := getOutputFromReader(stdoutPipe, stderrPipe)

	if outcome != nil {
		return fmt.Errorf("Failed to run load test: %v.\n%s\n", outcome, string(stderr))
	}

	return err
}

func (ts *TestStep) parseStats(ctx xcontext.Context, outputBuf *strings.Builder, transport transport.Transport) error {
	duration, err := time.ParseDuration(ts.Duration)
	if err != nil {
		return fmt.Errorf("wrong interval statement, valid units are ns, us, ms, s, m and h")
	}

	// Run cpu stats only for 96% of the load duration.
	interval := duration * 94 / 100
	offset := duration * 3 / 100

	// Start the cpu stats command with a small offset, so it is in the middle of the load duration.
	time.Sleep(offset)

	args := []string{
		ts.ToolPath,
		cmd,
		argument,
		fmt.Sprintf("--interval=%s", interval.String()),
		jsonFlag,
	}

	proc, err := transport.NewProcess(ctx, privileged, args, "")
	if err != nil {
		return fmt.Errorf("Failed to create proc: %w", err)
	}

	writeCommand(proc.String(), outputBuf)

	stdoutPipe, err := proc.StdoutPipe()
	if err != nil {
		return fmt.Errorf("Failed to pipe stdout: %v", err)
	}

	stderrPipe, err := proc.StderrPipe()
	if err != nil {
		return fmt.Errorf("Failed to pipe stderr: %v", err)
	}

	// try to start the process, if that succeeds then the outcome is the result of
	// waiting on the process for its result; this way there's a semantic difference
	// between "an error occured while launching" and "this was the outcome of the execution"
	outcome := proc.Start(ctx)
	if outcome == nil {
		outcome = proc.Wait(ctx)
	}

	stdout, stderr := getOutputFromReader(stdoutPipe, stderrPipe)

	if len(string(stdout)) > 0 {
		outputBuf.WriteString(fmt.Sprintf("Stats Stdout:\n%s\n", string(stdout)))
	} else if len(string(stderr)) > 0 {
		outputBuf.WriteString(fmt.Sprintf("Stats Stderr:\n%s\n", string(stderr)))
	}

	if outcome != nil {
		return fmt.Errorf("Failed to get CPU stats: %v.\nStats Stderr:\n%s\n", outcome, string(stderr))
	}

	if err = ts.parseOutput(ctx, outputBuf, stdout); err != nil {
		return err
	}

	return nil
}

// getOutputFromReader reads data from the provided io.Reader instances
// representing stdout and stderr, and returns the collected output as byte slices.
func getOutputFromReader(stdout, stderr io.Reader) ([]byte, []byte) {
	// Read from the stdout and stderr pipe readers``
	outBuffer, err := readBuffer(stdout)
	if err != nil {
		fmt.Printf("Failed to read from Stdout buffer: %v\n", err)
	}

	errBuffer, err := readBuffer(stderr)
	if err != nil {
		fmt.Printf("Failed to read from Stderr buffer: %v\n", err)
	}

	return outBuffer, errBuffer
}

// readBuffer reads data from the provided io.Reader and returns it as a byte slice.
// It dynamically accumulates the data using a bytes.Buffer.
func readBuffer(r io.Reader) ([]byte, error) {
	buf := &bytes.Buffer{}
	if _, err := io.Copy(buf, r); err != nil && err != io.EOF {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (ts *TestStep) parseOutput(ctx xcontext.Context, outputBuf *strings.Builder,
	stdout []byte,
) error {
	var (
		stats       cpu.Stats
		finalError  bool
		errorString string
	)

	if len(stdout) != 0 {
		if err := json.Unmarshal(stdout, &stats); err != nil {
			return fmt.Errorf("failed to unmarshal stdout: %v", err)
		}
	}

	for _, expect := range ts.Expect.General {
		if err := stats.CheckGeneralOption(expect, outputBuf); err != nil {
			errorString += fmt.Sprintf("failed to check general option '%s': %v\n", expect.Option, err)
			finalError = true
		}
	}

	for _, expect := range ts.Expect.Individual {
		if err := stats.CheckIndividualOption(expect, true, outputBuf); err != nil {
			errorString += fmt.Sprintf("failed to check individual option '%s': %v\n", expect.Option, err)
			finalError = true
		}
	}

	if finalError {
		return fmt.Errorf("Some expect options are not as expected:\n%s", errorString)
	}

	return nil
}
