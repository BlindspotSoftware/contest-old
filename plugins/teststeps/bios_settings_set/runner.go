package bios_settings_set

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
	cmd            = "wmi"
	argument       = "set"
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

		return events.EmitError(ctx, outputBuf.String(), target, r.ev)
	}

	if r.ts.Password == "" && r.ts.KeyPath == "" {
		err := fmt.Errorf("password or certificate file must be set")
		outputBuf.WriteString(fmt.Sprintf("%v", err))

		return events.EmitError(ctx, outputBuf.String(), target, r.ev)
	}

	if len(r.ts.BiosOptions) == 0 {
		err := fmt.Errorf("at least one bios option and value must be set")
		outputBuf.WriteString(fmt.Sprintf("%v", err))

		return events.EmitError(ctx, outputBuf.String(), target, r.ev)
	}

	if err := r.ts.runSet(ctx, &outputBuf, transportProto); err != nil {
		outputBuf.WriteString(fmt.Sprintf("%v", err))

		return events.EmitError(ctx, outputBuf.String(), target, r.ev)
	}

	return events.EmitLog(ctx, outputBuf.String(), target, r.ev)
}

func (ts *TestStep) runSet(
	ctx xcontext.Context, outputBuf *strings.Builder,
	transport transport.Transport,
) error {
	var (
		authString string
		finalErr   error
	)

	if ts.Password != "" {
		authString = fmt.Sprintf("--password=%s", ts.Password)
	} else if ts.KeyPath != "" {
		authString = fmt.Sprintf("--private-key=%s", ts.KeyPath)
	}

	for _, option := range ts.BiosOptions {
		args := []string{
			ts.ToolPath,
			cmd,
			argument,
			fmt.Sprintf("--option=%s", option.Option),
			fmt.Sprintf("--value=%s", option.Value),
			authString,
			jsonFlag,
		}

		proc, err := transport.NewProcess(ctx, privileged, args, "")
		if err != nil {
			err := fmt.Errorf("failed to create process: %v", err)
			outputBuf.WriteString(fmt.Sprintf("%v\n", err))

			return err
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

		stdout, stderr := getOutputFromReader(stdoutPipe, stderrPipe, outputBuf)

		if len(string(stdout)) > 0 {
			outputBuf.WriteString(fmt.Sprintf("Stdout:\n%s\n", string(stdout)))
		} else if len(string(stderr)) > 0 {
			outputBuf.WriteString(fmt.Sprintf("Stderr:\n%s\n", string(stderr)))
		}

		if outcome != nil {
			err := fmt.Errorf("failed to run bios set cmd for option '%s': %v", option.Option, outcome)
			outputBuf.WriteString(fmt.Sprintf("%v\n", err))
			finalErr = err

			continue
		}

		if err := ts.parseOutput(stderr, option.ShouldFail); err != nil {
			outputBuf.WriteString(fmt.Sprintf("%v\n", err))
			outputBuf.WriteString("\n\n")

			finalErr = fmt.Errorf("At least one bios setting could not be set.")

			continue
		}

		outputBuf.WriteString("\n\n")
	}

	return finalErr
}

// getOutputFromReader reads data from the provided io.Reader instances
// representing stdout and stderr, and returns the collected output as byte slices.
func getOutputFromReader(stdout, stderr io.Reader, outputBuf *strings.Builder) ([]byte, []byte) {
	// Read from the stdout and stderr pipe readers
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
	_, err := io.Copy(buf, r)
	if err != nil && err != io.EOF {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (ts *TestStep) parseOutput(stderr []byte, should_fail bool) error {
	err := Error{}
	if len(stderr) != 0 {
		if err := json.Unmarshal(stderr, &err); err != nil {
			return fmt.Errorf("failed to unmarshal stderr: %v", err)
		}
	}

	if err.Msg != "" {
		if err.Msg == "BIOS options are locked, needs unlocking." && should_fail {
			return nil
		} else if err.Msg != "" && should_fail {
			return nil
		} else {
			return errors.New(err.Msg)
		}
	} else if should_fail {
		return fmt.Errorf("Setting BIOS option should fail, but produced no error.")
	}

	return nil
}
