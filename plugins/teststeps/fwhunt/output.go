package fwhunt

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
		builder.WriteString("    RulesDirs:\n")
		for i, rulesdir := range ts.RulesDir {
			builder.WriteString(fmt.Sprintf("      RuleDir %d: %s\n", i+1, rulesdir))
		}
		builder.WriteString("    Rules:\n")
		for i, rules := range ts.Rules {
			builder.WriteString(fmt.Sprintf("      Rule %d: %s\n", i+1, rules))
		}
		builder.WriteString("\n")

		builder.WriteString("  Options:\n")
		builder.WriteString(fmt.Sprintf("    Timeout: %s\n", time.Duration(ts.options.Timeout)))
		builder.WriteString("\n\n")
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
