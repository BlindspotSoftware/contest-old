package bios_settings_get

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
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
	argument       = "get"
	jsonFlag       = "--json"
)

type TargetRunner struct {
	ts *TestStep
	ev testevent.Emitter
}

// Output is the data structure for a bios option that is returned
type Output struct {
	Data Data `json:"data"`
}

type Data struct {
	Name           string   `json:"name"`
	Path           string   `json:"path"`
	PossibleValues []string `json:"possible_values"`
	Value          string   `json:"value"`
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

	if err := r.ts.runGet(ctx, &outputBuf, transportProto); err != nil {
		outputBuf.WriteString(fmt.Sprintf("%v", err))

		return events.EmitError(ctx, outputBuf.String(), target, r.ev)
	}

	return events.EmitLog(ctx, outputBuf.String(), target, r.ev)
}

func (ts *TestStep) runGet(
	ctx xcontext.Context, outputBuf *strings.Builder, transport transport.Transport,
) error {
	var (
		finalErr   error
		parsingBuf strings.Builder
	)

	parsingBuf.WriteString("Test result:\n\n")

	for _, expect := range ts.Expect {
		args := []string{
			ts.ToolPath,
			cmd,
			argument,
			fmt.Sprintf("--option=%s", expect.Option),
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
			err := fmt.Errorf("failed to pipe stdout: %v", err)
			outputBuf.WriteString(fmt.Sprintf("%v\n", err))

			return err
		}

		stderrPipe, err := proc.StderrPipe()
		if err != nil {
			err := fmt.Errorf("failed to pipe stderr: %v", err)
			outputBuf.WriteString(fmt.Sprintf("%v\n", err))

			return err
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
			err := fmt.Errorf("failed to run bios get cmd for option '%s': %v", expect.Option, outcome)
			outputBuf.WriteString(fmt.Sprintf("%v\n", err))
			finalErr = err

			continue
		}

		if err = parseOutput(&parsingBuf, stdout, stderr, expect); err != nil {
			parsingBuf.WriteString(fmt.Sprintf("%v\n", err))

			finalErr = fmt.Errorf("At least one expect parameter is not as expected.")
			outputBuf.WriteString("\n")

			continue
		}

		outputBuf.WriteString("\n")
	}

	outputBuf.WriteString(fmt.Sprintf("%s\n", parsingBuf.String()))

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

func parseOutput(parsingBuf *strings.Builder, stdout, stderr []byte, expectOption Expect) error {
	output := Output{}
	if len(stdout) != 0 {
		if err := json.Unmarshal(stdout, &output); err != nil {
			return fmt.Errorf("failed to unmarshal stdout: %v", err)
		}
	}

	exp, err := regexp.Compile(expectOption.Value)
	if err != nil {
		return fmt.Errorf("failed to compile regular expresion from expected value: %v", err)
	}

	errMsg := Error{}

	if len(stderr) != 0 {
		if err := json.Unmarshal(stderr, &errMsg); err != nil {
			return fmt.Errorf("failed to unmarshal stderr: %v", err)
		}
	}

	if errMsg.Msg == "" {
		if expectOption.Option == output.Data.Name {
			if !exp.MatchString(output.Data.Value) {
				return fmt.Errorf("\u2717 BIOS setting '%s' is not as expected, have '%s' want '%s'.",
					expectOption.Option, output.Data.Value, expectOption.Value)
			} else {
				parsingBuf.WriteString(fmt.Sprintf("\u2713 BIOS setting '%s' is set as expected: '%s'.\n", expectOption.Option, expectOption.Value))
				return nil
			}
		}
	} else {
		return fmt.Errorf("\u2717 BIOS setting '%s' was not found in the attribute list: %s", expectOption.Option, errMsg.Msg)
	}

	return nil
}
