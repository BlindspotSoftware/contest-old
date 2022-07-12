// Copyright (c) Facebook, Inc. and id affiliates.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package dutctl

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/9elements/fti/pkg/dut"
	"github.com/9elements/fti/pkg/dutctl"
	"github.com/9elements/fti/pkg/remote_lab/client"
	"github.com/linuxboot/contest/pkg/event"
	"github.com/linuxboot/contest/pkg/event/testevent"
	"github.com/linuxboot/contest/pkg/target"
	"github.com/linuxboot/contest/pkg/test"
	"github.com/linuxboot/contest/pkg/xcontext"
	"github.com/linuxboot/contest/plugins/teststeps"
)

// Name is the name used to look this plugin up.
var Name = "dutctl"

// Dutctl is used to retrieve all the parameter, the plugin needs.
type Dutctl struct {
	serverAddr *test.Param // Addr to the server where the dut is running.
	command    *test.Param // Command that shall be run on the dut.
	argument   *test.Param // Argument that the command need.
	binary     *test.Param // Binary to write or read.
	substring  *test.Param // Substring that is expected in the serial.
	timeout    *test.Param // Timeout after that the cmd will terminate.
}

// Name returns the plugin name.
func (d Dutctl) Name() string {
	return Name
}

func emitEvent(ctx xcontext.Context, name event.Name, payload interface{}, tgt *target.Target, ev testevent.Emitter) error {
	payloadStr, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("cannot encode payload for event '%s': %w", name, err)
	}
	rm := json.RawMessage(payloadStr)
	evData := testevent.Data{
		EventName: name,
		Target:    tgt,
		Payload:   &rm,
	}
	if err := ev.Emit(ctx, evData); err != nil {
		return fmt.Errorf("cannot emit event EventURL: %w", err)
	}
	return nil
}

var Timeout time.Duration
var TimeTimeout time.Time

