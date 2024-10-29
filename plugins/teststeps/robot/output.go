package robot

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
		builder.WriteString(fmt.Sprintf("    FilePath: %s\n", ts.FilePath))
		builder.WriteString(fmt.Sprintf("    Args: %v\n", ts.Args))
		builder.WriteString(fmt.Sprintf("    ReportOnly: %v\n", ts.ReportOnly))

		builder.WriteString("\n")

		builder.WriteString("  Options:\n")
		builder.WriteString(fmt.Sprintf("    Timeout: %s\n", time.Duration(ts.options.Timeout)))

		builder.WriteString("Default Values:\n")
		builder.WriteString(fmt.Sprintf("  Timeout: %s\n", defaultTimeout))

		builder.WriteString("Executing Command:\n")

		cmd := "robot"
		for _, arg := range ts.Args {
			cmd += fmt.Sprintf(" -v %s", arg)
		}

		builder.WriteString(fmt.Sprintf("%s %s", cmd, ts.FilePath))

		builder.WriteString("\n\n")
	}
}

// Function to format command information and append it to a string builder.
func writeCommand(args string, builders ...*strings.Builder) {
	for _, builder := range builders {
		builder.WriteString("Operation:\n")
		builder.WriteString(args)
		builder.WriteString("\n\n")

		builder.WriteString("Output:\n")
	}
}
