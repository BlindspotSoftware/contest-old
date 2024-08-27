package dutctl

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/linuxboot/contest/pkg/event/testevent"
	"github.com/linuxboot/contest/pkg/events"
	"github.com/linuxboot/contest/pkg/target"
	"github.com/linuxboot/contest/pkg/xcontext"
	"github.com/linuxboot/contest/plugins/teststeps/abstraction/options"
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

	r.ts.writeTestStep(&stdoutMsg, &stderrMsg)

	if r.ts.Host != "" {
		var certFile string

		fp := "/" + filepath.Join("etc", "fti", "keys", "ca-cert.pem")
		if _, err := os.Stat(fp); err == nil {
			certFile = fp
		}

		if !strings.Contains(r.ts.Host, ":") {
			// Add default port
			if certFile == "" {
				r.ts.Host += ":10000"
			} else {
				r.ts.Host += ":10001"
			}
		}
	}

	writeCommand(r.ts.Command, r.ts.Args, &stdoutMsg, &stderrMsg)

	stdoutMsg.WriteString("Stdout:\n")
	stderrMsg.WriteString("Stderr:\n")

	switch r.ts.Command {
	case "power":
		if err := r.powerCmds(ctx, &stdoutMsg, &stderrMsg); err != nil {
			stderrMsg.WriteString(fmt.Sprintf("%v\n", err))

			return events.EmitError(ctx, stderrMsg.String(), target, r.ev, err)
		}
	case "flash":
		if err := r.flashCmds(ctx, &stdoutMsg, &stderrMsg); err != nil {
			stderrMsg.WriteString(fmt.Sprintf("%v\n", err))

			return events.EmitError(ctx, stderrMsg.String(), target, r.ev, err)
		}

	case "serial":
		if err := r.serialCmds(ctx, &stdoutMsg, &stderrMsg); err != nil {
			stderrMsg.WriteString(fmt.Sprintf("%v\n", err))

			return events.EmitError(ctx, stderrMsg.String(), target, r.ev, err)
		}

	default:
		return fmt.Errorf("Command '%s' is not valid. Possible values are 'power', 'flash' and 'serial'.", r.ts.Command)
	}

	if err := events.EmitLog(ctx, stdoutMsg.String(), target, r.ev); err != nil {
		return fmt.Errorf("cannot emit event: %v", err)
	}

	return nil
}
