package cpuset

import (
	"fmt"
	"strings"

	"github.com/linuxboot/contest/pkg/xcontext"
	"github.com/linuxboot/contest/plugins/teststeps/abstraction/transport"
)

// profileCmds is a helper function to call into the different core commands
func (ts *TestStep) profileCmd(ctx xcontext.Context, stdoutMsg, stderrMsg *strings.Builder, transport transport.Transport) error {
	args := []string{
		ts.ToolPath,
		cmd,
		"set-profile",
		fmt.Sprintf("--profile=%s", ts.Arg),
		jsonFlag,
	}

	proc, err := transport.NewProcess(ctx, privileged, args, "")
	if err != nil {
		return fmt.Errorf("Failed to create proc: %w", err)
	}

	writeCommand(proc.String(), stdoutMsg, stderrMsg)

	stderrPipe, err := proc.StderrPipe()
	if err != nil {
		return fmt.Errorf("Failed to pipe stderr: %v", err)
	}

	// try to start the process, if that succeeds then the outcome is the result of
	// waiting on the process for its result; this way there's a semantic difference
	// between "an error occured while launching" and "this was the outcome of the execution"
	outcome := proc.Start(ctx)
	if outcome == nil {
		outcome = proc.Wait(ctx)
	}

	stderr := getOutputFromReader(stderrPipe)

	if outcome != nil {
		stderrMsg.WriteString(fmt.Sprintf("Stderr:\n%s\n", string(stderr)))
		return fmt.Errorf("Failed to set acpi platform profile to '%s': %v.", ts.Arg, outcome)
	}

	stderrMsg.WriteString(fmt.Sprintf("Stderr:\n%s\n", string(stderr)))

	return nil
}
