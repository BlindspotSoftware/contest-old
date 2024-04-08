package copy

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/linuxboot/contest/pkg/event/testevent"
	"github.com/linuxboot/contest/pkg/target"
	"github.com/linuxboot/contest/pkg/test"
	"github.com/linuxboot/contest/pkg/xcontext"
	"github.com/linuxboot/contest/plugins/teststeps/abstraction/options"
	"github.com/linuxboot/contest/plugins/teststeps/abstraction/transport"
	sshTransport "github.com/linuxboot/contest/plugins/teststeps/abstraction/transport"
)

const (
	supportedProto = "ssh"
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

		return emitStderr(ctx, outputBuf.String(), target, r.ev, err)
	}

	if err := r.runCopy(ctx, &outputBuf, target, transportProto); err != nil {
		outputBuf.WriteString(fmt.Sprintf("%v\n", err))

		return emitStderr(ctx, outputBuf.String(), target, r.ev, err)
	}

	return emitStdout(ctx, outputBuf.String(), target, r.ev)
}

func (r *TargetRunner) runCopy(ctx xcontext.Context, outputBuf *strings.Builder, target *target.Target,
	transport transport.Transport,
) error {
	copy, err := transport.NewCopy(ctx, r.ts.SrcPath, r.ts.DstPath, r.ts.Recursive)
	if err != nil {
		return fmt.Errorf("Failed to copy data to target: %v", err)
	}

	if r.ts.transport.Proto == "ssh" {
		config := sshTransport.DefaultSSHTransportConfig()
		if err := json.Unmarshal(r.ts.transport.Options, &config); err != nil {
			return fmt.Errorf("Failed to unmarshal Transport options: %v", err)
		}

		writeCommand(copy.String(), fmt.Sprintf("%s:%d", config.Host, config.Port), outputBuf)
	} else {
		writeCommand(copy.String(), "localhost", outputBuf)
	}

	if err := copy.Copy(ctx); err != nil {
		return fmt.Errorf("Failed to copy data: %v", err)
	}

	writeCommandOutput(outputBuf, fmt.Sprintf("Successfully copied %s to %s", r.ts.SrcPath, r.ts.DstPath))

	return nil
}
