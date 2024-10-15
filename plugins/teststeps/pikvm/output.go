package pikvm

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
		builder.WriteString(fmt.Sprintf("    Image: %s\n", ts.Image))
		builder.WriteString("  Options:\n")
		builder.WriteString(fmt.Sprintf("    Timeout: %s\n", time.Duration(ts.options.Timeout)))
		builder.WriteString("\n")

		builder.WriteString("Default Values:\n")
		builder.WriteString(fmt.Sprintf("  Timeout: %s\n", defaultTimeout))

		builder.WriteString("\n")
	}
}

// Function to format command information and append it to a string builder.
func writeCommand(host, image string, builders ...*strings.Builder) {
	for _, builder := range builders {
		builder.WriteString(fmt.Sprintf("Mounting image '%s' on pikvm '%s'\n", image, host))
		builder.WriteString("\n")
	}
}
