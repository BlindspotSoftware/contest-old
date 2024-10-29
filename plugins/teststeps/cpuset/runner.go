package cpuset

import (
	"bytes"
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
	core           = "core"
	profile        = "profile"
	privileged     = "sudo"
	cmd            = "cpu"
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
	var stdoutMsg, stderrMsg strings.Builder

	ctx, cancel := options.NewOptions(ctx, defaultTimeout, r.ts.options.Timeout)
	defer cancel()

	pe := test.NewParamExpander(target)

	r.ts.writeTestStep(&stdoutMsg, &stderrMsg)

	transportProto, err := transport.NewTransport(r.ts.transport.Proto, []string{supportedProto}, r.ts.transport.Options, pe)
	if err != nil {
		err := fmt.Errorf("failed to create transport: %w", err)
		stderrMsg.WriteString(fmt.Sprintf("%v", err))

		return events.EmitError(ctx, stderrMsg.String(), target, r.ev, err)
	}

	switch r.ts.Command {
	case core:
		if err := r.ts.coreCmd(ctx, &stdoutMsg, &stderrMsg, transportProto); err != nil {
			stderrMsg.WriteString(fmt.Sprintf("%v\n", err))

			return events.EmitError(ctx, stderrMsg.String(), target, r.ev, err)
		}

	case profile:
		if err := r.ts.profileCmd(ctx, &stdoutMsg, &stderrMsg, transportProto); err != nil {
			stderrMsg.WriteString(fmt.Sprintf("%v\n", err))

			return events.EmitError(ctx, stderrMsg.String(), target, r.ev, err)
		}

	default:
		err := fmt.Errorf("Command '%s' is not valid. Possible values are '%s'.", r.ts.Command, core)
		stderrMsg.WriteString(fmt.Sprintf("%v\n", err))

		return events.EmitError(ctx, stderrMsg.String(), target, r.ev, err)
	}

	if err := events.EmitLog(ctx, stdoutMsg.String(), target, r.ev); err != nil {
		return fmt.Errorf("cannot emit event: %v", err)
	}

	return err
}

// getOutputFromReader reads data from the provided io.Reader instances
// representing stdout and stderr, and returns the collected output as byte slices.
func getOutputFromReader(stderr io.Reader) []byte {
	errBuffer, err := readBuffer(stderr)
	if err != nil {
		fmt.Printf("Failed to read from Stderr buffer: %v\n", err)
	}

	return errBuffer
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
