package hwaas

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/linuxboot/contest/pkg/event/testevent"
	"github.com/linuxboot/contest/pkg/events"
	"github.com/linuxboot/contest/pkg/target"
	"github.com/linuxboot/contest/pkg/xcontext"
	"github.com/linuxboot/contest/plugins/teststeps/abstraction/options"
)

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

const (
	power    = "power"
	flash    = "flash"
	keyboard = "keyboard"
)

func (r *TargetRunner) Run(ctx xcontext.Context, target *target.Target) error {
	var outputBuf strings.Builder

	ctx, cancel := options.NewOptions(ctx, defaultTimeout, r.ts.options.Timeout)
	defer cancel()

	r.ts.writeTestStep(&outputBuf)

	writeCommand(r.ts.Command, r.ts.Args, &outputBuf)

	switch r.ts.Command {
	case power:
		var err error

		try := 1

		for ; try <= 3; try++ {
			if err = r.ts.powerCmds(ctx, &outputBuf); err != nil {
				outputBuf.WriteString(fmt.Sprintf("%v failed on try %d\n", err, try))

				time.Sleep(5 * time.Second)
				continue
			}

			break
		}

		if try == 4 {
			return events.EmitError(ctx, outputBuf.String(), target, r.ev, err)
		}

	case flash:
		if err := r.ts.flashCmds(ctx, &outputBuf); err != nil {
			outputBuf.WriteString(fmt.Sprintf("%v\n", err))

			return events.EmitError(ctx, outputBuf.String(), target, r.ev, err)
		}

	case keyboard:
		if err := r.ts.keyboardCmds(ctx, &outputBuf); err != nil {
			outputBuf.WriteString(fmt.Sprintf("%v\n", err))

			return events.EmitError(ctx, outputBuf.String(), target, r.ev, err)
		}

	default:
		err := fmt.Errorf("Command '%s' is not valid. Possible values are 'power', 'flash' and 'keyboard'.", r.ts.Command)
		outputBuf.WriteString(fmt.Sprintf("%v\n", err))

		return events.EmitError(ctx, outputBuf.String(), target, r.ev, err)
	}

	return events.EmitLog(ctx, outputBuf.String(), target, r.ev)
}

// HTTPRequest triggerers a http request and returns the response. The parameter that can be set are:
// method: can be every http method
// endpoint: api endpoint that shall be requested
// body: the body of the request
func HTTPRequest(ctx xcontext.Context, method string, endpoint string, body io.Reader) (*http.Response, error) {
	client := &http.Client{}

	req, err := http.NewRequestWithContext(ctx, method, endpoint, body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
