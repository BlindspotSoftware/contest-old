package cpu

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/linuxboot/contest/pkg/event/testevent"
	"github.com/linuxboot/contest/pkg/target"
	"github.com/linuxboot/contest/pkg/test"
	"github.com/linuxboot/contest/pkg/xcontext"
	"github.com/linuxboot/contest/plugins/teststeps/abstraction/transport"
)

const (
	supportedProto             = "ssh"
	certDir                    = "/tmp/cert/"
	dirPerm        os.FileMode = 0o744
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
	ctx.Infof("Executing on target %s", target)

	// limit the execution time if specified
	timeout := r.ts.Options.Timeout
	if timeout != 0 {
		var cancel xcontext.CancelFunc
		ctx, cancel = xcontext.WithTimeout(ctx, time.Duration(timeout))
		defer cancel()
	}

	pe := test.NewParamExpander(target)

	var params inputStepParams
	if err := pe.ExpandObject(r.ts.inputStepParams, &params); err != nil {
		return err
	}

	if params.Transport.Proto != supportedProto {
		return fmt.Errorf("only %q is supported as protocol in this teststep", supportedProto)
	}

	transportProto, err := transport.NewTransport(params.Transport.Proto, params.Transport.Options, pe)
	if err != nil {
		return fmt.Errorf("failed to create transport: %w", err)
	}

	// for any ambiguity, outcome is an error interface, but it encodes whether the process
	// was launched sucessfully and it resulted in a failure; err means the launch failed
	var outcome outcome

	switch r.ts.inputStepParams.Command {
	case "core":
		outcome, err = r.runCoreInfo(ctx, target, transportProto, params)
		if err != nil {
			return err
		}
	case "turbostat":
		outcome, err = r.runCoreInfo(ctx, target, transportProto, params)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("command not supported")
	}

	return outcome
}

func (r *TargetRunner) runCoreInfo(
	ctx xcontext.Context, target *target.Target,
	transport transport.Transport, params inputStepParams,
) (outcome, error) {
	return nil, nil
}

func (r *TargetRunner) runTurboStat(
	ctx xcontext.Context, target *target.Target,
	transport transport.Transport, params inputStepParams,
) (outcome, error) {
	return nil, nil
}

func getOutputFromReader(stdout, stderr io.Reader) (string, string) {
	// Read from the stdout and stderr pipe readers
	outBuffer := make([]byte, 1024)
	_, err := stdout.Read(outBuffer)
	if err != nil {
		fmt.Printf("failed to read from Stdout buffer: %v\n", err)
	}

	errBuffer := make([]byte, 1024)
	_, err = stderr.Read(errBuffer)
	if err != nil {
		fmt.Printf("failed to read from Stderr buffer: %v\n", err)
	}

	return "", ""
}
