package cmd

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
		builder.WriteString("  Bin:\n")
		builder.WriteString(fmt.Sprintf("    Executable: %s\n", ts.Executable))
		builder.WriteString(fmt.Sprintf("    Args: %v\n", ts.Args))
		builder.WriteString(fmt.Sprintf("    WorkingDir: %s\n", ts.WorkingDir))
		builder.WriteString(fmt.Sprintf("    ReportOnly: %t\n", ts.ReportOnly))
		builder.WriteString("\n")

		builder.WriteString("  Transport:\n")
		builder.WriteString(fmt.Sprintf("    Protocol: %s\n", ts.transport.Proto))
		builder.WriteString("    Options: \n")
		optionsJSON, err := json.MarshalIndent(ts.transport.Options, "", "    ")
		if err != nil {
			builder.WriteString(fmt.Sprintf("%v\n", ts.transport.Options))
		} else {
			builder.WriteString(string(optionsJSON))
		}
		builder.WriteString("\n")

		builder.WriteString("  Options:\n")
		builder.WriteString(fmt.Sprintf("    Timeout: %s\n", time.Duration(ts.options.Timeout)))
		builder.WriteString("\n")

		builder.WriteString("Expect Parameter:\n")
		for i, expect := range ts.Expect {
			builder.WriteString(fmt.Sprintf("  Expect %d:\n", i+1))
			builder.WriteString(fmt.Sprintf("    Regex: %s\n", expect.Regex))
		}
		builder.WriteString("\n\n")

		builder.WriteString("Default Values:\n")
		builder.WriteString(fmt.Sprintf("  Timeout: %s\n", defaultTimeout))

		builder.WriteString("\n\n")
	}
}

// Function to format command information and append it to a string builder.
func writeCommand(command string, builders ...*strings.Builder) {
	for _, builder := range builders {
		builder.WriteString("Operation on DUT:\n")
		builder.WriteString(command)
		builder.WriteString("\n\n")

		builder.WriteString("Output:\n")
	}
}
