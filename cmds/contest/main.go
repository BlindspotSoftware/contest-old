// Copyright (c) Facebook, Inc. and its affiliates.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

// This is a generated file, edits should be made in the corresponding source
// file, and this file regenerated using
//   `contest-generator --from core-plugins.yml`
// followed by
//   `gofmt -w contest.go`

package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	_ "net/http/pprof"

	"github.com/linuxboot/contest/cmds/contest/server"
)

func main() {
	sigs := make(chan os.Signal, 1)
	defer close(sigs)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGUSR1)
	go func() {
		_ = http.ListenAndServe("localhost:8081", nil)
	}()

	if err := server.Main(os.Args[0], os.Args[1:], sigs); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
