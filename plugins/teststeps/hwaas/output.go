package hwaas

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
		builder.WriteString(fmt.Sprintf("    Command: %s\n", ts.Command))
		builder.WriteString(fmt.Sprintf("    Arguments: %s\n", ts.Args))
		builder.WriteString(fmt.Sprintf("    Host: %s\n", ts.Host))
		builder.WriteString(fmt.Sprintf("    ContextID: %s\n", ts.ContextID))
		builder.WriteString(fmt.Sprintf("    MachineID: %s\n", ts.MachineID))
		builder.WriteString(fmt.Sprintf("    DeviceID: %s\n", ts.DeviceID))
		builder.WriteString(fmt.Sprintf("    Version: %s\n", ts.Version))
		builder.WriteString(fmt.Sprintf("    Image: %s\n", ts.Image))

		builder.WriteString("  Options:\n")
		builder.WriteString(fmt.Sprintf("    Timeout: %s\n", time.Duration(ts.options.Timeout)))
		builder.WriteString("\n")

		builder.WriteString("Default Values:\n")
		builder.WriteString(fmt.Sprintf("  Timeout: %s\n", defaultTimeout))

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
