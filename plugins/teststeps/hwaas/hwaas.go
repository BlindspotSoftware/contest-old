package hwaas

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/linuxboot/contest/pkg/event"
	"github.com/linuxboot/contest/pkg/event/testevent"
	"github.com/linuxboot/contest/pkg/target"
	"github.com/linuxboot/contest/pkg/test"
	"github.com/linuxboot/contest/pkg/xcontext"
	"github.com/linuxboot/contest/plugins/teststeps"
)

// This data has to be filled with information to write it into our Context writer.
type Data struct {
	Index     int
	Name      string
	Label     string
	Args      []string
	Logs      []byte
	Error     error
	StartTime time.Time
	EndTime   time.Time
}

// http response structs
// this struct is the response for GET /flash
type getFlash struct {
	State string `json:"state"` // possible values: "ready", "busy" or ready
	Error string `json:"error"`
}

type postFlash struct {
	Action string `json:"action"` // possible values: "read" or "write"
}

// this struct is the response for GET /flash/file
type getFlashFile struct {
	Output []byte `json:"output"`
}

// this struct can be used for GET /vcc /led /reset
type getState struct {
	State string `json:"state"` // possible values: "on" or "off"
}

// Name is the name used to look this plugin up.
var Name = "HWaaS"

// We need a default timeout to avoid endless running tests.
const defaultTimeoutParameter = "15m"

// HWaaS is used to run arbitrary commands as test steps.
type HWaaS struct {
	hostname  *test.Param
	contextID *test.Param
	machineID *test.Param
	command   *test.Param  // Command that shall be run on the dut.
	args      []test.Param // Arguments that the command need.
}

// Name returns the plugin name.
func (hws HWaaS) Name() string {
	return Name
}

// Run executes the cmd step.
func (hws *HWaaS) Run(ctx xcontext.Context, ch test.TestStepChannels, params test.TestStepParameters, ev testevent.Emitter, resumeState json.RawMessage) (json.RawMessage, error) {
	log := ctx.Logger()
	// Validate the parameter
	if err := hws.validateAndPopulate(params); err != nil {
		return nil, err
	}

	output := Data{
		Name:      "name",
		Label:     "label",
		StartTime: time.Now(),
	}

	f := func(ctx xcontext.Context, target *target.Target) error {
		returnFunc := func(err error) {
			output.Error = err

			data, err := json.Marshal(output)
			if err != nil {
				log.Warnf("marshaling output data failed: %w", err)
			}

			if ctx.Writer() != nil {
				w := ctx.Writer()
				_, err := w.Write(data)
				if err != nil {
					log.Warnf("writing to ctx.Writer failed: %w", err)
				}
			}

			return
		}

		// expand all variables
		hostname, err := hws.hostname.Expand(target)
		if err != nil {
			returnFunc(fmt.Errorf("failed to expand variable 'hostname': %v", err))

			return output.Error
		}
		if hostname == "" {
			returnFunc(fmt.Errorf("variable 'hostname' must not be empty: %v", err))

			return output.Error
		}

		contextID, err := hws.contextID.Expand(target)
		if err != nil {
			returnFunc(fmt.Errorf("failed to expand variable 'contextID': %v", err))

			return output.Error
		}
		if contextID == "" {
			return fmt.Errorf("variable 'contextID' must not be empty: %v", err)
		}

		machineID, err := hws.machineID.Expand(target)
		if err != nil {
			returnFunc(fmt.Errorf("failed to expand variable 'machineID': %v", err))

			return output.Error
		}
		if machineID == "" {
			returnFunc(fmt.Errorf("variable 'machineID' must not be empty: %v", err))

			return output.Error
		}

		command, err := hws.command.Expand(target)
		if err != nil {
			returnFunc(fmt.Errorf("failed to expand variable 'command': %v", err))

			return output.Error
		}

		var args []string
		for _, arg := range hws.args {
			expArg, err := arg.Expand(target)
			if err != nil {
				returnFunc(fmt.Errorf("failed to expand argument '%s': %v", arg, err))

				return output.Error
			}
			args = append(args, expArg)
		}

		output.Args = args

		switch command {
		case "power":
			if len(args) >= 1 {
				switch args[0] {
				case "on":
					endpoint := fmt.Sprintf("%s/contexts/%s/machines/%s/power", hostname, contextID, machineID)

					resp, err := HTTPRequest(ctx, http.MethodPut, endpoint, bytes.NewBuffer(nil))
					if err != nil {
						returnFunc(fmt.Errorf("failed to do http request"))

						return output.Error
					}

					if resp.StatusCode == 200 {
						log.Infof("dut powered on.")

						return nil
					}

					returnFunc(fmt.Errorf("device could not be turned on"))

					return output.Error

				case "off":
					endpoint := fmt.Sprintf("%s/contexts/%s/machines/%s/power", hostname, contextID, machineID)

					resp, err := HTTPRequest(ctx, http.MethodDelete, endpoint, bytes.NewBuffer(nil))
					if err != nil {
						returnFunc(fmt.Errorf("failed to do http request"))

						return output.Error
					}

					if resp.StatusCode == 200 {
						log.Infof("dut powered off.")

						return nil
					}

					returnFunc(fmt.Errorf("device could not be turned off"))

					return output.Error

				default:
					returnFunc(fmt.Errorf("failed to execute the power command. The argument %q is not valid. Possible values are 'on' and 'off'.", args))

					return output.Error
				}

			} else {
				returnFunc(fmt.Errorf("failed to execute the power command. Args is empty. Possible values are 'on' and 'off'."))

				return output.Error
			}

		case "flash":
			if len(args) >= 2 {
				switch args[0] {
				case "write":
					if args[1] == "" {
						returnFunc(fmt.Errorf("no file was set to read or write: %v\n", err))

						return output.Error
					}

					endpoint := fmt.Sprintf("%s/contexts/%s/machines/%s/flash", hostname, contextID, machineID)

					if isBusy := isTargetBusy(ctx, endpoint); isBusy {
						returnFunc(fmt.Errorf("target is currently busy"))

						return output.Error
					}

					err = flashTarget(ctx, endpoint, args[1])
					if err != nil {
						returnFunc(fmt.Errorf("flashing %s failed: %v\n", args[1], err))

						return output.Error
					}

					log.Infof("successfully flashed binary")

				default:
					returnFunc(fmt.Errorf("Failed to execute the flash command. The argument %q is not valid. Possible values are 'read /path/to/binary' and 'write /path/to/binary'.", args))

					return output.Error
				}

			} else {
				returnFunc(fmt.Errorf("Failed to execute the power command. Args is not valid. Possible values are 'read /path/to/binary' and 'write /path/to/binary'."))

				return output.Error
			}

		default:
			returnFunc(fmt.Errorf("Command %q is not valid. Possible values are 'power' and 'flash'.", args))

			return output.Error
		}

		output.EndTime = time.Now()

		data, err := json.Marshal(output)
		if err != nil {
			log.Warnf("marshaling output data failed: %w", err)
		}

		if ctx.Writer() != nil {
			w := ctx.Writer()
			_, err := w.Write(data)
			if err != nil {
				log.Warnf("Writing to ctx.Writer failed: %w", err)
			}
		}

		return nil
	}

	return teststeps.ForEachTarget(Name, ctx, ch, f)
}

