package hwaas

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
)

// We need a default timeout to avoid endless running tests.
const (
	defaultTimeout    time.Duration = 5 * time.Minute
	defaultContextID  string        = "0fb4acd8-e429-11ed-b5ea-0242ac120002"
	defaultMachineID  string        = "ws"
	defaultDeviceID   string        = "flasher"
	defaultHost       string        = "http://9e-hwaas-aux1.lab.9e.network"
	parametersKeyword               = "parameters"
)

type parameters struct {
	Command   string   `json:"command,omitempty"`
	Args      []string `json:"args,omitempty"`
	Host      string   `json:"host,omitempty"`
	Version   string   `json:"version,omitempty"`
	ContextID string   `json:"context_id,omitempty"`
	MachineID string   `json:"machine_id,omitempty"`
	DeviceID  string   `json:"device_id,omitempty"`
	Image     string   `json:"image,omitempty"`
}

// Name is the name used to look this plugin up.
const Name = "HwaaS"

// TestStep implementation for this teststep plugin
type TestStep struct {
	parameters
	options options.Parameters
}

// Run executes the cmd step.
func (ts *TestStep) Run(ctx xcontext.Context, ch test.TestStepChannels, params test.TestStepParameters, ev testevent.Emitter, resumeState json.RawMessage) (json.RawMessage, error) {
	tr := NewTargetRunner(ts, ev)
	return teststeps.ForEachTarget(Name, ctx, ch, tr.Run)
}

func (ts *TestStep) validateAndPopulate(stepParams test.TestStepParameters) error {
	var parameters, optionsParams *test.Param

	if parameters = stepParams.GetOne(parametersKeyword); parameters.IsEmpty() {
		return fmt.Errorf("parameters cannot be empty")
	}

	if err := json.Unmarshal(parameters.JSON(), &ts.parameters); err != nil {
		return fmt.Errorf("failed to deserialize parameters: %v", err)
	}

	optionsParams = stepParams.GetOne(options.Keyword)

	if err := json.Unmarshal(optionsParams.JSON(), &ts.options); err != nil {
		return fmt.Errorf("failed to deserialize options: %v", err)
	}

	if ts.Host == "" {
		ts.Host = defaultHost
	}

	if ts.ContextID == "" {
		ts.ContextID = defaultContextID
	}

	if ts.MachineID == "" {
		ts.MachineID = defaultMachineID
	}

	if ts.DeviceID == "" {
		ts.DeviceID = defaultDeviceID
	}

	if ts.Command == "" {
		return fmt.Errorf("missing or empty 'command' parameter")
	}

	if len(ts.Args) == 0 {
		return fmt.Errorf("missing or empty 'args' parameter")
	}

	return nil
}

// ValidateParameters validates the parameters associated to the TestStep
func (ts *TestStep) ValidateParameters(_ xcontext.Context, params test.TestStepParameters) error {
	return ts.validateAndPopulate(params)
}

// New initializes and returns a new HWaaS test step.
func New() test.TestStep {
	return &TestStep{}
}

// Load returns the name, factory and events which are needed to register the step.
func Load() (string, test.TestStepFactory, []event.Name) {
	return Name, New, Events
}

// Name returns the plugin name.
func (ts TestStep) Name() string {
	return Name
}
