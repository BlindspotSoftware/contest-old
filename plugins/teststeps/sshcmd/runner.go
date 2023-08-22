package sshcmd

import (
	"bytes"
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"

	"github.com/insomniacslk/xjson"
	"github.com/linuxboot/contest/pkg/event/testevent"
	"github.com/linuxboot/contest/pkg/target"
	"github.com/linuxboot/contest/pkg/test"
	"github.com/linuxboot/contest/pkg/xcontext"
	"github.com/linuxboot/contest/plugins/teststeps/abstraction/transport"
)

type outcome error

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

	// limit the execution time if specified
	var cancel xcontext.CancelFunc

	if r.ts.Options.Timeout != 0 {
		ctx, cancel = xcontext.WithTimeout(ctx, time.Duration(r.ts.Options.Timeout))
		defer cancel()
	} else {
		r.ts.Options.Timeout = xjson.Duration(defaultTimeout)
		ctx, cancel = xcontext.WithTimeout(ctx, time.Duration(r.ts.Options.Timeout))
		defer cancel()
	}

	pe := test.NewParamExpander(target)

	var params inputStepParams

	if err := pe.ExpandObject(r.ts.inputStepParams, &params); err != nil {
		return err
	}

	transport, err := transport.NewTransport(params.Transport.Proto, params.Transport.Options, pe)
	if err != nil {
		return fmt.Errorf("failed to create transport: %w", err)
	}

	writeTestStep(r.ts, &outputBuf)

	_, err = r.ts.runCMD(ctx, &outputBuf, target, transport)
	if err != nil {
		outputBuf.WriteString(fmt.Sprintf("%v\n", err))

		return emitStderr(ctx, outputBuf.String(), target, r.ev, err)
	}

	return emitStdout(ctx, outputBuf.String(), target, r.ev)
}

func (ts *TestStep) runCMD(ctx xcontext.Context, outputBuf *strings.Builder, target *target.Target,
	transport transport.Transport,
) (outcome, error) {
	proc, err := transport.NewProcess(ctx, ts.Bin.Executable, ts.Bin.Args, ts.Bin.WorkingDir)
	if err != nil {
		err := fmt.Errorf("Failed to create proc: %w", err)
		outputBuf.WriteString(fmt.Sprintf("%v\n", err))

		return nil, err
	}

	writeCommand(proc.String(), outputBuf, outputBuf)

	outputBuf.WriteString("Command Stderr:\n")
	outputBuf.WriteString("Command Stdout:\n")

	stdoutPipe, err := proc.StdoutPipe()
	if err != nil {
		err := fmt.Errorf("failed to pipe stdout: %v", err)
		outputBuf.WriteString(fmt.Sprintf("%v\n", err))

		return nil, err
	}

	stderrPipe, err := proc.StderrPipe()
	if err != nil {
		err := fmt.Errorf("failed to pipe stderr: %v", err)
		outputBuf.WriteString(fmt.Sprintf("%v\n", err))

		return nil, err
	}

	// try to start the process, if that succeeds then the outcome is the result of
	// waiting on the process for its result; this way there's a semantic difference
	// between "an error occured while launching" and "this was the outcome of the execution"
	outcome := proc.Start(ctx)
	if outcome == nil {
		outcome = proc.Wait(ctx)
	}

	stdout, stderr := getOutputFromReader(stdoutPipe, stderrPipe, outputBuf)

	if outcome != nil {
		return nil, fmt.Errorf("Error executing command: %v.\nLogs:\n%s\n", outcome, string(stderr))
	}

	outputBuf.WriteString(fmt.Sprintf("%s\n", string(stdout)))

	if err = ts.parseOutput(outputBuf, stdout); err != nil {
		return nil, err
	}

	return outcome, err
}

// getOutputFromReader reads data from the provided io.Reader instances
// representing stdout and stderr, and returns the collected output as byte slices.
func getOutputFromReader(stdout, stderr io.Reader, outputBuf *strings.Builder) ([]byte, []byte) {
	// Read from the stdout and stderr pipe readers
	stdoutBuffer, err := readBuffer(stdout)
	if err != nil {
		outputBuf.WriteString(fmt.Sprintf("Failed to read from Stdout buffer: %v\n", err))
	}

	stderrBuffer, err := readBuffer(stderr)
	if err != nil {
		outputBuf.WriteString(fmt.Sprintf("Failed to read from Stderr buffer: %v\n", err))
	}

	return stdoutBuffer, stderrBuffer
}

// readBuffer reads data from the provided io.Reader and returns it as a byte slice.
// It dynamically accumulates the data using a bytes.Buffer.
func readBuffer(r io.Reader) ([]byte, error) {
	buf := &bytes.Buffer{}
	_, err := io.Copy(buf, r)
	if err != nil && err != io.EOF {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (ts *TestStep) parseOutput(outputBuf *strings.Builder, stdout []byte) error {
	var err error

	for _, expect := range ts.expectStepParams {
		re, err := regexp.Compile(expect.Regex)
		if err != nil {
			err = fmt.Errorf("Failed to parse the regex: %v", err)
			outputBuf.WriteString(err.Error())
		}

		matches := re.FindAll(stdout, -1)
		if len(matches) > 0 {
			outputBuf.WriteString(fmt.Sprintf("Found the expected string in Stdout: '%s'\n", expect))
		} else {
			err = fmt.Errorf("Could not find the expected string '%s' in Stdout.\n", expect)
			outputBuf.WriteString(err.Error())
		}
	}

	return err
}
