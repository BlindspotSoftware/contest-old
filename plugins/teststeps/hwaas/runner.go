package hwaas

import (
	"fmt"
	"strings"
	"time"

	"github.com/insomniacslk/xjson"
	"github.com/linuxboot/contest/pkg/event/testevent"
	"github.com/linuxboot/contest/pkg/target"
	"github.com/linuxboot/contest/pkg/test"
	"github.com/linuxboot/contest/pkg/xcontext"
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
	var builder strings.Builder

	ctx.Infof("Executing on target %s", target)

	// limit the execution time if specified
	var cancel xcontext.CancelFunc

	if r.ts.Options.Timeout != 0 {
		ctx, cancel = xcontext.WithTimeout(ctx, time.Duration(r.ts.Options.Timeout))
		defer cancel()
	} else {
		r.ts.Options.Timeout = xjson.Duration(defaultTimeout)
		ctx, cancel = xcontext.WithTimeout(ctx, time.Duration(r.ts.Options.Timeout))
		defer cancel()
	}

	pe := test.NewParamExpander(target)

	var params inputStepParams

	if err := pe.ExpandObject(r.ts.inputStepParams, &params); err != nil {
		return err
	}

	writeTestStep(&builder, r.ts)
	writeCommand(&builder, params.Parameter.Command, params.Parameter.Args...)

	switch params.Parameter.Command {
	case "power":
		if err := r.powerCmds(ctx, &builder, target, params.Parameter.Args); err != nil {
			if err := emitEvent(ctx, EventStderr, eventPayload{Msg: err.Error()}, target, r.ev); err != nil {
				return fmt.Errorf("cannot emit event: %v", err)
			}

			return err
		}

	case "flash":
		if err := r.flashCmds(ctx, &builder, target, params.Parameter.Args); err != nil {
			if err := emitEvent(ctx, EventStderr, eventPayload{Msg: err.Error()}, target, r.ev); err != nil {
				return fmt.Errorf("cannot emit event: %v", err)
			}

			return err
		}

	default:
		err := fmt.Errorf("Command %q is not valid. Possible values are 'power' and 'flash'.", params.Parameter.Args)
		if err := emitEvent(ctx, EventStderr, eventPayload{Msg: err.Error()}, target, r.ev); err != nil {
			return fmt.Errorf("cannot emit event: %v", err)
		}

		return err
	}

	if err := emitEvent(ctx, EventStdout, eventPayload{Msg: builder.String()}, target, r.ev); err != nil {
		return fmt.Errorf("Failed to emit event: %v", err)
	}

	return nil
}

// powerCmds is a helper function to call into the different power commands
func (r *TargetRunner) powerCmds(ctx xcontext.Context, builder *strings.Builder, target *target.Target, args []string) error {
	if len(args) >= 1 {

		switch args[0] {

		case "on":
			if err := r.ts.powerOn(ctx, builder, target, r.ev); err != nil {
				return err
			}

			return nil

		case "off":
			if err := r.ts.powerOffSoft(ctx, builder, target, r.ev); err != nil {
				return err
			}

			if len(args) >= 2 {
				if args[1] == "hard" {
					if err := r.ts.powerOffHard(ctx, builder, target, r.ev); err != nil {
						return err
					}
				} else {
					return fmt.Errorf("failed to execute the power off command. The last argument is not valid. The only possible value is 'hard'.")
				}
			}

			return nil

		default:
			return fmt.Errorf("failed to execute the power command. The argument %q is not valid. Possible values are 'on' and 'off'.", args)
		}
	} else {
		return fmt.Errorf("failed to execute the power command. Args is empty. Possible values are 'on' and 'off'.")
	}
}

func (r *TargetRunner) flashCmds(ctx xcontext.Context, builder *strings.Builder, target *target.Target, args []string) error {
	if len(args) >= 2 {

		switch args[0] {

		case "write":
			if err := r.ts.flashWrite(ctx, builder, args[1], target, r.ev); err != nil {
				return err
			}

			return nil

		case "read":
			if err := r.ts.flashRead(ctx, builder, args[1], target, r.ev); err != nil {
				return err
			}

			return nil

		default:
			return fmt.Errorf("Failed to execute the flash command. The argument %q is not valid. Possible values are 'read /path/to/binary' and 'write /path/to/binary'.", args)
		}
	} else {
		return fmt.Errorf("Failed to execute the power command. Args is not valid. Possible values are 'read /path/to/binary' and 'write /path/to/binary'.")
	}
}
