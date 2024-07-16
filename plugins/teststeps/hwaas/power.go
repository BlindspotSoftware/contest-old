package hwaas

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/linuxboot/contest/pkg/xcontext"
)

const (
	on     = "on"
	off    = "off"
	reset  = "reset"
	vcc    = "vcc"
	trials = 5
)

// powerCmds is a helper function to call into the different power commands
func (ts *TestStep) powerCmds(ctx xcontext.Context, outputBuf *strings.Builder) error {
	if len(ts.Args) >= 1 {
		switch ts.Args[0] {

		case "on":
			if err := ts.powerOn(ctx, outputBuf); err != nil {
				return err
			}

			return nil

		case "off":
			if err := ts.powerOffSoft(ctx, outputBuf); err != nil {
				outputBuf.WriteString(fmt.Sprintf("Failed to power off the device: %s. Trying to power off the device hard now.\n", err))
				if err := ts.powerOffHard(ctx, outputBuf); err != nil {
					return err
				}
			}

			if len(ts.Args) >= 2 {
				if ts.Args[1] != "hard" {
					outputBuf.WriteString(fmt.Sprintf("Failed to execute the reboot command with arguments: %v. The last argument is not valid.\nThe only possible value is 'hard'. Executing a hard reset instead now.", ts.Args))
				}
				if err := ts.powerOffHard(ctx, outputBuf); err != nil {
					return err
				}
			}

			return nil

		case "reboot":
			if len(ts.Args) >= 2 {
				if ts.Args[1] != "hard" {
					outputBuf.WriteString(fmt.Sprintf("Failed to execute the reboot command with arguments: %v. The last argument is not valid.\nThe only possible value is 'hard'. Executing a hard reset instead now.", ts.Args))
				}

				if err := ts.powerOffHard(ctx, outputBuf); err != nil {
					return err
				}
			} else {
				if err := ts.powerOffSoft(ctx, outputBuf); err != nil {
					outputBuf.WriteString(fmt.Sprintf("Failed to power off the device: %s. Trying to power off the device hard now.\n", err))
					if err := ts.powerOffHard(ctx, outputBuf); err != nil {
						return err
					}
				}
			}

			if err := ts.powerOn(ctx, outputBuf); err != nil {
				return err
			}

			return nil

		default:
			return fmt.Errorf("failed to execute the power command. The argument '%s' is not valid. Possible values are 'on', 'off' and 'reboot'.", ts.Args)
		}
	} else {
		return fmt.Errorf("failed to execute the power command. Arguments are empty. Possible values are 'on', 'off' and 'reboot'.")
	}
}

// powerOn turns on the device. To power the device on we have to fulfill this requirements -> reset is off -> pdu is on.
func (ts *TestStep) powerOn(ctx xcontext.Context, outputBuf *strings.Builder) error {
	if err := ts.unresetDUT(ctx); err != nil {
		return fmt.Errorf("failed to power on DUT: %v", err)
	}

	var (
		state string
		err   error
	)

	// Check the if the device is powered on
	state, err = ts.getState(ctx, power)
	if err != nil {
		return err
	}

	if state == off {
		if ts.Image != "" {
			if err := ts.mountImage(ctx, outputBuf); err != nil {
				return fmt.Errorf("failed to mount image: %w", err)
			}
		}

		time.Sleep(time.Second)

		if err := ts.postPower(ctx, on); err != nil {
			return fmt.Errorf("failed to power on DUT: %v", err)
		}
	} else if state == on {
		outputBuf.WriteString("DUT was already powered on.\n")

		return nil
	}

	// Check if the device is on
	state, err = ts.getState(ctx, power)
	if err != nil {
		return err
	}

	if state != on {
		return fmt.Errorf("failed to power on DUT: State is '%s'", state)
	}

	outputBuf.WriteString("DUT was powered on successfully.\n")

	return nil
}

