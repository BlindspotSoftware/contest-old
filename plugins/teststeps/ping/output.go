package ping

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
		builder.WriteString(fmt.Sprintf("    Port: %d\n", ts.Port))
		builder.WriteString("  Expect:\n")
		builder.WriteString(fmt.Sprintf("    ShouldFail: %t\n", ts.Expect.ShouldFail))
		builder.WriteString("  Options:\n")
		builder.WriteString(fmt.Sprintf("    Timeout: %s\n", time.Duration(ts.options.Timeout)))
		builder.WriteString("\n")

		builder.WriteString("Default Values:\n")
		builder.WriteString(fmt.Sprintf("  Port: %d\n", defaultPort))
		builder.WriteString(fmt.Sprintf("  ShouldFail: %t\n", defaultShouldFail))
		builder.WriteString(fmt.Sprintf("  Timeout: %s\n", defaultTimeout))

		builder.WriteString("\n")
	}
}

// Function to format command information and append it to a string builder.
func writeCommand(addr string, builders ...*strings.Builder) {
	for _, builder := range builders {
		builder.WriteString(fmt.Sprintf("Running Ping on %s\n", addr))
		builder.WriteString("\n")
	}
}
