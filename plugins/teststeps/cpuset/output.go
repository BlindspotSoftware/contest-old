package cpuset

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
		builder.WriteString(fmt.Sprintf("    Command: %s\n", ts.Command))
		builder.WriteString(fmt.Sprintf("    Args: %v\n", ts.Arg))
		builder.WriteString("\n")

		builder.WriteString("  Options:\n")
		builder.WriteString(fmt.Sprintf("    Timeout: %s\n", time.Duration(ts.options.Timeout)))

		builder.WriteString("Default Values:\n")
		builder.WriteString(fmt.Sprintf("  Timeout: %s", defaultTimeout))

		builder.WriteString("\n\n")
	}
}

// Function to format command information and append it to a string builder.
func writeCommand(args string, builders ...*strings.Builder) {
	for _, builder := range builders {
		builder.WriteString("Operation on DUT:\n")
		builder.WriteString(args)
		builder.WriteString("\n\n")

		builder.WriteString("Output:\n")
	}
}
