package s0ixselftest

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/linuxboot/contest/pkg/event"
	"github.com/linuxboot/contest/pkg/event/testevent"
	"github.com/linuxboot/contest/pkg/events"
	"github.com/linuxboot/contest/pkg/test"
	"github.com/linuxboot/contest/pkg/xcontext"
	"github.com/linuxboot/contest/plugins/teststeps"
	"github.com/linuxboot/contest/plugins/teststeps/abstraction/options"
	"github.com/linuxboot/contest/plugins/teststeps/abstraction/transport"
)

// Name is the name used to look this plugin up.
var Name = "S0ix-Selftest"

// We need a default timeout to avoid endless running tests.
const (
	defaultTimeout = 10 * time.Minute
)

// TestStep implementation for this teststep plugin
type TestStep struct {
	transport transport.Parameters
	options   options.Parameters
}

// Run executes the cmd step.
func (ts *TestStep) Run(ctx xcontext.Context, ch test.TestStepChannels, params test.TestStepParameters, ev testevent.Emitter, resumeState json.RawMessage) (json.RawMessage, error) {
	tr := NewTargetRunner(ts, ev)
	return teststeps.ForEachTarget(Name, ctx, ch, tr.Run)
}

func (ts *TestStep) populateParams(stepParams test.TestStepParameters) error {
	var transportParams, optionsParams *test.Param

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
func (ts *TestStep) ValidateParameters(_ xcontext.Context, params test.TestStepParameters) error {
	return ts.populateParams(params)
}

// New initializes and returns a new HWaaS test step.
func New() test.TestStep {
	return &TestStep{}
}

// Load returns the name, factory and events which are needed to register the step.
func Load() (string, test.TestStepFactory, []event.Name) {
	return Name, New, events.Events
}

// Name returns the plugin name.
func (ts TestStep) Name() string {
	return Name
}