// powerOffSoft turns off the device.
func (ts *TestStep) powerOffSoft(ctx xcontext.Context, outputBuf *strings.Builder) error {
	var (
		state string
		err   error
	)

	// First check if device needs to be powered down
	state, err = ts.getState(ctx, power)
	if err != nil {
		return err
	}

	if state == on {
		if err := ts.postPower(ctx, off); err != nil {
			return fmt.Errorf("failed to power off DUT: %v", err)
		}
	}

	state, err = ts.getState(ctx, power)
	if err != nil {
		return err
	}

	if state == off {
		outputBuf.WriteString("DUT was powered off successfully.\n")
	} else {
		return fmt.Errorf("failed to power off DUT: DUT is still on")
	}

	return nil
}

// powerOffHard ensures that -> pdu is off & reset is on.
func (ts *TestStep) powerOffHard(ctx xcontext.Context, stdoutMsg *strings.Builder) error {
	if err := ts.resetDUT(ctx); err != nil {
		return fmt.Errorf("failed to reset DUT: %v", err)
	}

	stdoutMsg.WriteString("DUT was resetted successfully.\n")

	return nil
}

type postPower struct {
	State string `json:"state"` // possible values: on/off
}

// postPower sets the power state into the desired 'state'.
func (ts *TestStep) postPower(ctx xcontext.Context, state string) error {
	endpoint := fmt.Sprintf("%s%s/contexts/%s/machines/%s/auxiliaries/%s/api/power",
		ts.Host, ts.Version, ts.ContextID, ts.MachineID, ts.DeviceID)

	postPower := postPower{
		State: state,
	}

	powerBody, err := json.Marshal(postPower)
	if err != nil {
		return fmt.Errorf("failed to marshal body: %w", err)
	}

	for i := 0; i < 5; i++ {
		resp, err := HTTPRequest(ctx, http.MethodPost, endpoint, bytes.NewBuffer(powerBody))
		if err != nil {
			return fmt.Errorf("failed to do HTTP request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return fmt.Errorf("could not extract response body: %v", err)
			}

			if resp.StatusCode == http.StatusConflict && strings.Contains(string(body), "power button is busy") {
				time.Sleep(time.Second)

				continue
			}

			return fmt.Errorf("failed to Post to Power. Statuscode: %d, Response Body: %v", resp.StatusCode, string(body))
		}

		break
	}

	return nil
}

// pressPDU toggles the PDU as you define the method input parameter.
// http.MethodDelete does power off the pdu.
// http.MethodPut does power on the pdu.
func (ts *TestStep) pressPDU(ctx xcontext.Context, method string) error {
	if method != http.MethodDelete && method != http.MethodPut {
		return fmt.Errorf("invalid method '%s'. Only supported methods for toggeling the PDU are: '%s' and '%s'", method, http.MethodDelete, http.MethodPut)
	}

	endpoint := fmt.Sprintf("%s%s/contexts/%s/machines/%s/power",
		ts.Host, ts.Version, ts.ContextID, ts.MachineID)

	resp, err := HTTPRequest(ctx, method, endpoint, bytes.NewBuffer(nil))
	if err != nil {
		return fmt.Errorf("failed to do HTTP request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("could not extract response body: %v", err)
		}

		return fmt.Errorf("PDU could not be set to the correct state. Statuscode: %d, Response Body: %v", resp.StatusCode, string(body))
	} else {
		time.Sleep(time.Second)

		powerState, err := ts.getPDUState(ctx)
		if err != nil {
			return err
		}

		if method == http.MethodPut && !powerState || method == http.MethodDelete && powerState {
			return fmt.Errorf("failed to toggle PDU. Method: '%s', State: '%t'", method, powerState)
		}
	}

	return nil
}

type postReset struct {
	State string `json:"state"` // possible values: "on" or "off"
}

