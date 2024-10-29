package options

import (
	"context"
	"time"

	"github.com/insomniacslk/xjson"
	"github.com/linuxboot/contest/pkg/xcontext"
)

const (
	Keyword = "options"
)

type Parameters struct {
	Timeout xjson.Duration `json:"timeout,omitempty"`
} 

//NewOptions returns a new instance of the options Parameters, creating a context and its cancle function for the specified timeout
func NewOptions(ctx xcontext.Context, defaultTimeout time.Duration, timeout xjson.Duration) (xcontext.Context, context.CancelFunc) {
	var cancel xcontext.CancelFunc

	if timeout == 0 {
		timeout = xjson.Duration(defaultTimeout)
	}

	ctx, cancel = xcontext.WithTimeout(ctx, time.Duration(timeout))
	
	return ctx, cancel
}
