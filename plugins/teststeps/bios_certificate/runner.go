package bios_certificate

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/linuxboot/contest/pkg/event/testevent"
	"github.com/linuxboot/contest/pkg/events"
	"github.com/linuxboot/contest/pkg/target"
	"github.com/linuxboot/contest/pkg/test"
	"github.com/linuxboot/contest/pkg/xcontext"
	"github.com/linuxboot/contest/plugins/teststeps/abstraction/options"
	"github.com/linuxboot/contest/plugins/teststeps/abstraction/transport"
)

const (
	supportedProto = "ssh"
	privileged     = "sudo"
	cmd            = "cert"
	jsonFlag       = "--json"
)

type Error struct {
	Msg string `json:"error"`
}

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

	switch r.ts.Command {
	case "enable":
		if err := r.ts.runEnable(ctx, &outputBuf, transportProto); err != nil {
			outputBuf.WriteString(fmt.Sprintf("%v", err))

			return events.EmitError(ctx, outputBuf.String(), target, r.ev, err)
		}

	case "update":
		if err := r.ts.runUpdate(ctx, &outputBuf, transportProto); err != nil {
			outputBuf.WriteString(fmt.Sprintf("%v", err))

			return events.EmitError(ctx, outputBuf.String(), target, r.ev, err)
		}

	case "disable":
		if err := r.ts.runDisable(ctx, &outputBuf, transportProto); err != nil {
			outputBuf.WriteString(fmt.Sprintf("%v", err))

			return events.EmitError(ctx, outputBuf.String(), target, r.ev, err)
		}

	case "check":
		if err := r.ts.runCheck(ctx, &outputBuf, transportProto); err != nil {
			outputBuf.WriteString(fmt.Sprintf("%v", err))

			return events.EmitError(ctx, outputBuf.String(), target, r.ev, err)
		}

	default:
		err := fmt.Errorf("command not supported")
		outputBuf.WriteString(fmt.Sprintf("%v", err))

		return err
	}

	return events.EmitLog(ctx, outputBuf.String(), target, r.ev)
}

func (ts *TestStep) runEnable(
	ctx xcontext.Context, outputBuf *strings.Builder, transport transport.Transport,
) error {
	if ts.Password == "" || ts.CertPath == "" {
		return fmt.Errorf("password and certificate file must be set")
	}

	args := []string{
		ts.ToolPath,
		cmd,
		"enable",
		fmt.Sprintf("--password=%s", ts.Password),
		fmt.Sprintf("--cert=%s", ts.CertPath),
		jsonFlag,
	}

	proc, err := transport.NewProcess(ctx, privileged, args, "")
	if err != nil {
		return fmt.Errorf("failed to create process: %v", err)
	}

	writeCommand(proc.String(), outputBuf)

	stdoutPipe, err := proc.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to pipe stdout: %v", err)
	}

	stderrPipe, err := proc.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to pipe stderr: %v", err)
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
		outputBuf.WriteString(fmt.Sprintf("Stdout:\n%s\n", string(stdout)))
	} else if len(string(stderr)) > 0 {
		outputBuf.WriteString(fmt.Sprintf("Stderr:\n%s\n", string(stderr)))
	}

	if outcome != nil {
		return fmt.Errorf("failed to run bios certificate cmd: %v", outcome)
	}

	err = parseOutput(stderr)
	if ts.Expect.ShouldFail && err != nil {
		return nil
	}

	return err
}

func (ts *TestStep) runUpdate(
	ctx xcontext.Context, outputBuf *strings.Builder, transport transport.Transport,
) error {
	if ts.CertPath == "" || ts.KeyPath == "" {
		return fmt.Errorf("new certificate and old private key file must be set")
	}

	args := []string{
		ts.ToolPath,
		cmd,
		"update",
		fmt.Sprintf("--private-key=%s", ts.KeyPath),
		fmt.Sprintf("--cert=%s", ts.CertPath),
		jsonFlag,
	}

	proc, err := transport.NewProcess(ctx, privileged, args, "")
	if err != nil {
		return fmt.Errorf("failed to create process: %v", err)
	}

	writeCommand(proc.String(), outputBuf)

	stdoutPipe, err := proc.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to pipe stdout: %v", err)
	}

	stderrPipe, err := proc.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to pipe stderr: %v", err)
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
		outputBuf.WriteString(fmt.Sprintf("Stdout:\n%s\n", string(stdout)))
	} else if len(string(stderr)) > 0 {
		outputBuf.WriteString(fmt.Sprintf("Stderr:\n%s\n", string(stderr)))
	}

	if outcome != nil {
		return fmt.Errorf("failed to run bios certificate cmd: %v", outcome)
	}

	err = parseOutput(stderr)
	if ts.Expect.ShouldFail && err != nil {
		return nil
	}

	return err
}

