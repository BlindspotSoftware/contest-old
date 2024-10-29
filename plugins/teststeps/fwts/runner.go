package fwts

import (
	"bytes"
	"fmt"
	"io"
	"regexp"
	"strconv"
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
	cmd            = "fwts"
	outputFlag     = "--results-output=/tmp/output"
	outputPath     = "/tmp/output.log"
	jsonOutputPath = "/tmp/output.json"
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

		return events.EmitError(ctx, outputBuf.String(), target, r.ev, err)
	}

	if err := r.ts.runFWTS(ctx, &outputBuf, transportProto); err != nil {
		outputBuf.WriteString(fmt.Sprintf("%v", err))

		return events.EmitError(ctx, outputBuf.String(), target, r.ev, err)
	}

	return events.EmitLog(ctx, outputBuf.String(), target, r.ev)
}

func (ts *TestStep) runFWTS(ctx xcontext.Context, outputBuf *strings.Builder, transport transport.Transport,
) error {
	args := []string{
		cmd,
		strings.Join(ts.Flags, " "),
		outputFlag,
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
		_ = proc.Wait(ctx)
	}

	stdout, stderr := getOutputFromReader(stdoutPipe, stderrPipe, outputBuf)

	if len(string(stdout)) > 0 {
		outputBuf.WriteString(fmt.Sprintf("Stdout:\n%s\n", string(stdout)))
	} else if len(string(stderr)) > 0 {
		outputBuf.WriteString(fmt.Sprintf("Stderr:\n%s\n", string(stderr)))
	}

	if err = ts.parseOutput(ctx, outputBuf, transport, outputPath); err != nil {
		return err
	}

	return nil
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

func (ts *TestStep) parseOutput(ctx xcontext.Context, outputBuf *strings.Builder,
	transport transport.Transport, path string,
) error {
	proc, err := transport.NewProcess(ctx, "cat", []string{path}, "")
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
	_ = proc.Start(ctx)

	stdout, stderr := getOutputFromReader(stdoutPipe, stderrPipe, outputBuf)

	if len(stderr) != 0 {
		return fmt.Errorf("Error retrieving the output. Error: %s", string(stderr))
	}

	outputBuf.WriteString(fmt.Sprintf("\n\nTest logs:\n%s\n", string(stdout)))

	if len(stdout) != 0 {
		re, err := regexp.Compile(`Total:\s+\|\s+\d+\|\s+\d+\|\s+\d+\|\s+\d+\|\s+\d+\|\s+\d+\|`)
		if err != nil {
			return fmt.Errorf("Failed to create the regex: %v", err)
		}

		match := re.FindString(string(stdout))
		if len(match) == 0 {
			outputBuf.WriteString("Failed to parse stdout. Could not find result.\n")
		} else {
			data, err := parseLine(string(match))
			if err != nil {
				return err
			}

			if ts.ReportOnly {
				outputBuf.WriteString(fmt.Sprintf("Test result:\n%s", printData(data)))
				return nil
			}

			if data.Failed > 0 {
				return fmt.Errorf("At least one Test failed. Test result:\n%s", printData(data))
			}

			outputBuf.WriteString(fmt.Sprintf("No Test failed. Test result:\n%s", printData(data)))
		}
	}

	return nil
}

type Data struct {
	Passed   int
	Failed   int
	Aborted  int
	Warnings int
	Skipped  int
	Info     int
}

func parseLine(line string) (Data, error) {
	// Remove leading and trailing spaces
	line = strings.TrimSpace(line)
	line = strings.TrimRight(line, "|")

	// Split the line into individual fields
	fields := strings.Split(line, "|")

	// Create a Data struct
	data := Data{}

	// Extract and parse each field
	for i, field := range fields {
		if i == 0 {
			continue
		}

		// Remove leading and trailing spaces
		field = strings.TrimSpace(field)

		// Parse the field value to an integer
		value, err := strconv.Atoi(field)
		if err != nil {
			return data, fmt.Errorf("failed to parse field %d: %w", i, err)
		}

		// Assign the parsed value to the corresponding field in the struct
		switch i {
		case 1:
			data.Passed = value
		case 2:
			data.Failed = value
		case 3:
			data.Aborted = value
		case 4:
			data.Warnings = value
		case 5:
			data.Skipped = value
		case 6:
			data.Info = value
		}
	}

	return data, nil
}

func printData(data Data) string {
	return fmt.Sprintf("Passed: %d\nFailed: %d\nAborted: %d\nWarnings: %d\nSkipped: %d\nInfo: %d",
		data.Passed, data.Failed, data.Aborted, data.Warnings, data.Skipped, data.Info)
}