func (hws *HWaaS) validateAndPopulate(params test.TestStepParameters) error {
	// validate the hwaas hostname
	hws.hostname = params.GetOne("hostname")
	if hws.hostname.IsEmpty() {
		return errors.New("invalid or missing 'hostname' parameter, must be exactly one string")
	}

	// validate the hwaas context ID
	hws.contextID = params.GetOne("contextID")
	if hws.contextID.IsEmpty() {
		return errors.New("invalid or missing 'contextID' parameter, must be exactly one string")
	}

	// validate the hwaas machine ID
	hws.machineID = params.GetOne("machineID")
	if hws.machineID.IsEmpty() {
		return errors.New("invalid or missing 'machineID' parameter, must be exactly one string")
	}

	// validate the hwaas command
	hws.command = params.GetOne("command")
	if hws.command.IsEmpty() {
		return fmt.Errorf("missing or empty 'command' parameter")
	}

	// validate the hwaas command args
	hws.args = params.Get("args")

	return nil
}

// ValidateParameters validates the parameters associated to the TestStep
func (ts *HWaaS) ValidateParameters(_ xcontext.Context, params test.TestStepParameters) error {
	return ts.validateAndPopulate(params)
}

// New initializes and returns a new HWaaS test step.
func New() test.TestStep {
	return &HWaaS{}
}

// Load returns the name, factory and events which are needed to register the step.
func Load() (string, test.TestStepFactory, []event.Name) {
	return Name, New, nil
}

func HTTPRequest(ctx context.Context, method string, endpoint string, body io.Reader) (*http.Response, error) {
	client := &http.Client{}

	req, err := http.NewRequestWithContext(ctx, method, endpoint, body)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return resp, nil
}

func isTargetBusy(ctx xcontext.Context, endpoint string) bool {
	log := ctx.Logger()

	resp, err := HTTPRequest(ctx, http.MethodGet, endpoint, bytes.NewBuffer(nil))
	if err != nil {
		log.Warnf("failed to do http request")
	}

	data := getFlash{}

	json.NewDecoder(resp.Body).Decode(&data)

	if data.State == "busy" {
		return true
	}

	return false
}

func flashTarget(ctx context.Context, endpoint string, filePath string) error {
	file, _ := os.Open(filePath)
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	form, _ := writer.CreateFormFile("file", filepath.Base(filePath))
	io.Copy(form, file)
	writer.Close()

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, fmt.Sprintf("%s%s", endpoint, "/file"), body)
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", writer.FormDataContentType())

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("failed to upload binary")
	}

	postFlash := postFlash{
		Action: "write",
	}

	flashBody, err := json.Marshal(postFlash)
	if err != nil {
		return fmt.Errorf("failed to marshal body: %w", err)
	}

	resp, err = HTTPRequest(ctx, http.MethodPost, endpoint, bytes.NewBuffer(flashBody))
	if err != nil {
		return fmt.Errorf("failed to do http request")
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("failed to flash binary on target")
	}

	return nil
}
