package qemu

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/insomniacslk/xjson"
	"github.com/linuxboot/contest/pkg/event"
	"github.com/linuxboot/contest/pkg/event/testevent"
	"github.com/linuxboot/contest/pkg/events"
	"github.com/linuxboot/contest/pkg/test"
	"github.com/linuxboot/contest/pkg/xcontext"
	"github.com/linuxboot/contest/plugins/teststeps"
	"github.com/linuxboot/contest/plugins/teststeps/abstraction/options"
)

// Name of the plugin
var Name = "Qemu"

const (
	parametersKeyword = "parameters"
)

const (
	defaultTimeout = 10 * time.Minute
	defaultNproc   = "3"
	defaultMemory  = "5000"
)

type parameters struct {
	Executable string `json:"executable"`
	Firmware   string `json:"firmware"`
	Nproc      int    `json:"nproc,omitempty"`
	Mem        int    `json:"mem,omitempty"`
	Image      string `json:"image,omitempty"`
	Logfile    string `json:"logfile,omitempty"`
	Steps      []struct {
		Send    string         `json:"send,omitempty"`
		Timeout xjson.Duration `json:"timeout,omitempty"`
		Expect  struct {
			Regex string `json:"regex"`
		}
	} `json:"steps"`
}

// TestStep implementation for this teststep plugin
type TestStep struct {
	parameters
	options options.Parameters
}

// Run executes the step.
func (ts *TestStep) Run(ctx xcontext.Context, ch test.TestStepChannels, params test.TestStepParameters, ev testevent.Emitter, resumeState json.RawMessage) (json.RawMessage, error) {
	tr := NewTargetRunner(ts, ev)
	return teststeps.ForEachTarget(Name, ctx, ch, tr.Run)
}

func (ts *TestStep) populateParams(stepParams test.TestStepParameters) error {
	var parameters, optionsParams *test.Param

	if parameters = stepParams.GetOne(parametersKeyword); parameters.IsEmpty() {
		return fmt.Errorf("parameters cannot be empty")
	}

	if err := json.Unmarshal(parameters.JSON(), &ts.parameters); err != nil {
		return fmt.Errorf("failed to deserialize parameters: %v", err)
	}

	// basic checks whether the executable is usable
	if abs := filepath.IsAbs(ts.Executable); !abs {
		_, err := exec.LookPath(ts.Executable)
		if err != nil {
			return fmt.Errorf("unable to find qemu executable in PATH: %w", err)
		}
	}

	optionsParams = stepParams.GetOne(options.Keyword)

	if !optionsParams.IsEmpty() {
		if err := json.Unmarshal(optionsParams.JSON(), &ts.options); err != nil {
			return fmt.Errorf("failed to deserialize options: %v", err)
		}
	}

	return nil
}

// ValidateParameters validates the parameters associated to the TestStep
func (ts *TestStep) ValidateParameters(ctx xcontext.Context, params test.TestStepParameters) error {
	return ts.populateParams(params)
}

// Needed for the Teststep interface. Returns a Teststep instance.
func New() test.TestStep {
	return &TestStep{}
}

// Needed for the Teststep interface. Returns the Name, the New() Function and
// the events the teststep can emit (which are no events).
func Load() (string, test.TestStepFactory, []event.Name) {
	return Name, New, events.Events
}

// Name returns the name of the Step
func (ts TestStep) Name() string {
	return Name
}
