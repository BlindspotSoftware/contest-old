package chipsec

import (
	"bytes"
	"encoding/json"
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
	supportedProto = "ssh"
	privileged     = "sudo"
	cmd            = "python3"
	bin            = "chipsec_main.py"
	nixOSBin       = "chipsec_main"
	jsonFlag       = "--json"
	outputFile     = "output.json"
)

type Output struct {
	Result string `json:"result"`
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

		return emitStderr(ctx, outputBuf.String(), target, r.ev, err)
	}

	if err := r.ts.runModule(ctx, &outputBuf, transportProto); err != nil {
		outputBuf.WriteString(fmt.Sprintf("%v", err))

		return emitStderr(ctx, outputBuf.String(), target, r.ev, err)
	}

	return emitStdout(ctx, outputBuf.String(), target, r.ev)
}

func (ts *TestStep) runModule(
	ctx xcontext.Context,
	outputBuf *strings.Builder,
	transp transport.Transport,
) error {
	var (
		err      error
		proc     transport.Process
		finalErr error
	)

	for _, module := range ts.Modules {
		outputBuf.WriteString("\n\n\n\n")
		outputBuf.WriteString(fmt.Sprintf("Running tests for chipsec module '%s' now.\n", module))

		var optionalArgs []string

		if ts.Platform != "" {
			optionalArgs = append(optionalArgs, "--platform", ts.Platform)
		}

		if ts.PCH != "" {
			optionalArgs = append(optionalArgs, "--pch", ts.PCH)
		}

		args := []string{
			"-m",
			module,
			jsonFlag,
			outputFile,
		}

		args = append(args, optionalArgs...)

		proc, err = transp.NewProcess(ctx, nixOSBin, args, "")
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
			_ = proc.Wait(ctx)
		}

		stdout, stderr := getOutputFromReader(stdoutPipe, stderrPipe, outputBuf)

		if len(string(stdout)) > 0 {
			outputBuf.WriteString(fmt.Sprintf("Stdout:\n%s\n", string(stdout)))
		} else if len(string(stderr)) > 0 {
			outputBuf.WriteString(fmt.Sprintf("Stderr:\n%s\n", string(stderr)))
		}

		if err := ts.parseOutput(ctx, outputBuf, transp, module); err != nil {
			finalErr = err

			continue
		}
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

func (ts *TestStep) parseOutput(
	ctx xcontext.Context,
	outputBuf *strings.Builder,
	transport transport.Transport,
	module string,
) error {
	args := []string{
		"cat",
		outputFile,
	}

	proc, err := transport.NewProcess(ctx, privileged, args, "")
	if err != nil {
		return fmt.Errorf("Failed to parse Output: %w", err)
	}

	stdoutPipe, err := proc.StdoutPipe()
	if err != nil {
		return fmt.Errorf("Failed to parse Output: %w", err)
	}

	stderrPipe, err := proc.StderrPipe()
	if err != nil {
		return fmt.Errorf("Failed to parse Output: %w", err)
	}

	// try to start the process, if that succeeds then the outcome is the result of
	// waiting on the process for its result; this way there's a semantic difference
	// between "an error occured while launching" and "this was the outcome of the execution"
	outcome := proc.Start(ctx)
	if outcome == nil {
		_ = proc.Wait(ctx)
	}

	stdout, stderr := getOutputFromReader(stdoutPipe, stderrPipe, outputBuf)

	if len(stderr) != 0 {
		return fmt.Errorf("Error retrieving the output. Error: %s", string(stderr))
	}

	data := make(map[string]Output)

	if len(stdout) != 0 {
		if err := json.Unmarshal(stdout, &data); err != nil {
			return fmt.Errorf("Failed to unmarshal stdout: %v", err)
		}
	}

	switch data[fmt.Sprintf("chipsec.modules.%s", module)].Result {
	case "Passed":
		outputBuf.WriteString("ChipSec test passed.")

		return nil

	case "Failed":
		return fmt.Errorf("ChipSec test failed.")

	case "Warning":
		outputBuf.WriteString("ChipSec test resulted in a warning.")

		return nil

	case "NotApplicable":
		outputBuf.WriteString("ChipSec test is not applicable. Module is not supported on this platform.")

		return nil

	case "Information":
		outputBuf.WriteString("ChipSec test only prints out information.")

		return nil

	case "Error":
		return fmt.Errorf("ChipSec test failed while executing.")

	default:
		return fmt.Errorf("Failed to parse chipsec output.\nOutput:\n%s", string(stdout))
	}
}
