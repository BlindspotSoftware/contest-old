package ping

import (
	"fmt"
	"net"
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

type Error struct {
	Msg string `json:"error"`
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

	if r.ts.Port == 0 {
		r.ts.Port = defaultPort
	}

	// for any ambiguity, outcome is an error interface, but it encodes whether the process
	// was launched sucessfully and it resulted in a failure; err means the launch failed
	if err := r.runPing(&outputBuf); err != nil {
		return events.EmitError(ctx, outputBuf.String(), target, r.ev, err)
	}

	return events.EmitLog(ctx, outputBuf.String(), target, r.ev)
}

func (r *TargetRunner) runPing(outputBuf *strings.Builder) error {
	// Set timeout
	timeTimeout := time.After(time.Duration(r.ts.options.Timeout))
	ticker := time.NewTicker(time.Second)

	writeCommand(fmt.Sprintf("'%s:%d'", r.ts.Host, r.ts.Port), outputBuf)

	for {
		select {
		case <-timeTimeout:
			if r.ts.Expect.ShouldFail {
				outputBuf.WriteString(fmt.Sprintf("Ping Output:\nCouldn't connect to host '%s' on port '%d'", r.ts.Host, r.ts.Port))

				return nil
			}
			err := fmt.Errorf("Timeout, port %d was not opened in time.", r.ts.Port)

			outputBuf.WriteString(fmt.Sprintf("Ping Output:\n%s", err.Error()))

			return err

		case <-ticker.C:
			conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", r.ts.Host, r.ts.Port))
			if err != nil {
				break
			}
			defer conn.Close()

			outputBuf.WriteString(fmt.Sprintf("Ping Output:\nSuccessfully pinged '%s' on port '%d'", r.ts.Host, r.ts.Port))

			return nil
		}
	}
}
