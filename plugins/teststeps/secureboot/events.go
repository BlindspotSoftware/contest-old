package secureboot

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
func (ts TestStep) writeEnrollKeysTestStep(builders ...*strings.Builder) {
	for _, builder := range builders {
		builder.WriteString("Input Parameter:\n")
		builder.WriteString("  Transport:\n")
		builder.WriteString(fmt.Sprintf("    Protocol: %s\n", ts.transport.Proto))
		builder.WriteString("    Options: \n")
		optionsJSON, err := json.MarshalIndent(ts.transport.Options, "", "    ")
		if err != nil {
			builder.WriteString(fmt.Sprintf("%v", ts.transport.Options))
		} else {
			builder.WriteString(string(optionsJSON))
		}
		builder.WriteString("\n")

		builder.WriteString("  Parameter:\n")
		builder.WriteString(fmt.Sprintf("    ToolPath: %s\n", ts.ToolPath))
		builder.WriteString(fmt.Sprintf("    Hierarchy: %s\n", ts.Hierarchy))
		builder.WriteString(fmt.Sprintf("    Append: %t\n", ts.Append))
		builder.WriteString(fmt.Sprintf("    KeyFilePath: %s\n", ts.KeyFile))
		builder.WriteString(fmt.Sprintf("    CertFilePath: %s\n", ts.CertFile))
		builder.WriteString(fmt.Sprintf("    SigningKeyFilePath: %s\n", ts.SigningKeyFile))
		builder.WriteString(fmt.Sprintf("    SigningCertFilePath: %s\n", ts.SigningCertFile))
		builder.WriteString("\n")

		builder.WriteString("  Expect:\n")
		builder.WriteString(fmt.Sprintf("    ShouldFail: %t\n", ts.Expect.ShouldFail))
		builder.WriteString("\n")

		builder.WriteString("  Options:\n")
		builder.WriteString(fmt.Sprintf("    Timeout: %s\n", time.Duration(ts.options.Timeout)))
		builder.WriteString("\n")

		builder.WriteString("Default Values:\n")
		builder.WriteString(fmt.Sprintf("  Timeout: %s\n", defaultTimeout))

		builder.WriteString("\n\n")
	}
}

// Function to format teststep information and append it to a string builder.
func (ts TestStep) writeRotateKeysTestStep(builders ...*strings.Builder) {
	for _, builder := range builders {
		builder.WriteString("Input Parameter:\n")
		builder.WriteString("  Transport:\n")
		builder.WriteString(fmt.Sprintf("    Protocol: %s\n", ts.transport.Proto))
		builder.WriteString("    Options: \n")
		optionsJSON, err := json.MarshalIndent(ts.transport.Options, "", "    ")
		if err != nil {
			builder.WriteString(fmt.Sprintf("%v", ts.transport.Options))
		} else {
			builder.WriteString(string(optionsJSON))
		}
		builder.WriteString("\n")

		builder.WriteString("  Parameter:\n")
		builder.WriteString(fmt.Sprintf("    ToolPath: %s\n", ts.ToolPath))
		builder.WriteString(fmt.Sprintf("    Hierarchy: %s\n", ts.Hierarchy))
		builder.WriteString(fmt.Sprintf("    KeyFilePath: %s\n", ts.KeyFile))
		builder.WriteString(fmt.Sprintf("    CertFilePath: %s\n", ts.CertFile))
		builder.WriteString("\n")

		builder.WriteString("  Expect:\n")
		builder.WriteString(fmt.Sprintf("    ShouldFail: %t\n", ts.Expect.ShouldFail))
		builder.WriteString("\n")

		builder.WriteString("  Options:\n")
		builder.WriteString(fmt.Sprintf("    Timeout: %s\n", time.Duration(ts.options.Timeout)))
		builder.WriteString("\n\n")
	}
}

