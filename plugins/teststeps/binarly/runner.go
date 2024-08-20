package binarly

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/linuxboot/contest/pkg/event/testevent"
	"github.com/linuxboot/contest/pkg/events"
	"github.com/linuxboot/contest/pkg/target"
	"github.com/linuxboot/contest/pkg/xcontext"
	"github.com/linuxboot/contest/plugins/teststeps/abstraction/options"
)

type Error struct {
	Msg string `json:"error"`
}

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

	if r.ts.Token == "" {
		outputBuf.WriteString(fmt.Sprintf("%v", fmt.Errorf("Token is required")))

		return events.EmitError(ctx, outputBuf.String(), target, r.ev)
	}

	client := &http.Client{}

	id, err := r.ts.postFileToAPI(ctx, client)
	if err != nil {
		outputBuf.WriteString(fmt.Sprintf("Failed to post file to binarly: %v", err))

		return events.EmitError(ctx, outputBuf.String(), target, r.ev)
	}

	result, err := r.ts.awaitResult(ctx, client, id)
	if err != nil {
		outputBuf.WriteString(fmt.Sprintf("Failed to get results from binarly: %v", err))

		return events.EmitError(ctx, outputBuf.String(), target, r.ev)
	}

	if err := events.EmitOuput(ctx, "binarly", result, target, r.ev); err != nil {
		outputBuf.WriteString(fmt.Sprintf("Failed to emit output: %v", err))

		return events.EmitError(ctx, outputBuf.String(), target, r.ev)
	}

	return events.EmitLog(ctx, outputBuf.String(), target, r.ev)
}

type StartScanResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

func (ts *TestStep) postFileToAPI(ctx context.Context, client *http.Client) (string, error) {
	// Prepare the file to be uploaded
	file, err := os.Open(ts.File)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	// Create a new multipart form request
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", filepath.Base(file.Name()))
	if err != nil {
		return "", fmt.Errorf("failed to create form file: %v", err)
	}
	_, err = io.Copy(part, file)
	if err != nil {
		return "", fmt.Errorf("failed to copy file: %v", err)
	}
	writer.Close()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, ts.URL, body)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+ts.Token)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	var response StartScanResponse
	if err = json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("failed to decode response: %v", err)
	}

	if response.Status != "ok" {
		return "", fmt.Errorf("failed to start scan: %v", response.Status)
	}

	return response.ID, nil
}

type ScanStatusResponse struct {
	FileName   string          `json:"file_name"`
	Scan       json.RawMessage `json:"scan"`
	Ratings    json.RawMessage `json:"ratings"`
	Status     string          `json:"status"`
	UploadTime string          `json:"upload_time"`
}

func (ts *TestStep) awaitResult(ctx context.Context, client *http.Client, id string) (json.RawMessage, error) {
	var statusResponse ScanStatusResponse

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/%s", ts.URL, id), nil)
	if err != nil {
		return json.RawMessage{}, fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+ts.Token)

	for {
		select {
		case <-ctx.Done():
			return json.RawMessage{}, fmt.Errorf("timed out waiting for results")
		case <-time.After(30 * time.Second):
			resp, err := client.Do(req)
			if err != nil {
				return json.RawMessage{}, fmt.Errorf("failed to send request: %v", err)
			}
			defer resp.Body.Close()

			if err = json.NewDecoder(resp.Body).Decode(&statusResponse); err != nil {
				return json.RawMessage{}, fmt.Errorf("failed to decode response: %v", err)
			}

			// Check the status
			if statusResponse.Status == "ok" {
				return statusResponse.Scan, nil
			}
		}
	}
}
