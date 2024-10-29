package pikvm

import (
	"errors"
	"fmt"
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

	r.ts.writeTestStep(&outputBuf)

	r.ts.Host = fmt.Sprintf("%s/api/msd", r.ts.Host)

	// for any ambiguity, outcome is an error interface, but it encodes whether the process
	// was launched sucessfully and it resulted in a failure; err means the launch failed
	if err := r.ts.runPiKVM(ctx, &outputBuf); err != nil {
		return events.EmitError(ctx, outputBuf.String(), target, r.ev, err)
	}

	return events.EmitLog(ctx, outputBuf.String(), target, r.ev)
}

func (ts *TestStep) runPiKVM(ctx xcontext.Context, outputBuf *strings.Builder) error {
	writeCommand(ts.Host, ts.Image, outputBuf)

	hashSum, err := calcSHA256(ts.Image)
	if err != nil {
		return err
	}

	status, err := ts.getUsbPlugStatus(ctx)
	if err != nil {
		return fmt.Errorf("failed to check usb plug state: %v", err)
	}

	if status.Drive.Connected {
		outputBuf.WriteString("Unplug USB port.\n")

		if err := ts.plugUSB(ctx, unplug); err != nil {
			return fmt.Errorf("failed to unplug the usb device: %v", err)
		}
	}

	if err := ts.checkMountedImages(ctx, hashSum); errors.Is(err, ErrMissingImage) {
		outputBuf.WriteString("Post image to pikvm.\n")

		if err := ts.postMountImage(ctx); err != nil {
			return fmt.Errorf("failed to post image to api: %v", err)
		}
	} else if err != nil {
		return fmt.Errorf("failed to check mounted images: %v", err)
	} else {
		outputBuf.WriteString("Image already exists.\n")
	}

	outputBuf.WriteString("Configure image.\n")

	if err := ts.configureUSB(ctx, hashSum); err != nil {
		return fmt.Errorf("failed to configure usb device: %v", err)
	}

	outputBuf.WriteString("Plug USB port.\n")

	if err := ts.plugUSB(ctx, plug); err != nil {
		return fmt.Errorf("failed to plug the usb device: %v", err)
	}

	outputBuf.WriteString("Successfully mounted image.\n")

	return nil
}
