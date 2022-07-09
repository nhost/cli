/*
MIT License

Copyright (c) Nhost

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package cmd

import (
	"context"
	"errors"
	"fmt"
	"github.com/avast/retry-go/v4"
	"github.com/nhost/cli/logger"
	"github.com/nhost/cli/nhost"
	"github.com/nhost/cli/nhost/service"
	"github.com/nhost/cli/util"
	"github.com/nhost/cli/watcher"
	"github.com/spf13/cobra"
	"os"
	"os/signal"
	"syscall"
	"time"
)

/*

	---------------------------------
	`nhost dev` Operational Strategy
	---------------------------------

	1.	Initialize your running environment.
	2.	Fetch the list of running containers.
	3.	Wrap those existing containers as services inside the runtime environment.
		This will save the container ID in the service structure, so that it can be used to simply
		restart the container later, instead of creating it from scratch.
	4.	Parse the Nhost project configuration from config.yaml,
		and wrap it on existing services configurations.
		This will update all the fields of the service, which until now, only contained the container ID.
		This also includes initializing the service config and host config.
	5. 	Run the services.
		5.1	If the service ID exists --> start the same container
			else {
			--> create the container from configuration, and attach it to the network.
			--> now start the newly created container.
		}
		5.2	Once the container has been started, save the new container ID and assigned Port, and updated address.
			This will ensure that the new port is used for attaching a reverse proxy to this service, if required.
*/

//  devCmd represents the dev command
var dev2Cmd = &cobra.Command{
	Use:        "dev2 [-p port]",
	Aliases:    []string{"up"},
	SuggestFor: []string{"list", "init"},
	Short:      "Start local development environment",
	Long:       `Initialize a local Nhost environment for development and testing.`,
	PreRunE: func(cmd *cobra.Command, args []string) error {

		//  check if nhost/ exists
		if !util.PathExists(nhost.NHOST_DIR) {
			status.Infoln("Initialize new app by running 'nhost init'")
			return errors.New("app not found in this directory")
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		var mgr service.Manager

		ctx, cancel := context.WithCancel(cmd.Context())
		defer cancel()

		config, err := nhost.GetConfiguration()
		if err != nil {
			return err
		}

		projectName, err := nhost.GetDockerComposeProjectName()
		if err != nil {
			return err
		}

		env, err := nhost.Env()
		if err != nil {
			return fmt.Errorf("failed to read .env.development: %v", err)
		}

		mgr = service.NewDockerComposeManager(config, env, nhost.GetCurrentBranch(), projectName, log, status, logger.DEBUG)
		gw := watcher.NewGitWatcher(status, log)

		go gw.Watch(ctx, 700*time.Millisecond, func(branch, ref string) error {
			err := retry.Do(func() error {
				err := mgr.SyncExec(ctx, func(ctx context.Context) error {
					branchWithRef := fmt.Sprintf("Using branch %s", branch)
					if ref != "" {
						branchWithRef = fmt.Sprintf("Using branch %s [#%s]", branch, ref[:7])
					}

					status.Executingln(branchWithRef)
					mgr.SetGitBranch(branch)
					err := mgr.StopSvc(ctx, "postgres")
					if err != nil {
						status.Errorln("Failed to stop postgres")
						return err
					}

					err = mgr.Start(ctx)
					if err != nil {
						status.Errorln("Failed to start services")
						return err
					}

					if err != nil {
						status.Errorln("Failed to restart services")
						return err
					}

					return nil
				})

				if err != nil {
					return err
				}
				return nil
			}, retry.Attempts(3))

			return err
		})

		signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

		go func() {
			err = mgr.SyncExec(ctx, func(ctx context.Context) error {
				startCtx, cancel := context.WithTimeout(ctx, time.Minute*3)
				defer cancel()

				return retry.Do(func() error {
					return mgr.Start(startCtx)
				}, retry.Attempts(3))
			})

			if ctx.Err() == context.Canceled {
				return
			}

			if err != nil {
				status.Errorln("Failed to start services")
				log.WithError(err).Error("Failed to start services")
				os.Exit(1)
			}

			if !noBrowser {
				openbrowser(fmt.Sprintf("http://localhost:%s", cmd.Flag("port").Value.String()))
			}
		}()

		// wait for stop signal
		<-stop
		cancel()

		exitCtx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()

		status.Executing("Exiting...")
		log.Debug("Exiting...")
		err = mgr.SyncExec(exitCtx, func(ctx context.Context) error {
			return mgr.Stop(exitCtx)
		})
		if err != nil {
			status.Errorln("Failed to stop services")
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(dev2Cmd)

	//  Here you will define your flags and configuration settings.

	//  Cobra supports Persistent Flags which will work for this command
	//  and all subcommands, e.g.:
	dev2Cmd.PersistentFlags().StringVarP(&env.Port, "port", "p", "1337", "Port for dev proxy")
	dev2Cmd.PersistentFlags().BoolVar(&noBrowser, "no-browser", false, "Don't open browser windows automatically")
}
