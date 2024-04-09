package robot

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/linuxboot/contest/pkg/event/testevent"
	"github.com/linuxboot/contest/pkg/target"
	"github.com/linuxboot/contest/pkg/test"
	"github.com/linuxboot/contest/pkg/xcontext"
	"github.com/linuxboot/contest/plugins/teststeps/abstraction/options"
	"github.com/linuxboot/contest/plugins/teststeps/abstraction/transport"
)

const (
	local = "local"
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

	transportProto, err := transport.NewTransport(r.ts.transport.Proto, []string{local}, r.ts.transport.Options, pe)
	if err != nil {
		err := fmt.Errorf("failed to create transport: %w", err)
		outputBuf.WriteString(fmt.Sprintf("%v", err))

		return emitStderr(ctx, outputBuf.String(), target, r.ev, err)
	}

	if err := r.ts.runRobot(ctx, &outputBuf, transportProto); err != nil {
		return emitStderr(ctx, outputBuf.String(), target, r.ev, err)
	}

	return emitStdout(ctx, outputBuf.String(), target, r.ev)
}

func (ts *TestStep) runRobot(ctx xcontext.Context, outputBuf *strings.Builder, transport transport.Transport,
) error {
	var args []string

	for _, arg := range ts.Args {
		args = append(args, "-v", arg)
	}

	args = append(args, ts.FilePath)

	proc, err := transport.NewProcess(ctx, "/usr/local/bin/robot", args, "")
	if err != nil {
		outputBuf.WriteString(fmt.Sprintf("Failed to create proc: %v", err))

		return fmt.Errorf("Failed to create proc: %w", err)
	}

	writeCommand(proc.String(), outputBuf)

	stdoutPipe, err := proc.StdoutPipe()
	if err != nil {
		outputBuf.WriteString(fmt.Sprintf("Failed to pipe stdout: %v", err))

		return fmt.Errorf("Failed to pipe stdout: %v", err)
	}

	stderrPipe, err := proc.StderrPipe()
	if err != nil {
		outputBuf.WriteString(fmt.Sprintf("Failed to pipe stderr: %v", err))

		return fmt.Errorf("Failed to pipe stderr: %v", err)
	}

	// try to start the process, if that succeeds then the outcome is the result of
	// waiting on the process for its result; this way there's a semantic difference
	// between "an error occured while launching" and "this was the outcome of the execution"
	outcome := proc.Start(ctx)

	stdout, stderr := getOutputFromReader(outputBuf, stdoutPipe, stderrPipe)

	outputBuf.WriteString(fmt.Sprintf("Stdout:\n%s\n", string(stdout)))
	outputBuf.WriteString(fmt.Sprintf("Stderr:\n%s\n", string(stderr)))

	if outcome == nil {
		outcome = proc.Wait(ctx)
	}

	if outcome != nil && !ts.ReportOnly {
		outputBuf.WriteString(fmt.Sprintf("Tests failed: %v", outcome))

		return fmt.Errorf("Tests failed: %v", outcome)
	}

	return nil
}

// getOutputFromReader reads data from the provided io.Reader instances
// representing stdout and stderr, and returns the collected output as byte slices.
func getOutputFromReader(outputBuf *strings.Builder, stdout, stderr io.Reader) ([]byte, []byte) {
	// Read from the stdout and stderr pipe readers``
	outBuffer, err := readBuffer(stdout)
	if err != nil {
		outputBuf.WriteString(fmt.Sprintf("Failed to read from Stdout buffer: %v\n", err))
	}

	errBuffer, err := readBuffer(stderr)
	if err != nil {
		outputBuf.WriteString(fmt.Sprintf("Failed to read from Stderr buffer: %v\n", err))
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