// postReset toggles the Reset button regarding the state that is passed in.
// A valid state is either 'on' or 'off'.
func (ts *TestStep) postReset(ctx xcontext.Context, wantState string) error {
	if wantState != on && wantState != off {
		return fmt.Errorf("invalid state '%s'. Only supported states for reset are: '%s' and '%s'", wantState, on, off)
	}

	endpoint := fmt.Sprintf("%s%s/contexts/%s/machines/%s/auxiliaries/%s/api/reset",
		ts.Host, ts.Version, ts.ContextID, ts.MachineID, ts.DeviceID)

	postReset := postReset{
		State: wantState,
	}

	resetBody, err := json.Marshal(postReset)
	if err != nil {
		return fmt.Errorf("failed to marshal body: %w", err)
	}

	resp, err := HTTPRequest(ctx, http.MethodPost, endpoint, bytes.NewBuffer(resetBody))
	if err != nil {
		return fmt.Errorf("failed to do HTTP request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("could not extract response body: %v", err)
		}

		return fmt.Errorf("reset could not be set to state '%s': %s", wantState, string(body))
	} else {
		state, err := ts.getState(ctx, reset)
		if err != nil {
			return err
		}

		if state != wantState {
			return fmt.Errorf("reset could not be set to state '%s'. State is '%s'", wantState, state)
		}
	}

	return nil
}

// this struct can be used for GET /vcc /power /reset
type getState struct {
	State string `json:"state"` // possible values: "on" or "off"
}

// getState returns the state of either: 'power', 'reset' or 'vcc'.
// The input parameter command should have one of this values.
func (ts *TestStep) getState(ctx xcontext.Context, command string) (string, error) {
	endpoint := fmt.Sprintf("%s%s/contexts/%s/machines/%s/auxiliaries/%s/api/%s",
		ts.Host, ts.Version, ts.ContextID, ts.MachineID, ts.DeviceID, command)

	resp, err := HTTPRequest(ctx, http.MethodGet, endpoint, bytes.NewBuffer(nil))
	if err != nil {
		return "", fmt.Errorf("failed to do HTTP request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("could not extract response body: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf(" %s state could not be retrieved. Statuscode: %d, Response Body: %s", command, resp.StatusCode, string(body))
	}

	data := getState{}

	if err := json.Unmarshal(body, &data); err != nil {
		return "", fmt.Errorf("could not unmarshal response body: %v", err)
	}

	return data.State, nil
}

// getPDUState returns the state of the PDU.
func (ts *TestStep) getPDUState(ctx xcontext.Context) (bool, error) {
	endpoint := fmt.Sprintf("%s%s/contexts/%s/machines/%s/power",
		ts.Host, ts.Version, ts.ContextID, ts.MachineID)

	resp, err := HTTPRequest(ctx, http.MethodGet, endpoint, bytes.NewBuffer(nil))
	if err != nil {
		return false, fmt.Errorf("failed to do HTTP request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, fmt.Errorf("could not extract response body: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf(" pdu state could not be retrieved. Statuscode: %d, Response Body: %s", resp.StatusCode, string(body))
	}

	var state bool

	if err := json.Unmarshal(body, &state); err != nil {
		return false, fmt.Errorf("could not unmarshal response body: %v", err)
	}

	return state, nil
}

// resetDUT sets the dut into a state, were it cannot be booted. In this state it is safe to
// do all flash operations.
func (ts *TestStep) resetDUT(ctx xcontext.Context) error {
	if err := ts.postReset(ctx, on); err != nil {
		return err
	}

	if err := ts.pressPDU(ctx, http.MethodDelete); err != nil {
		return err
	}

	time.Sleep(time.Second)

	return nil
}

// unresetDUT sets the dut into a state, were it can be booted again. PDU has to be turned on
// and reset has to pull on off.
func (ts *TestStep) unresetDUT(ctx xcontext.Context) error {
	if err := ts.postReset(ctx, off); err != nil {
		return err
	}

	if err := ts.pressPDU(ctx, http.MethodPut); err != nil {
		return err
	}

	time.Sleep(time.Second)

	return nil
}
