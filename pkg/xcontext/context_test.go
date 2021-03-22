// Copyright (c) Facebook, Inc. and its affiliates.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package xcontext

import (
	"context"
	"fmt"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
)

func tryLeak() {
	ctx := Background()
	ctx, _ = WithCancel(ctx)
	ctx, _ = WithNotify(ctx, ErrPaused)
	ctx.Until(nil)
	ctx = WithResetSignalers(ctx)
	ctx = WithStdContext(ctx, context.Background())
	_ = ctx
}

func TestGoroutineLeak(t *testing.T) {
	runtime.GC()
	runtime.Gosched()
	runtime.GC()
	runtime.Gosched()
	old := runtime.NumGoroutine()

	tryLeak()
	runtime.GC()
	runtime.Gosched()
	runtime.GC()
	runtime.Gosched()

	stack := make([]byte, 65536)
	n := runtime.Stack(stack, true)
	stack = stack[:n]
	require.GreaterOrEqual(t, old, runtime.NumGoroutine(), fmt.Sprintf("%s", stack))
}
