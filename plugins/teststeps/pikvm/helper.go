package pikvm

import (
	"crypto/sha256"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/linuxboot/contest/pkg/xcontext"
)

const (
	plug   = true
	unplug = false
)

type Response struct {
	Ok     bool            `json:"ok"`
	Result json.RawMessage `json:"result"`
}

type ErrorResponse struct {
	Ok     bool `json:"ok"`
	Result struct {
		Error    string `json:"error"`
		ErrorMsg string `json:"error_msg"`
	} `json:"result"`
}
type StatusResponse struct {
	Ok     bool   `json:"ok"`
	Result Status `json:"result"`
}

type Status struct {
	Busy  bool `json:"busy"`
	Drive struct {
		Cdrom     bool `json:"cdrom"`
		Connected bool `json:"connected"`
		Image     struct {
			Complete  bool   `json:"complete"`
			InStorage bool   `json:"in_storage"`
			Removable bool   `json:"removable"`
			Name      string `json:"name"`
			Size      uint64 `json:"size"`
		} `json:"image"`
		Rw bool `json:"rw"`
	} `json:"drive"`
	Enabled bool `json:"enabled"`
	Online  bool `json:"online"`
	Storage struct {
		Downloading json.RawMessage `json:"downloading"`
		Images      json.RawMessage `json:"images"`
		Uploading   json.RawMessage `json:"uploading"`
	} `json:"storage"`
}

type StorageImages map[string]json.RawMessage

type StorageImage struct {
	Complete  bool    `json:"complete"`
	ModeTS    float32 `json:"mod_ts"`
	Removable bool    `json:"removable"`
	Size      uint64  `json:"size"`
}

var (
	transport = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client = http.Client{
		Transport: transport,
		Timeout:   20 * time.Minute,
	}
)

func (ts *TestStep) getUsbPlugStatus(ctx xcontext.Context) (Status, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, ts.Host, nil)
	if err != nil {
		return Status{}, err
	}

	req.SetBasicAuth(ts.Username, ts.Password)

	resp, err := client.Do(req)
	if err != nil {
		return Status{}, err
	}
	defer resp.Body.Close()

	var response StatusResponse

	if resp.StatusCode != http.StatusOK {
		var response ErrorResponse

		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			return Status{}, err
		}

		return Status{}, fmt.Errorf("failed to post request: %v", response)
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return Status{}, err
	}

	if !response.Ok {
		return Status{}, err
	}

	if response.Result.Busy {
		return Status{}, fmt.Errorf("pikvm mass-storage is busy")
	}
	if !response.Result.Enabled || !response.Result.Online {
		return Status{}, fmt.Errorf("pikvm mass-storage is not enabled or online")
	}

	return response.Result, nil
}

func (ts *TestStep) plugUSB(ctx xcontext.Context, plug bool) error {
	status, err := ts.getUsbPlugStatus(ctx)
	if err != nil {
		return err
	}

	if status.Drive.Connected == plug {
		return fmt.Errorf("virtual usb plug is already in the desired state")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("%s/set_connected?connected=%d", ts.Host, boolToInt(plug)), nil)
	if err != nil {
		return err
	}

	req.SetBasicAuth(ts.Username, ts.Password)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var response Response

	if resp.StatusCode != http.StatusOK {
		var response ErrorResponse

		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			return err
		}

		return fmt.Errorf("failed to post request: %d: %v", resp.StatusCode, response)
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return err
	}

	if !response.Ok {
		return fmt.Errorf("%v", response.Result)
	}

	return nil
}

var ErrMissingImage = errors.New("image not found in storage")

func (ts *TestStep) checkMountedImages(ctx xcontext.Context, hashSum string) error {
	status, err := ts.getUsbPlugStatus(ctx)
	if err != nil {
		return err
	}

	var storageImages map[string]StorageImage

	dec := json.NewDecoder(strings.NewReader(string(status.Storage.Images)))
	if err := dec.Decode(&storageImages); err != nil {
		return err
	}

	for storageImageHash := range storageImages {
		if storageImageHash == hashSum {
			return nil
		}
	}

	return ErrMissingImage
}

func (ts *TestStep) postMountImage(ctx xcontext.Context) error {
	file, err := os.Open(ts.Image)
	if err != nil {
		return fmt.Errorf("failed to open the image at the provided path: %v", err)
	}

	fileStat, err := file.Stat()
	if err != nil {
		return err
	}

	dataHash, err := calcSHA256(ts.Image)
	if err != nil {
		return err
	}

	r, w := io.Pipe()

	go func() {
		defer w.Close()
		if err != nil {
			w.CloseWithError(err)

			return
		}
		if _, err = io.Copy(w, file); err != nil {
			w.CloseWithError(err)

			return
		}
	}()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("%s/write?image=%s", ts.Host, dataHash), r)
	if err != nil {
		return err
	}

	req.Header.Set("Accept", "*/*")

	req.ContentLength = fileStat.Size()
	req.SetBasicAuth(ts.Username, ts.Password)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		var response Response

		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			if !response.Ok {
				return fmt.Errorf("%v", response.Result)
			}
		}

		return nil
	default:
		responseBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		return fmt.Errorf("failed to post request: %d: %s", resp.StatusCode, string(responseBytes))

	}
}

func (ts *TestStep) configureUSB(ctx xcontext.Context, imageName string) error {
	status, err := ts.getUsbPlugStatus(ctx)
	if err != nil {
		return err
	}

	if status.Drive.Connected {
		return fmt.Errorf("virtual usb plug is currently plugged in")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("%s/set_params?image=%s&cdrom=0&rw=1", ts.Host, imageName), nil)
	if err != nil {
		return err
	}

	req.SetBasicAuth(ts.Username, ts.Password)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var response Response

	if resp.StatusCode != http.StatusOK {
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			return err
		}

		return fmt.Errorf("failed to post request: %d: %v", resp.StatusCode, string(response.Result))
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return err
	}

	if !response.Ok {
		return fmt.Errorf("%v", response.Result)
	}

	return nil
}

func calcSHA256(path string) (string, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("image not found: %v", err)
	}

	hash := sha256.New()
	hash.Write(file)

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

func boolToInt(b bool) int {
	if b {
		return 1
	}

	return 0
}
