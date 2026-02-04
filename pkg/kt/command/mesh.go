package command

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"strings"

	"github.com/alibaba/kt-connect/pkg/kt/command/general"
	"github.com/alibaba/kt-connect/pkg/kt/command/mesh"
	opt "github.com/alibaba/kt-connect/pkg/kt/command/options"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// NewMeshCommand return new mesh command
func NewMeshCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mesh",
		Short: "Redirect marked requests of specified kubernetes service to local",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("name of service to mesh is required")
			} else if len(args) > 1 {
				return fmt.Errorf("too many service names are spcified (%s), should be one", strings.Join(args, ","))
			}
			return general.Prepare()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return Mesh(args[0])
		},
		Example: "ktctl mesh <service-name> [command options]",
	}

	cmd.SetUsageTemplate(general.UsageTemplate(true))
	opt.SetOptions(cmd, cmd.Flags(), opt.Get().Mesh, opt.MeshFlags())
	return cmd
}

//Mesh exchange kubernetes workload
func Mesh(resourceName string) error {
	ch, err := general.SetupProcess(util.ComponentMesh)
	if err != nil {
		return err
	}

	if opt.Get().Mesh.SkipPortChecking {
		if port := util.FindBrokenLocalPort(opt.Get().Mesh.Expose); port != "" {
			return fmt.Errorf("no application is running on port %s", port)
		}
	}

	// Setup signal file watcher
	signalFile := filepath.Join(os.TempDir(), fmt.Sprintf("ktctl-mesh-signal-%d", os.Getpid()))
	go watchMeshSignalFile(signalFile, ch)

	// Get service to mesh
	svc, err := general.GetServiceByResourceName(resourceName, opt.Get().Global.Namespace)
	if err != nil {
		// Clean up signal file
		os.RemoveAll(signalFile)
		return err
	}

	if port := util.FindInvalidRemotePort(opt.Get().Mesh.Expose, general.GetTargetPorts(svc)); port != "" {
		// Clean up signal file
		os.RemoveAll(signalFile)
		return fmt.Errorf("target port %s not exists in service %s", port, svc.Name)
	}

	log.Info().Msgf("Using %s mode", opt.Get().Mesh.Mode)
	if opt.Get().Mesh.Mode == util.MeshModeManual {
		err = mesh.ManualMesh(svc)
	} else if opt.Get().Mesh.Mode == util.MeshModeAuto {
		err = mesh.AutoMesh(svc)
	} else {
		err = fmt.Errorf("invalid mesh method '%s', supportted are %s, %s", opt.Get().Mesh.Mode,
			util.MeshModeAuto, util.MeshModeManual)
	}

	// Move signal file cleanup to deferred function to ensure it's only cleaned up at the end
	defer os.RemoveAll(signalFile)

	if err != nil {
		return err
	}

	log.Info().Msg("---------------------------------------------------------------")
	log.Info().Msgf(" Now all request to %s '%s' will be redirected to local", svc.Kind, svc.Name)
	log.Info().Msg("---------------------------------------------------------------")

	if util.IsWindows() {
		log.Info().Msgf("You can stop the mesh by creating a signal file:")
		log.Info().Msgf("PowerShell:   \"stop\" | Out-File -FilePath %s -Encoding ASCII", signalFile)
		log.Info().Msgf("Command Prompt: echo stop > %s", signalFile)
	} else {
		log.Info().Msgf("You can stop the mesh by creating a signal file: echo stop > %s", signalFile)
	}

	// watch background process, clean the workspace and exit if background process occur exception
	s := <-ch
	log.Info().Msgf("Terminal Signal is %s", s)

	return nil
}

func watchMeshSignalFile(signalFile string, ch chan os.Signal) {
	// Create the signal file to indicate mesh is ready
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