func (ts *TestStep) runDisable(
	ctx xcontext.Context, outputBuf *strings.Builder, transport transport.Transport,
) error {
	if ts.Password == "" || ts.KeyPath == "" {
		return fmt.Errorf("password and private key file must be set")
	}

	args := []string{
		ts.ToolPath,
		cmd,
		"disable",
		fmt.Sprintf("--password=%s", ts.Password),
		fmt.Sprintf("--private-key=%s", ts.KeyPath),
		jsonFlag,
	}

	proc, err := transport.NewProcess(ctx, privileged, args, "")
	if err != nil {
		return fmt.Errorf("failed to create process: %v", err)
	}

	writeCommand(proc.String(), outputBuf)

	stdoutPipe, err := proc.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to pipe stdout: %v", err)
	}

	stderrPipe, err := proc.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to pipe stderr: %v", err)
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
		outputBuf.WriteString(fmt.Sprintf("Stdout:\n%s\n", string(stdout)))
	} else if len(string(stderr)) > 0 {
		outputBuf.WriteString(fmt.Sprintf("Stderr:\n%s\n", string(stderr)))
	}

	if outcome != nil {
		return fmt.Errorf("failed to run bios certificate cmd: %v", outcome)
	}

	err = parseOutput(stderr)
	if ts.Expect.ShouldFail && err != nil {
		return nil
	}

	return err
}

func (ts *TestStep) runCheck(
	ctx xcontext.Context, outputBuf *strings.Builder, transport transport.Transport,
) error {
	if ts.CertPath == "" {
		return fmt.Errorf("certificate file must be set for the comparison")
	}

	args := []string{
		ts.ToolPath,
		cmd,
		"check",
		fmt.Sprintf("--cert=%s", ts.CertPath),
	}

	proc, err := transport.NewProcess(ctx, privileged, args, "")
	if err != nil {
		return fmt.Errorf("failed to create process: %v", err)
	}

	writeCommand(proc.String(), outputBuf)

	stdoutPipe, err := proc.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to pipe stdout: %v", err)
	}

	stderrPipe, err := proc.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to pipe stderr: %v", err)
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
		outputBuf.WriteString(fmt.Sprintf("Stdout:\n%s\n", string(stdout)))
	} else if len(string(stderr)) > 0 {
		outputBuf.WriteString(fmt.Sprintf("Stderr:\n%s\n", string(stderr)))
	}

	if outcome != nil {
		return fmt.Errorf("failed to run bios certificate cmd: %v", outcome)
	}

	err = parseOutput(stderr)
	if ts.Expect.ShouldFail && err != nil {
		return nil
	}

	return err
}

// getOutputFromReader reads data from the provided io.Reader instances
// representing stdout and stderr, and returns the collected output as byte slices.
func getOutputFromReader(stdout, stderr io.Reader) ([]byte, []byte) {
	// Read from the stdout and stderr pipe readers
	outBuffer, err := readBuffer(stdout)
	if err != nil {
		fmt.Printf("failed to read from Stdout buffer: %v\n", err)
	}

	errBuffer, err := readBuffer(stderr)
	if err != nil {
		fmt.Printf("failed to read from Stderr buffer: %v\n", err)
	}

	return outBuffer, errBuffer
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

func parseOutput(stderr []byte) error {
	err := Error{}
	if len(stderr) != 0 {
		if err := json.Unmarshal(stderr, &err); err != nil {
			return fmt.Errorf("failed to unmarshal stderr '%s': %v", string(stderr), err)
		}
	}

	if err.Msg != "" {
		return errors.New(err.Msg)
	}

	return nil
}
