package qemu

import (
	"fmt"
	"strings"
	"time"
)

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
