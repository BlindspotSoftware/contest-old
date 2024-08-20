package qemu

import (
	"fmt"
	"io/fs"
	"os"
	"regexp"
	"strings"
	"time"

	expect "github.com/google/goexpect"
	"github.com/linuxboot/contest/pkg/event/testevent"
	"github.com/linuxboot/contest/pkg/events"
	"github.com/linuxboot/contest/pkg/multiwriter"
	"github.com/linuxboot/contest/pkg/target"
	"github.com/linuxboot/contest/pkg/xcontext"
	"github.com/linuxboot/contest/plugins/teststeps/abstraction/options"
)

type TargetRunner struct {
	ts *TestStep
	ev testevent.Emitter
}

func NewTargetRunner(ts *TestStep, ev testevent.Emitter) *TargetRunner {
	return &TargetRunner{
		ts: ts,
		ev: ev,
	}
}

func (r *TargetRunner) Run(ctx xcontext.Context, target *target.Target) error {
	var outputBuf strings.Builder

	ctx, cancel := options.NewOptions(ctx, defaultTimeout, r.ts.options.Timeout)
	defer cancel()

	r.ts.writeTestStep(&outputBuf)

	if err := r.ts.runQemu(ctx, &outputBuf); err != nil {
		outputBuf.WriteString(fmt.Sprintf("%v\n", err))

		return events.EmitError(ctx, outputBuf.String(), target, r.ev)
	}

	return events.EmitLog(ctx, outputBuf.String(), target, r.ev)
}

func (ts *TestStep) runQemu(ctx xcontext.Context, outputBuf *strings.Builder) error {
	// no graphical output and no network access
	command := []string{ts.Executable, "-nographic", "-nic", "none", "-bios", ts.Firmware}
	qemuOpts := []string{"-m", fmt.Sprintf("%d", ts.Mem), "-smp", fmt.Sprintf("%d", ts.Nproc)}

	command = append(command, qemuOpts...)
	if ts.Image != "" {
		command = append(command, ts.Image)
	}

	var logfile *os.File
	if ts.Logfile != "" {
		logfile, err := os.Create(ts.Logfile)
		if err != nil {
			return fmt.Errorf("Could not create Logfile: %w", err)
		}

		defer logfile.Close()
	} else {
		logfile, err := os.OpenFile("/dev/null", os.O_WRONLY, fs.ModeDevice)
		if err != nil {
			return fmt.Errorf("Could not redirect output to '/dev/null': %w", err)
		}

		defer logfile.Close()
	}

	mw := multiwriter.NewMultiWriter()
	if ctx.Writer() != nil {
		mw.AddWriter(ctx.Writer())
	}
	mw.AddWriter(logfile)

	gExpect, errchan, err := expect.SpawnWithArgs(
		command,
		time.Duration(ts.options.Timeout),
		expect.Tee(mw),
		expect.CheckDuration(time.Minute),
		expect.PartialMatch(false),
		expect.SendTimeout(time.Duration(ts.options.Timeout)),
	)
	if err != nil {
		return fmt.Errorf("Could not start qemu: %w", err)
	}
	defer gExpect.Close()

	outputBuf.WriteString(fmt.Sprintf("Started Qemu with command: %v", command))

	defer func() {
		err := <-errchan
		outputBuf.WriteString(fmt.Sprintf("Error from Qemu: %v", err))
	}()

	// loop over all steps and expect/ send the given strings
	for _, step := range ts.Steps {
		// Expect and Send fields must not both be empty
		if step.Expect.Regex == "" && step.Send == "" {
			return fmt.Errorf("%s is not a valid step statement", step)
		}

		// process expect step
		if step.Expect.Regex != "" {
			if _, _, err := gExpect.Expect(regexp.MustCompile(step.Expect.Regex), time.Duration(step.Timeout)); err != nil {
				return fmt.Errorf("Error while expecting '%s': %w", step.Expect.Regex, err)
			}

			outputBuf.WriteString(fmt.Sprintf("Completed expect step: '%v' with timeout: %v \n", step.Expect.Regex, time.Duration(step.Timeout).String()))
		}

		// process send step
		if step.Send != "" {
			if err := gExpect.Send(step.Send + "\n"); err != nil {
				return fmt.Errorf("Unable to send '%s': %w", step.Send, err)
			}

			// notify the user if the timeout field is used incorrectly
			if step.Expect.Regex == "" && step.Timeout.String() != "" {
				outputBuf.WriteString(fmt.Sprintf("The Timeout %v for send step: %v will be ignored.", step.Timeout, step.Send))
			}

			outputBuf.WriteString(fmt.Sprintf("Completed Send Step: '%v'", step.Send))
		}
	}

	return nil
}
