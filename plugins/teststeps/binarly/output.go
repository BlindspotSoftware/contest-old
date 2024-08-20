package binarly

import (
	"fmt"
	"strings"
	"time"
)

// Name returns the name of the Step
func (ts TestStep) Name() string {
	return Name
}

// Function to format teststep information and append it to a string builder.
func (ts TestStep) writeTestStep(builders ...*strings.Builder) {
	for _, builder := range builders {

		builder.WriteString("\n")

		builder.WriteString("  Parameter:\n")
		builder.WriteString("    Token: *hidden*\n")
		builder.WriteString(fmt.Sprintf("    File: %s\n", ts.parameters.File))
		builder.WriteString("\n")

		builder.WriteString("  Options:\n")
		builder.WriteString(fmt.Sprintf("    Timeout: %s\n", time.Duration(ts.options.Timeout)))
		builder.WriteString("\n")

		builder.WriteString("Default Values:\n")
		builder.WriteString(fmt.Sprintf("  Timeout: %s\n", defaultTimeout))
		builder.WriteString("\n\n")
	}
}
