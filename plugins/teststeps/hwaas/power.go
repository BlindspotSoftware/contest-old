package hwaas

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/linuxboot/contest/pkg/xcontext"
)

func HTTPRequest(ctx xcontext.Context, method string, endpoint string, body io.Reader) (*http.Response, error) {
	log := ctx.Logger()

	client := &http.Client{}

	req, err := http.NewRequestWithContext(ctx, method, endpoint, body)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	jsonBody, err := json.Marshal(resp.Body)
	if err != nil {
		log.Warnf("failed to marshal resp.Body")

		return nil, err
	}

	if ctx.Writer() != nil {
		writer := ctx.Writer()
		_, err := writer.Write(jsonBody)
		if err != nil {
			log.Warnf("writing to ctx.Writer failed: %w", err)
		}
	}

	return resp, nil
}

func (p *Parameter) powerOn(ctx xcontext.Context) error {
	log := ctx.Logger()

	returnFunc := func(err error) {
		if ctx.Writer() != nil {
			writer := ctx.Writer()
			_, err := writer.Write([]byte(err.Error()))
			if err != nil {
				log.Warnf("writing to ctx.Writer failed: %w", err)
			}
		}

		return
	}

	// First pull reset switch on off
	endpoint := fmt.Sprintf("%s:%s/contexts/%s/machines/%s/auxiliaries/%s/api/reset",
		p.hostname, p.port, p.contextID, p.machineID, p.deviceID)

	postReset := postReset{
		State: "off",
	}

	resetBody, err := json.Marshal(postReset)
	if err != nil {
		return fmt.Errorf("failed to marshal body: %w", err)
	}

	resetResp, err := HTTPRequest(ctx, http.MethodPost, endpoint, bytes.NewBuffer(resetBody))
	if err != nil {
		returnFunc(fmt.Errorf("failed to do http request"))

		return err
	}
	defer resetResp.Body.Close()

	if resetResp.StatusCode == 200 {
		log.Infof("reset switch is off")
	} else {
		returnFunc(fmt.Errorf("device could not be set on reset"))

		return fmt.Errorf("device could not be set on reset")
	}

	// Than turn on the pdu again
	endpoint = fmt.Sprintf("%s:%s/contexts/%s/machines/%s/power", p.hostname, p.port, p.contextID, p.machineID)

	pduResp, err := HTTPRequest(ctx, http.MethodPut, endpoint, bytes.NewBuffer(nil))
	if err != nil {
		returnFunc(fmt.Errorf("failed to do http request"))

		return err
	}
	defer pduResp.Body.Close()

	if pduResp.StatusCode == 200 {
		log.Infof("pdu powered on")
	} else {
		returnFunc(fmt.Errorf("device could not be turned on"))

		return fmt.Errorf("pdu could not be powered off")
	}

	// Than press the power button
	endpoint = fmt.Sprintf("%s:%s/contexts/%s/machines/%s/auxiliaries/%s/api/power",
		p.hostname, p.port, p.contextID, p.machineID, p.deviceID)

	postPower := postPower{
		Duration: "3s",
	}

	powerBody, err := json.Marshal(postPower)
	if err != nil {
		return fmt.Errorf("failed to marshal body: %w", err)
	}

	powerResp, err := HTTPRequest(ctx, http.MethodPost, endpoint, bytes.NewBuffer(powerBody))
	if err != nil {
		returnFunc(fmt.Errorf("failed to do http request"))

		return err
	}
	defer powerResp.Body.Close()

	if powerResp.StatusCode == 200 {
		log.Infof("dut is starting")
		time.Sleep(1 * time.Second)
	} else {
		returnFunc(fmt.Errorf("device could not be turned on"))

		return fmt.Errorf("device could not be turned on")
	}

	// Check the led if the device is on
	endpoint = fmt.Sprintf("%s:%s/contexts/%s/machines/%s/auxiliaries/%s/api/led",
		p.hostname, p.port, p.contextID, p.machineID, p.deviceID)

	ledResp, err := HTTPRequest(ctx, http.MethodGet, endpoint, bytes.NewBuffer(nil))
	if err != nil {
		returnFunc(fmt.Errorf("failed to do http request"))

		return err
	}
	defer ledResp.Body.Close()

	body, err := io.ReadAll(ledResp.Body)
	if err != nil {
		returnFunc(fmt.Errorf("could not extract response body: %v", err))

		return fmt.Errorf("could not extract response body: %v", err)
	}

	if ledResp.StatusCode != 200 {
		returnFunc(fmt.Errorf("led status could not be retrieved"))

		return fmt.Errorf("led status could not be retrieved")
	}

	data := getState{}

	if err := json.Unmarshal(body, &data); err != nil {
		returnFunc(fmt.Errorf("could not unmarshal response body: %v", err))

		return fmt.Errorf("could not unmarshal response body: %v", err)
	}

	log.Infof("%s", data.State)

	if data.State != "on" {
		return fmt.Errorf("dut is not on")
	}

	return nil
}

func (p *Parameter) powerOff(ctx xcontext.Context) error {
	log := ctx.Logger()

	returnFunc := func(err error) {
		if ctx.Writer() != nil {
			writer := ctx.Writer()
			_, err := writer.Write([]byte(err.Error()))
			if err != nil {
				log.Warnf("writing to ctx.Writer failed: %w", err)
			}
		}

		return
	}

	// First press power button for 3s
	endpoint := fmt.Sprintf("%s:%s/contexts/%s/machines/%s/auxiliaries/%s/api/power",
		p.hostname, p.port, p.contextID, p.machineID, p.deviceID)

	postPower := postPower{
		Duration: "3s",
	}

	powerBody, err := json.Marshal(postPower)
	if err != nil {
		return fmt.Errorf("failed to marshal body: %w", err)
	}

	resp, err := HTTPRequest(ctx, http.MethodPost, endpoint, bytes.NewBuffer(powerBody))
	if err != nil {
		returnFunc(fmt.Errorf("failed to do http request"))

		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		log.Infof("dut is shutting down")

		time.Sleep(15 * time.Second)
	} else {
		log.Infof("dut is was not powered down gracefully")

		returnFunc(fmt.Errorf("device could not be turned off"))
	}

	// Than turn off the pdu, even if the graceful shutdown was not working
	endpoint = fmt.Sprintf("%s:%s/contexts/%s/machines/%s/power", p.hostname, p.port, p.contextID, p.machineID)

	resp, err = HTTPRequest(ctx, http.MethodDelete, endpoint, bytes.NewBuffer(nil))
	if err != nil {
		returnFunc(fmt.Errorf("failed to do http request"))

		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		log.Infof("pdu powered off")
	} else {
		returnFunc(fmt.Errorf("device could not be turned on"))

		return fmt.Errorf("pdu could not be powered off")
	}

	// Than pull the reset switch on on
	endpoint = fmt.Sprintf("%s:%s/contexts/%s/machines/%s/auxiliaries/%s/api/reset",
		p.hostname, p.port, p.contextID, p.machineID, p.deviceID)

	postReset := postReset{
		State: "on",
	}

	resetBody, err := json.Marshal(postReset)
	if err != nil {
		return fmt.Errorf("failed to marshal body: %w", err)
	}

	resp, err = HTTPRequest(ctx, http.MethodPost, endpoint, bytes.NewBuffer(resetBody))
	if err != nil {
		returnFunc(fmt.Errorf("failed to do http request"))

		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		log.Infof("reset is in state on")
	} else {
		returnFunc(fmt.Errorf("device could not be set on reset"))

		return fmt.Errorf("device could not be set on reset")
	}

	log.Infof("successfully powered down dut")

	return nil
}
