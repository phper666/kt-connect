package command

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"strings"

	"github.com/alibaba/kt-connect/pkg/kt/command/exchange"
	"github.com/alibaba/kt-connect/pkg/kt/command/general"
	opt "github.com/alibaba/kt-connect/pkg/kt/command/options"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// NewExchangeCommand return new exchange command
func NewExchangeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "exchange",
		Short: "Redirect all requests of specified kubernetes service to local",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("name of service to exchange is required")
			} else if len(args) > 1 {
				return fmt.Errorf("too many service names are spcified (%s), should be one", strings.Join(args, ","))
			}
			return general.Prepare()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return Exchange(args[0])
		},
		Example: "ktctl exchange <service-name> [command options]",
	}

	cmd.SetUsageTemplate(general.UsageTemplate(true))
	opt.SetOptions(cmd, cmd.Flags(), opt.Get().Exchange, opt.ExchangeFlags())
	return cmd
}

//Exchange exchange kubernetes workload
func Exchange(resourceName string) error {
	ch, err := general.SetupProcess(util.ComponentExchange)
	if err != nil {
		return err
	}

	if opt.Get().Exchange.SkipPortChecking {
		if port := util.FindBrokenLocalPort(opt.Get().Exchange.Expose); port != "" {
			return fmt.Errorf("no application is running on port %s", port)
		}
	}

	// Setup signal file watcher
	signalFile := filepath.Join(os.TempDir(), fmt.Sprintf("ktctl-exchange-signal-%d", os.Getpid()))
	go watchExchangeSignalFile(signalFile, ch)

	log.Info().Msgf("Using %s mode", opt.Get().Exchange.Mode)
	if opt.Get().Exchange.Mode == util.ExchangeModeScale {
		err = exchange.ByScale(resourceName)
	} else if opt.Get().Exchange.Mode == util.ExchangeModeEphemeral {
		err = exchange.ByEphemeralContainer(resourceName)
	} else if opt.Get().Exchange.Mode == util.ExchangeModeSelector {
		err = exchange.BySelector(resourceName)
	} else {
		err = fmt.Errorf("invalid exchange method '%s', supportted are %s, %s, %s", opt.Get().Exchange.Mode,
			util.ExchangeModeSelector, util.ExchangeModeScale, util.ExchangeModeEphemeral)
	}
	if err != nil {
		// Clean up signal file
		os.RemoveAll(signalFile)
		return err
	}
	resourceType, realName := toTypeAndName(resourceName)
	log.Info().Msg("---------------------------------------------------------------")
	log.Info().Msgf(" Now all request to %s '%s' will be redirected to local", resourceType, realName)
	log.Info().Msg("---------------------------------------------------------------")

	if util.IsWindows() {
		log.Info().Msgf("You can stop the exchange by creating a signal file:")
		log.Info().Msgf("PowerShell:   \"stop\" | Out-File -FilePath %s -Encoding ASCII", signalFile)
		log.Info().Msgf("Command Prompt: echo stop > %s", signalFile)
	} else {
		log.Info().Msgf("You can stop the exchange by creating a signal file: echo stop > %s", signalFile)
	}

	// watch background process, clean the workspace and exit if background process occur exception
	s := <-ch
	log.Info().Msgf("Terminal Signal is %s", s)

	// Clean up signal file
	os.RemoveAll(signalFile)
	return nil
}

func toTypeAndName(name string) (string, string) {
	parts := strings.Split(name, "/")
	if len(parts) > 1 {
		return parts[0], parts[1]
	} else {
		return "service", parts[0]
	}
}

func watchExchangeSignalFile(signalFile string, ch chan os.Signal) {
	// Create the signal file to indicate exchange is ready
	os.Create(signalFile)

	for {
		time.Sleep(1 * time.Second)

		// Check if signal file contains "stop"
		if content, err := os.ReadFile(signalFile); err == nil {
			if strings.TrimSpace(string(content)) == "stop" {
				// Send interrupt signal to the main routine
				ch <- os.Interrupt
				return
			}
		}
	}
}