// Run executes the Dutctl action.
func (d *Dutctl) Run(ctx xcontext.Context, ch test.TestStepChannels, params test.TestStepParameters,
	ev testevent.Emitter, resumeState json.RawMessage) (json.RawMessage, error) {
	log := ctx.Logger()
	// Validate the parameter
	if err := d.validateAndPopulate(params); err != nil {
		return nil, err
	}
	f := func(ctx xcontext.Context, target *target.Target) error {
		var err error
		var dutInterface dutctl.DutCtl
		var stdout dut.LogginFunc
		var o dut.FlashOptions

		// expand args
		serverAddr, err := d.serverAddr.Expand(target)
		if err != nil {
			return fmt.Errorf("failed to expand argument 'serverAddr': %v", err)
		}
		command, err := d.command.Expand(target)
		if err != nil {
			return fmt.Errorf("failed to expand argument 'command': %v", err)
		}
		argument, err := d.argument.Expand(target)
		if err != nil {
			return fmt.Errorf("failed to expand argument 'argument': %v", err)
		}
		binary, err := d.binary.Expand(target)
		if err != nil {
			return fmt.Errorf("failed to expand argument 'binary': %v", err)
		}
		substring, err := d.substring.Expand(target)
		if substring == "" && command == "serial" {
			return fmt.Errorf("'substring' has to be set if you want to use the command 'serial'")

		}
		if err != nil {
			return fmt.Errorf("failed to expand argument 'substring': %v", err)
		}
		timeoutStr, err := d.timeout.Expand(target)
		if timeoutStr == "" && command == "serial" {
			return fmt.Errorf("'timeout' has to be set if you want to use the command 'serial'")

		}
		if err != nil {
			return fmt.Errorf("failed to expand argument 'timeout' %s: %v", timeoutStr, err)
		}

		if timeoutStr != "" {
			Timeout, err = time.ParseDuration(timeoutStr)
			if err != nil {
				return fmt.Errorf("cannot parse timeout parameter: %v", err)
			}
			TimeTimeout = time.Now().Add(Timeout)
		}

		if serverAddr != "" {
			var certFile string

			fp := "/" + filepath.Join("etc", "fti", "keys", "ca-cert.pem")
			if _, err := os.Stat(fp); err == nil {
				certFile = fp
			}

			if !strings.Contains(serverAddr, ":") {
				// Add default port
				if certFile == "" {
					serverAddr += ":10000"
				} else {
					serverAddr += ":10001"
				}
			}

			dutInterface, err = client.NewDutCtl("", false, serverAddr, false, "", 0, 2)
			if err != nil {
				// Try insecure on port 10000
				if strings.Contains(serverAddr, ":10001") {
					serverAddr = strings.Split(serverAddr, ":")[0] + ":10000"
				}

				dutInterface, err = client.NewDutCtl("", false, serverAddr, false, "", 0, 2)
				if err != nil {
					return err
				}
			}

		}

		defer func() {
			if dutInterface != nil {
				dutInterface.Close()
			}
		}()

		switch command {
		case "power":
			err = dutInterface.InitPowerPlugins(stdout)
			if err != nil {
				return fmt.Errorf("Failed to init power plugins: %v\n", err)
			}
			switch argument {
			case "on":
				err = dutInterface.PowerOn()
				if err != nil {
					return fmt.Errorf("Failed to power on: %v\n", err)
				}
				log.Infof("dut powered on.")
				if substring != "" {
					serial(ctx, dutInterface, substring)
				}
			case "off":
				err = dutInterface.PowerOff()
				if err != nil {
					return fmt.Errorf("Failed to power off: %v\n", err)
				}
				log.Infof("dut powered off.")
			default:
				return fmt.Errorf("Failed to execute the power command. The argument %q is not valid. Possible values are 'on' and 'off'.", argument)
			}
		case "flash":
			err = dutInterface.InitPowerPlugins(stdout)
			if err != nil {
				return fmt.Errorf("Failed to init power plugins: %v\n", err)
			}
			err = dutInterface.InitProgrammerPlugins(stdout)
			if err != nil {
				return fmt.Errorf("Failed to init programmer plugins: %v\n", err)
			}

			switch argument {
			case "read", "write", "verify":
				if binary == "" {
					return fmt.Errorf("No file was set to read, write or verify: %v\n", err)
				}
			default:
				return fmt.Errorf("Failed to execute the flash command. The argument %q is not valid. Possible values are 'read', 'write' and 'verify'.", argument)
			}

			switch argument {
			case "read":
				s, err := dutInterface.FlashSupportsRead()
				if err != nil {
					return fmt.Errorf("Error calling FlashSupportsRead\n")
				}
				if !s {
					return fmt.Errorf("Programmer doesn't support read op\n")
				}

				rom, err := dutInterface.FlashRead()
				if err != nil {
					return fmt.Errorf("Fail to read: %v\n", err)
				}

				err = ioutil.WriteFile(binary, rom, 0660)
				if err != nil {
					return fmt.Errorf("Failed to write file: %v\n", err)
				}

				log.Infof("dut flash was read.")
			case "write":
				rom, err := ioutil.ReadFile(binary)
				if err != nil {
					return fmt.Errorf("File '%s' could not be read successfully: %v\n", binary, err)
				}

				err = dutInterface.FlashWrite(rom, &o)
				if err != nil {
					return fmt.Errorf("Failed to write rom: %v\n", err)
				}

				log.Infof("dut flash was written.")

			case "verify":
				s, err := dutInterface.FlashSupportsVerify()
				if err != nil {
					return fmt.Errorf("FlashSupportsVerify returned error: %v\n", err)
				}

				if !s {
					return fmt.Errorf("Programmer doesn't support verify op\n")
				}

				rom, err := ioutil.ReadFile(binary)
				if err != nil {
					return fmt.Errorf("File could not be read successfully: %v", err)
				}

				err = dutInterface.FlashVerify(rom)
				if err != nil {
					return fmt.Errorf("Failed to verify: %v\n", err)
				}

				log.Infof("dut flash was verified.")

			default:
				return fmt.Errorf("Failed to execute the flash command. The argument %q is not valid. Possible values are 'read', 'write' and 'verify'.", argument)
			}
		case "serial":
			serial(ctx, dutInterface, substring)
		default:
			return fmt.Errorf("Command %q is not valid. Possible values are 'power', 'flash' and 'serial'.", argument)
		}

		return nil
	}

	return teststeps.ForEachTarget(Name, ctx, ch, f)
}

