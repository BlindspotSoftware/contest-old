package secureboot

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/linuxboot/contest/pkg/event"
	"github.com/linuxboot/contest/pkg/event/testevent"
	"github.com/linuxboot/contest/pkg/test"
	"github.com/linuxboot/contest/pkg/xcontext"
	"github.com/linuxboot/contest/plugins/teststeps"
	"github.com/linuxboot/contest/plugins/teststeps/abstraction/options"
	"github.com/linuxboot/contest/plugins/teststeps/abstraction/transport"
)

// Name is the name used to look this plugin up.
var Name = "Secure Boot Management"

const (
	parametersKeyword = "parameters"
)

const (
	defaultTimeout = time.Minute
)

type parameters struct {
	Command         string `json:"command"`
	ToolPath        string `json:"tool_path"`
	Hierarchy       string `json:"hierarchy,omitempty"`
	Append          bool   `json:"append,omitempty"`
	KeyFile         string `json:"key_file,omitempty"`
	CertFile        string `json:"cert_file,omitempty"`
	SigningKeyFile  string `json:"signing_key_file,omitempty"`
	SigningCertFile string `json:"signing_cert_file,omitempty"`
	CustomKeyFile   string `json:"custom_key_file,omitempty"`

	Expect struct {
		SecureBoot bool `json:"secure_boot,omitempty"`
		SetupMode  bool `json:"setup_mode,omitempty"`
		ShouldFail bool `json:"should_fail,omitempty"`
	} `json:"expect"`
}

// TestStep implementation for this teststep plugin
type TestStep struct {
	parameters
	transport transport.Parameters
	options   options.Parameters
}

// Run executes the step.
func (ts *TestStep) Run(ctx xcontext.Context, ch test.TestStepChannels, params test.TestStepParameters, ev testevent.Emitter, resumeState json.RawMessage) (json.RawMessage, error) {
	tr := NewTargetRunner(ts, ev)
	return teststeps.ForEachTarget(Name, ctx, ch, tr.Run)
}

func (ts *TestStep) populateParams(stepParams test.TestStepParameters) error {
	var parameters, transportParams, optionsParams *test.Param

	if parameters = stepParams.GetOne(parametersKeyword); parameters.IsEmpty() {
		return fmt.Errorf("parameters cannot be empty")
	}

	if err := json.Unmarshal(parameters.JSON(), &ts.parameters); err != nil {
		return fmt.Errorf("failed to deserialize parameters: %v", err)
	}

	if transportParams = stepParams.GetOne(transport.Keyword); transportParams.IsEmpty() {
		return fmt.Errorf("transport cannot be empty")
	}

	if err := json.Unmarshal(transportParams.JSON(), &ts.transport); err != nil {
		return fmt.Errorf("failed to deserialize transport: %v", err)
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

// New initializes and returns a new exec step.
func New() test.TestStep {
	return &TestStep{}
}

// Load returns the name, factory and events which are needed to register the step.
func Load() (string, test.TestStepFactory, []event.Name) {
	return Name, New, Events
}

// Name returns the name of the Step
func (ts TestStep) Name() string {
	return Name
}
