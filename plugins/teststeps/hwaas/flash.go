package hwaas

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"github.com/linuxboot/contest/pkg/xcontext"
)

func (p *Parameter) flashWrite(ctx xcontext.Context, arg string) error {
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

	if arg == "" {
		returnFunc(fmt.Errorf("no file was set to read or write"))

		return fmt.Errorf("no file was set to read or write")
	}

	endpoint := fmt.Sprintf("%s:%s/contexts/%s/machines/%s/flash", p.hostname, p.port, p.contextID, p.machineID)

	if isBusy := isTargetBusy(ctx, endpoint); isBusy {
		returnFunc(fmt.Errorf("target is currently busy"))

		return fmt.Errorf("target is currently busy")
	}

	err := flashTarget(ctx, endpoint, arg)
	if err != nil {
		returnFunc(fmt.Errorf("flashing %s failed: %v\n", arg, err))

		return err
	}

	log.Infof("successfully flashed binary")

	return nil
}

func isTargetBusy(ctx xcontext.Context, endpoint string) bool {
	log := ctx.Logger()

	resp, err := HTTPRequest(ctx, http.MethodGet, endpoint, bytes.NewBuffer(nil))
	if err != nil {
		log.Warnf("failed to do http request")
	}

	jsonBody, err := json.Marshal(resp.Body)
	if err != nil {
		log.Warnf("failed to marshal resp.Body")

		return false
	}

	if ctx.Writer() != nil {
		writer := ctx.Writer()
		_, err := writer.Write(jsonBody)
		if err != nil {
			log.Warnf("writing to ctx.Writer failed: %w", err)
		}
	}

	data := getFlash{}

	json.NewDecoder(resp.Body).Decode(&data)

	if data.State == "busy" {
		return true
	}

	return false
}

func flashTarget(ctx xcontext.Context, endpoint string, filePath string) error {
	log := ctx.Logger()

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

	jsonBody, err := json.Marshal(resp.Body)
	if err != nil {
		log.Warnf("failed to marshal resp.Body")

		return err
	}

	if ctx.Writer() != nil {
		writer := ctx.Writer()
		_, err := writer.Write(jsonBody)
		if err != nil {
			log.Warnf("writing to ctx.Writer failed: %w", err)
		}
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

	jsonBody, err = json.Marshal(resp.Body)
	if err != nil {
		log.Warnf("failed to marshal resp.Body")

		return err
	}

	if ctx.Writer() != nil {
		writer := ctx.Writer()
		_, err := writer.Write(jsonBody)
		if err != nil {
			log.Warnf("writing to ctx.Writer failed: %w", err)
		}
	}

	return nil
}
