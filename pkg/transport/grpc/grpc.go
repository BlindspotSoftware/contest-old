package grpc

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"

	"github.com/linuxboot/contest/pkg/api"
	"github.com/linuxboot/contest/pkg/job"
	"github.com/linuxboot/contest/pkg/types"
	"github.com/linuxboot/contest/pkg/xcontext"
	"github.com/linuxboot/contest/plugins/listeners/grpclistener"

	"github.com/insomniacslk/xjson"
)

type Endpoint struct {
	buffer io.Reader
}

// HttpPartiallyDecodedResponse is a httplistener.GRPCAPIResponse, but with the Data not fully decoded yet
type GRPCResponse struct {
	ServerID string
	Type     string
	Data     json.RawMessage
	Error    *xjson.Error
}

// GRPC communicates with ConTest Server via http(s)/json transport
// GRPC implements the Transport interface
type GRPC struct {
	Addr string
}

func (g *GRPC) Version(ctx xcontext.Context, requestor string) (*api.VersionResponse, error) {
	return nil, nil
}

func (g *GRPC) Start(ctx xcontext.Context, requestor string, jobDescriptor string) (*api.StartResponse, error) {
	params := url.Values{}
	params.Add("jobDesc", jobDescriptor)
	resp, err := g.request(ctx, requestor, "start", params)
	if err != nil {
		return nil, err
	}
	data := api.ResponseDataStart{}
	if string(resp.Data) != "" {
		if err := json.Unmarshal([]byte(resp.Data), &data); err != nil {
			return nil, fmt.Errorf("cannot decode json response: %v", err)
		}
	}
	return &api.StartResponse{ServerID: resp.ServerID, Data: data, Err: resp.Error}, nil
}

func (g *GRPC) Stop(ctx xcontext.Context, requestor string, jobID types.JobID) (*api.StopResponse, error) {
	return nil, nil
}

func (g *GRPC) Status(ctx xcontext.Context, requestor string, jobID types.JobID) (*api.StatusResponse, error) {
	params := url.Values{}
	params.Add("jobID", strconv.Itoa(int(jobID)))
	resp, err := g.request(ctx, requestor, "status", params)
	if err != nil {
		return nil, err
	}
	data := api.ResponseDataStatus{}
	if string(resp.Data) != "" {
		if err := json.Unmarshal([]byte(resp.Data), &data); err != nil {
			return nil, fmt.Errorf("cannot decode json response: %v", err)
		}
	}
	return &api.StatusResponse{ServerID: resp.ServerID, Data: data, Err: resp.Error}, nil
}

func (g *GRPC) Retry(ctx xcontext.Context, requestor string, jobID types.JobID) (*api.RetryResponse, error) {
	return nil, nil
}

func (g *GRPC) List(ctx xcontext.Context, requestor string, states []job.State, tags []string) (*api.ListResponse, error) {
	return nil, nil
}

func (g *GRPC) request(ctx xcontext.Context, requestor string, verb string, params url.Values) (*GRPCResponse, error) {
	logger := xcontext.LoggerFrom(ctx)

	params.Set("requestor", requestor)
	u, err := url.Parse(g.Addr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse server address '%s': %v", g.Addr, err)
	}
	if u.Scheme == "" {
		return nil, errors.New("server URL scheme not specified")
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, fmt.Errorf("unsupported URL scheme '%s', please specify either http or https", u.Scheme)
	}
	u.Path += "/" + verb
	for k, v := range params {
		logger = logger.WithField(k, v)
	}
	logger.Debugf("Requesting URL %s with requestor ID '%s'\n", u.String(), requestor)
	resp, err := http.PostForm(u.String(), params)
	if err != nil {
		return nil, fmt.Errorf("GRPC POST failed: %v", err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("cannot read GRPC response: %v", err)
	}
	xcontext.LoggerFrom(ctx).Debugf("The server responded with status %s\n", resp.Status)

	var apiResp GRPCResponse
	if resp.StatusCode == http.StatusOK {
		// the Data field of apiResp will result in a map[string]interface{}
		if err := json.Unmarshal(body, &apiResp); err != nil {
			return nil, fmt.Errorf("response is not a valid GRPC API response object: '%s': %v", body, err)
		}
		if err != nil {
			return nil, fmt.Errorf("cannot marshal GRPCAPIResponse: %v", err)
		}
	} else {
		var apiErr grpclistener.GRPCAPIError
		if err := json.Unmarshal(body, &apiErr); err != nil {
			return nil, fmt.Errorf("response is not a valid GRPC API Error object: '%s': %v", body, err)
		}
		apiResp.Error = xjson.NewError(errors.New(apiErr.Msg))
	}

	return &apiResp, nil
}
