package dutctl

import (
	"fmt"
	"strings"
	"time"
)

// Function to format teststep information and append it to a string builder.
func (ts TestStep) writeTestStep(builders ...*strings.Builder) {
	for _, builder := range builders {
		builder.WriteString("Input Parameter:\n")
		builder.WriteString("  Parameter:\n")
		builder.WriteString(fmt.Sprintf("    Host: %s\n", ts.Host))
		builder.WriteString(fmt.Sprintf("    Command: %s\n", ts.Command))
		builder.WriteString(fmt.Sprintf("    Args: %s\n", ts.Args))

		if ts.Command == "serial" {
			builder.WriteString(fmt.Sprintf("    UART: %d\n", ts.UART))
		}

		builder.WriteString(fmt.Sprintf("    Input: %s\n", ts.Input))
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
	}
}

// Function to format command information and append it to a string builder.
func writeCommand(command string, args []string, builders ...*strings.Builder) {
	for _, builder := range builders {
		builder.WriteString("Operation on DUT:\n")
		builder.WriteString(fmt.Sprintf("%s %s", command, strings.Join(args, " ")))
		builder.WriteString("\n\n")
	}
}
