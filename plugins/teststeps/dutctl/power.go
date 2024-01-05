package dutctl

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/9elements/fti/pkg/dutctl"
	"github.com/9elements/fti/pkg/remote_lab/client"
	"github.com/9elements/fti/pkg/tools"
	"github.com/linuxboot/contest/pkg/target"
	"github.com/linuxboot/contest/pkg/xcontext"
)

const (
	tryTimeout = 15 * time.Second
)

func (r *TargetRunner) powerCmds(ctx xcontext.Context, stdoutMsg, stderrMsg *strings.Builder, target *target.Target) error {
	var (
		err          error
		dutInterface dutctl.DutCtl
		stdout       tools.LogginFunc
	)

	for tries := 0; tries < 2; tries += 1 {

		dutInterface, err = client.NewDutCtl("", false, r.ts.Parameter.Host, false, "", 0, 2)
		if err != nil {
			// Try insecure on port 10000
			if strings.Contains(r.ts.Parameter.Host, ":10001") {
				r.ts.Parameter.Host = strings.Split(r.ts.Parameter.Host, ":")[0] + ":10000"
			}

			dutInterface, err = client.NewDutCtl("", false, r.ts.Parameter.Host, false, "", 0, 2)
			if err != nil {
				stderrMsg.WriteString(fmt.Sprintf("Failed to connect to DUT: %v\n", err))

				time.Sleep(tryTimeout)
				continue
			}
		}

		defer func() {
			if dutInterface != nil {
				dutInterface.Close()
			}
		}()

		err = dutInterface.InitPowerPlugins(stdout)
		if err != nil {
			stderrMsg.WriteString(fmt.Sprintf("Failed to init power plugins: %v\n", err))

			time.Sleep(tryTimeout)
			continue
		}

		if len(r.ts.Parameter.Args) == 0 {
			return fmt.Errorf("Failed to execute the power command. Args is empty. Possible values are 'on', 'off' and 'powercycle'.")
		}

		switch r.ts.Parameter.Args[0] {
		case "on":
			if err := dutInterface.PowerOn(); err != nil {
				stderrMsg.WriteString(fmt.Sprintf("Failed to power on: %v\n", err))

				time.Sleep(tryTimeout)
				continue
			}

			stdoutMsg.WriteString("Successfully powered on DUT.\n")

			if len(r.ts.expectStepParams) != 0 {
				regexList, err := r.getRegexList()
				if err != nil {
					return fmt.Errorf("Failed to parse regex list: %v\n", err)
				}

				if err := r.serial(ctx, stdoutMsg, stderrMsg, dutInterface, regexList); err != nil {
					return fmt.Errorf("the expect '%s' was not found in the logs", r.ts.expectStepParams)
				}
			}

			return nil
		case "off":
			if err := dutInterface.PowerOff(); err != nil {
				stderrMsg.WriteString(fmt.Sprintf("Failed to power off: %v\n", err))

				time.Sleep(tryTimeout)
				continue
			}

			stdoutMsg.WriteString("Successfully powered off DUT.\n")

			return nil
		case "hardreset":
			if !dutInterface.HasHardReset() {
				return fmt.Errorf("The DUT does not support hardreset.")
			}
			if err := dutInterface.HardReset(); err != nil {
				stderrMsg.WriteString(fmt.Sprintf("Failed to hardreset: %v\n", err))
			}

			stdoutMsg.WriteString("Successfully reset(hard) the DUT.\n")

			if len(r.ts.expectStepParams) != 0 {
				regexList, err := r.getRegexList()
				if err != nil {
					return fmt.Errorf("Failed to parse regex list: %v\n", err)
				}

				if err := r.serial(ctx, stdoutMsg, stderrMsg, dutInterface, regexList); err != nil {
					return fmt.Errorf("the expect '%s' was not found in the logs", r.ts.expectStepParams)
				}
			}

			return nil

		case "powercycle":

			if len(r.ts.Parameter.Args) != 2 {
				return fmt.Errorf("You have to add only a second argument to specify how often you want to powercycle.")
			}

			reboots, err := strconv.Atoi(r.ts.Parameter.Args[1])
			if err != nil {
				return fmt.Errorf("powercycle amount could not be parsed: %v\n", err)
			}

			regexList, err := r.getRegexList()
			if err != nil {
				return fmt.Errorf("Failed to parse regex list: %v\n", err)
			}

			for i := 1; i < reboots; i++ {
				err = dutInterface.PowerOff()
				if err != nil {
					return fmt.Errorf("Failed to power off: %v\n", err)
				}

				err = dutInterface.PowerOn()
				if err != nil {
					return fmt.Errorf("Failed to power on: %v\n", err)
				}

				if err := r.serial(ctx, stdoutMsg, stderrMsg, dutInterface, regexList); err != nil {
					return fmt.Errorf("the expect '%v' was not found in the logs", r.ts.expectStepParams)
				}
			}

			stdoutMsg.WriteString(fmt.Sprintf("Successfully powercycled the DUT '%s'.\n", r.ts.Parameter.Args[1]))

			return nil
		default:
			return fmt.Errorf("Failed to execute the power command. The argument '%s' is not valid. Possible values are 'on', 'off' and 'powercycle'.", r.ts.Parameter.Args)
		}
	}

	return fmt.Errorf("Maximum number of retries for power command reached.")
}
