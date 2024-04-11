package qemu

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/linuxboot/contest/pkg/event"
	"github.com/linuxboot/contest/pkg/event/testevent"
	"github.com/linuxboot/contest/pkg/target"
	"github.com/linuxboot/contest/pkg/xcontext"
)

// events that we may emit during the plugin's lifecycle
const (
	EventStdout = event.Name("Stdout")
	EventStderr = event.Name("Stderr")
)

// Events defines the events that a TestStep is allow to emit. Emitting an event
// that is not registered here will cause the plugin to terminate with an error.
var Events = []event.Name{
	EventStdout,
	EventStderr,
}

type eventPayload struct {
	Msg string
}

func emitEvent(ctx xcontext.Context, name event.Name, payload interface{}, tgt *target.Target, ev testevent.Emitter) error {
	payloadData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("cannot marshal payload for event '%s': %w", name, err)
	}

	msg := json.RawMessage(payloadData)
	data := testevent.Data{
		EventName: name,
		Target:    tgt,
		Payload:   &msg,
	}

	if err := ev.Emit(ctx, data); err != nil {
		return fmt.Errorf("cannot emit event EventCmdStart: %w", err)
	}

	return nil
}

// Function to format teststep information and append it to a string builder.
func (ts TestStep) writeTestStep(builders ...*strings.Builder) {
	for _, builder := range builders {
		builder.WriteString("Input Parameters:\n")
		builder.WriteString(fmt.Sprintf("  Executable: %s\n", ts.Executable))
		builder.WriteString(fmt.Sprintf("  Firmware: %s\n", ts.Firmware))
		builder.WriteString(fmt.Sprintf("  Nproc: %d\n", ts.Nproc))
		builder.WriteString(fmt.Sprintf("  Mem: %d\n", ts.Mem))
		builder.WriteString(fmt.Sprintf("  Image: %s\n", ts.Image))
		builder.WriteString(fmt.Sprintf("  Logfile: %s\n", ts.Logfile))
		builder.WriteString("  Steps:\n")
		for i, step := range ts.Steps {
			builder.WriteString(fmt.Sprintf("  Step %d:\n", i+1))
			builder.WriteString(fmt.Sprintf("    Send: %s\n", step.Send))
			builder.WriteString(fmt.Sprintf("    Timeout: %s\n", step.Timeout))
			builder.WriteString(fmt.Sprintf("    Expect Regex: %s\n", step.Expect.Regex))
		}
		builder.WriteString("\n\n")

		builder.WriteString("  Options:\n")
		builder.WriteString(fmt.Sprintf("    Timeout: %s\n", time.Duration(ts.options.Timeout)))

		builder.WriteString("Default Values:\n")
		builder.WriteString(fmt.Sprintf("  Timeout: %s", defaultTimeout))
		builder.WriteString("\n\n")
	}
}

// emitStderr emits the whole error message to Stderr and returns the error
func emitStderr(ctx xcontext.Context, message string, tgt *target.Target, ev testevent.Emitter, err error) error {
	if err := emitEvent(ctx, EventStderr, eventPayload{Msg: message}, tgt, ev); err != nil {
		return fmt.Errorf("cannot emit event: %v", err)
	}

	return err
}

// emitStdout emits the whole message to Stdout
func emitStdout(ctx xcontext.Context, message string, tgt *target.Target, ev testevent.Emitter) error {
	if err := emitEvent(ctx, EventStdout, eventPayload{Msg: message}, tgt, ev); err != nil {
		return fmt.Errorf("cannot emit event: %v", err)
	}

	return nil
}