// Function to format teststep information and append it to a string builder.
func (ts TestStep) writeResetTestStep(builders ...*strings.Builder) {
	for _, builder := range builders {
		builder.WriteString("Input Parameter:\n")
		builder.WriteString("  Transport:\n")
		builder.WriteString(fmt.Sprintf("    Protocol: %s\n", ts.transport.Proto))
		builder.WriteString("    Options: \n")
		optionsJSON, err := json.MarshalIndent(ts.transport.Options, "", "    ")
		if err != nil {
			builder.WriteString(fmt.Sprintf("%v", ts.transport.Options))
		} else {
			builder.WriteString(string(optionsJSON))
		}
		builder.WriteString("\n")

		builder.WriteString("  Parameter:\n")
		builder.WriteString(fmt.Sprintf("    ToolPath: %s\n", ts.ToolPath))
		builder.WriteString(fmt.Sprintf("    Hierarchy: %s\n", ts.Hierarchy))
		builder.WriteString(fmt.Sprintf("    SigningKeyFilePath: %s\n", ts.SigningKeyFile))
		builder.WriteString(fmt.Sprintf("    SigningCertFilePath: %s\n", ts.SigningCertFile))
		builder.WriteString(fmt.Sprintf("    CertFilePath: %s\n", ts.CertFile))
		builder.WriteString("\n")

		builder.WriteString("  Expect:\n")
		builder.WriteString(fmt.Sprintf("    ShouldFail: %t\n", ts.Expect.ShouldFail))
		builder.WriteString("\n")

		builder.WriteString("  Options:\n")
		builder.WriteString(fmt.Sprintf("    Timeout: %s\n", time.Duration(ts.options.Timeout)))
		builder.WriteString("\n\n")
	}
}

// Function to format teststep information and append it to a string builder.
func (ts TestStep) writeCustomKeyTestStep(builders ...*strings.Builder) {
	for _, builder := range builders {
		builder.WriteString("Input Parameter:\n")
		builder.WriteString("  Transport:\n")
		builder.WriteString(fmt.Sprintf("    Protocol: %s\n", ts.transport.Proto))
		builder.WriteString("    Options: \n")
		optionsJSON, err := json.MarshalIndent(ts.transport.Options, "", "    ")
		if err != nil {
			builder.WriteString(fmt.Sprintf("%v", ts.transport.Options))
		} else {
			builder.WriteString(string(optionsJSON))
		}
		builder.WriteString("\n")

		builder.WriteString("  Parameter:\n")
		builder.WriteString(fmt.Sprintf("    ToolPath: %s\n", ts.ToolPath))
		builder.WriteString(fmt.Sprintf("    Hierarchy: %s\n", ts.Hierarchy))
		builder.WriteString(fmt.Sprintf("    CustomKeyFilePath: %s\n", ts.CustomKeyFile))
		builder.WriteString("\n")

		builder.WriteString("  Expect:\n")
		builder.WriteString(fmt.Sprintf("    ShouldFail: %t\n", ts.Expect.ShouldFail))
		builder.WriteString("\n")

		builder.WriteString("  Options:\n")
		builder.WriteString(fmt.Sprintf("    Timeout: %s\n", time.Duration(ts.options.Timeout)))
		builder.WriteString("\n\n")
	}
}

// Function to format teststep information and append it to a string builder.
func (ts TestStep) writeStatusTestStep(builders ...*strings.Builder) {
	for _, builder := range builders {
		builder.WriteString("Input Parameter:\n")
		builder.WriteString("  Transport:\n")
		builder.WriteString(fmt.Sprintf("    Protocol: %s\n", ts.transport.Proto))
		builder.WriteString("    Options: \n")
		optionsJSON, err := json.MarshalIndent(ts.transport.Options, "", "    ")
		if err != nil {
			builder.WriteString(fmt.Sprintf("%v", ts.transport.Options))
		} else {
			builder.WriteString(string(optionsJSON))
		}
		builder.WriteString("\n")

		builder.WriteString("  Parameter:\n")
		builder.WriteString(fmt.Sprintf("    ToolPath: %s\n", ts.ToolPath))
		builder.WriteString("\n")

		builder.WriteString("  Expect:\n")
		builder.WriteString(fmt.Sprintf("      Secure Boot: %t\n", ts.Expect.SecureBoot))
		builder.WriteString(fmt.Sprintf("      Setup Mode: %t\n", ts.Expect.SetupMode))
		builder.WriteString("\n")

		builder.WriteString("  Options:\n")
		builder.WriteString(fmt.Sprintf("    Timeout: %s\n", time.Duration(ts.options.Timeout)))
		builder.WriteString("\n\n")
	}
}

// Function to format command information and append it to a string builder.
func writeCommand(privileged bool, command string, args []string, builders ...*strings.Builder) {
	for _, builder := range builders {
		builder.WriteString("Executing Command:\n")
		switch privileged {
		case false:
			builder.WriteString(fmt.Sprintf("%s %s", command, strings.Join(args, " ")))
		case true:
			builder.WriteString(fmt.Sprintf("sudo %s %s", command, strings.Join(args, " ")))

		}
		builder.WriteString("\n\n")
	}
}

// emitStderr emits the whole error message an returns the error
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
