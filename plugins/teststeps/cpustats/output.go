package cpustats

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// Function to format teststep information and append it to a string builder.
func (ts TestStep) writeTestStep(builders ...*strings.Builder) {
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

		builder.WriteString("  Options:\n")
		builder.WriteString(fmt.Sprintf("    Timeout: %s\n", time.Duration(ts.options.Timeout)))

		builder.WriteString("Default Values:\n")
		builder.WriteString(fmt.Sprintf("  Timeout: %s", defaultTimeout))

		builder.WriteString("\n\n")

		builder.WriteString("Expect Parameter:\n")
		builder.WriteString("  General expectations:\n")
		for _, expect := range ts.Expect.General {
			builder.WriteString(fmt.Sprintf("    Option: %s\n", expect.Option))
			builder.WriteString(fmt.Sprintf("    Value: %s\n", expect.Value))
		}

		builder.WriteString("  Core specific expectations:\n")
		for _, expect := range ts.Expect.Individual {
			builder.WriteString(fmt.Sprintf("  CPUs %v:\n", expect.CPUs))
			builder.WriteString(fmt.Sprintf("    Option: %s\n", expect.Option))
			builder.WriteString(fmt.Sprintf("    Value: %s\n", expect.Value))
		}
		builder.WriteString("\n\n")
	}
}

// Function to format command information and append it to a string builder.
func writeCommand(command string, builders ...*strings.Builder) {
	for _, builder := range builders {
		builder.WriteString("Operation on DUT:\n")
		builder.WriteString(command)
		builder.WriteString("\n")
	}
}
