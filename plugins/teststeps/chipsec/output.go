package chipsec

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// Function to format teststep information and append it to a string builder.
func (ts TestStep) writeTestStep(builders ...*strings.Builder) {
	for _, builder := range builders {
		platform := ts.Platform
		pch := ts.PCH

		if platform == "" {
			platform = "auto"
		}

		if pch == "" {
			pch = "auto"
		}

		builder.WriteString("Input Parameter:\n")
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

		builder.WriteString("  Parameter:\n")
		builder.WriteString(fmt.Sprintf("    Platform: %s\n", platform))
		builder.WriteString(fmt.Sprintf("    PCH: %s\n", pch))

		builder.WriteString("    Modules:\n")
		for i, module := range ts.Modules {
			builder.WriteString(fmt.Sprintf("      Module %d: %s\n", i+1, module))
		}
		builder.WriteString("\n")

		builder.WriteString("  Options:\n")
		builder.WriteString(fmt.Sprintf("    Timeout: %s\n", time.Duration(ts.options.Timeout)))
		builder.WriteString("\n")

		builder.WriteString("Default Values:\n")
		builder.WriteString(fmt.Sprintf("  Timeout: %s", defaultTimeout))
		builder.WriteString("\n")
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