// Retrieve all the parameters defines through the jobDesc
func (d *Dutctl) validateAndPopulate(params test.TestStepParameters) error {
	// Retrieving parameter as json Raw.Message
	// validate the dut server addr
	d.serverAddr = params.GetOne("serverAddr")
	if d.serverAddr.IsEmpty() {
		return fmt.Errorf("missing or empty 'serverAddr' parameter")
	}
	// validate the dutctl cmd
	d.command = params.GetOne("command")
	if d.command.IsEmpty() {
		return fmt.Errorf("missing or empty 'command' parameter")
	}
	// validate the dutctl cmd argument
	d.argument = params.GetOne("argument")
	if d.argument.IsEmpty() {
		return fmt.Errorf("missing or empty 'argument' parameter")
	}
	// validate the dutctl cmd binary
	d.binary = params.GetOne("binary")

	// validate the dutctl cmd substring
	d.substring = params.GetOne("substring")

	// validate the dutctl cmd timeout
	d.timeout = params.GetOne("timeout")

	return nil
}

// ValidateParameters validates the parameters associated to the TestStep
func (d *Dutctl) ValidateParameters(_ xcontext.Context, params test.TestStepParameters) error {
	return d.validateAndPopulate(params)
}

// New initializes and returns a new awsDutctl test step.
func New() test.TestStep {
	return &Dutctl{}
}

// Load returns the name, factory and evend which are needed to register the step.
func Load() (string, test.TestStepFactory, []event.Name) {
	return Name, New, nil
}

func serial(ctx xcontext.Context, dutInterface dutctl.DutCtl, substring string) error {
	log := ctx.Logger()

	err := dutInterface.InitSerialPlugins()
	if err != nil {
		return fmt.Errorf("Failed to init serial plugins: %v\n", err)
	}
	iface, err := dutInterface.GetSerial(0)
	if err != nil {
		return fmt.Errorf("Failed to get serial: %v\n", err)
	}

	//quit := make(chan bool)
	go func(ctx xcontext.Context) {
		defer func() {
			iface.Close()
		}()
		for {
			var n int
			select {
			case <-ctx.Done():
				return
			default:
				buf := make([]byte, 16)
				n, err = os.Stdin.Read(buf)
				if err != nil {
					return
				}
				for i := 0; i < n; i++ {
					if buf[i] == 0xd {
						buf[i] = 0xa
					}
				}
				_, err = iface.Write(buf[:n])
				if err != nil {
					return
				}
			}
		}
	}(ctx)
	_, err = os.Create("/tmp/dutctlserial")
	if err != nil {
		return fmt.Errorf("Creating serial dst file failed: %v", err)
	}
	dst, err := os.OpenFile("/tmp/dutctlserial", os.O_RDWR|os.O_APPEND, 0666)
	if err != nil {
		return fmt.Errorf("Opening serial dst file failed: %v", err)
	}
	defer dst.Close()

	go func(ctx xcontext.Context) error {
		defer func() {
			iface.Close()
		}()
		for {
			select {
			case <-ctx.Done():
				return nil
			default:
				_, err = io.Copy(dst, iface)
				if err != nil {
					return fmt.Errorf("Failed to copy serial to buffer: %v", err)
				}
			}
		}
	}(ctx)

	log.Infof("Greping serial from dut.")

	for {
		if time.Now().After(TimeTimeout) {
			return fmt.Errorf("timed out after %s", Timeout)
		}
		serial, err := ioutil.ReadFile("/tmp/dutctlserial")
		if err != nil {
			return fmt.Errorf("Failed to read serial file: %v", err)
		}
		if strings.Contains(string(serial), substring) {
			ctx.Done()
			return nil
		}
	}
}
