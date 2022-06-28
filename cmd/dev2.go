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
	"github.com/nhost/cli/environment"
	"github.com/nhost/cli/nhost"
	"github.com/nhost/cli/nhost/compose"
	"github.com/nhost/cli/util"
	"github.com/spf13/cobra"
	"os"
	"os/signal"
	"sync"
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

		//  create .nhost/ if it doesn't exist
		if err := os.MkdirAll(nhost.DOT_NHOST, os.ModePerm); err != nil {
			status.Errorln("Failed to initialize nhost data directory")
			return err
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO: pick a random port if default port 1337 is in use
		//	Fixes GH #129, by moving it from pre-run to run.
		//
		//	If the default port is not available,
		//	choose a random one.
		//if !util.PortAvailable(env.Port) {
		//	status.Info("Choose a different port with `nhost dev [--port]`")
		//	status.Fatal(fmt.Sprintf("port %s not available", env.Port))
		//}

		env.UpdateState(environment.Executing)

		var wg sync.WaitGroup
		wg.Add(1)

		//  add cleanup action in case of signal interruption
		signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
		go func() {
			<-stop
			wg.Done()
			os.Exit(0)
		}()

		//  Initialize cancellable context for this specific execution
		//ctx := env.Context
		ctx := context.Background()
		env.ExecutionContext, env.ExecutionCancel = context.WithCancel(ctx)

		//  Update environment state
		env.UpdateState(environment.Active)

		//p, err := loader.Load(types.ConfigDetails{
		//	WorkingDir: nhost.NHOST_DIR,
		//	ConfigFiles: []types.ConfigFile{
		//		{
		//			Content: []byte(testYaml),
		//			Config:  nil,
		//		},
		//	},
		//})
		//
		//if err != nil {
		//	panic(err)
		//}

		//spew.Dump(p)

		//  Start the services
		//dockerComposeConfig, err := compose.NewConfig(nil).BuildJSON()
		//if err != nil {
		//	return err
		//}

		// start the services
		// - spin up the services
		// - wait for the services to be ready
		// - apply migrations if they exist
		// - apply metadata
		// - restart auth and storage containers (to re-apply their metadata (and migrations))
		// - check healthchecks and open a browser window with hasura console

		ds := compose.DataStreams{Stderr: os.Stderr, Stdout: os.Stdout}

		dc, err := compose.WrapperCmd([]string{"up", "-d"}, ds)
		if err != nil {
			return err
		}

		err = dc.Run()
		if err != nil {
			return fmt.Errorf("failed to start services: %w", err)
		}

		// sleep for a bit to allow the services to be ready
		time.Sleep(time.Second * 10)

		// apply migrations if they exist
		{
			files, _ := os.ReadDir(nhost.MIGRATIONS_DIR)
			if len(files) > 0 {
				log.Debug("Applying migrations")
				migrate, err := compose.WrapperCmd([]string{"run", "--rm", "hasura-console", "hasura", "migrate", "apply", "--all-databases", "--disable-interactive"}, compose.DataStreams{})
				if err != nil {
					return fmt.Errorf("failed to apply migrations: %w", err)
				}

				out, err := migrate.CombinedOutput()
				if err != nil {
					return fmt.Errorf("failed to apply migrations: %w, %s", err, string(out))
				}

			}
		}

		// export & apply metadata
		{
			metaFiles, err := os.ReadDir(nhost.METADATA_DIR)
			if err != nil {
				return err
			}

			if len(metaFiles) == 0 {
				// Export metadata
				log.Debug("Exporting metadata")

				export, err := compose.WrapperCmd([]string{"run", "--rm", "hasura-console", "hasura", "metadata", "export"}, compose.DataStreams{})
				if err != nil {
					return fmt.Errorf("failed to export metadata: %w", err)
				}

				out, err := export.CombinedOutput()
				if err != nil {
					status.Errorln("Failed to export metadata")
					return fmt.Errorf("failed to export migrations: %w, %s", err, string(out))
				}
			}

			// apply metadata
			log.Debug("Applying metadata")

			metadata, err := compose.WrapperCmd([]string{"run", "--rm", "hasura-console", "hasura", "metadata", "apply"}, compose.DataStreams{})
			if err != nil {
				return fmt.Errorf("failed to apply metadata: %w", err)
			}

			out, err := metadata.CombinedOutput()
			if err != nil {
				status.Errorln("Failed to apply metadata")
				return fmt.Errorf("failed to apply metadata: %w, %s", err, string(out))
			}
		}

		// restart auth and storage containers (to re-apply their metadata (and migrations))
		{
			c, err := compose.WrapperCmd([]string{"restart", "auth", "storage"}, compose.DataStreams{})
			if err != nil {
				return fmt.Errorf("failed to restart auth and storage containers: %w", err)
			}

			out, err := c.CombinedOutput()
			if err != nil {
				status.Errorln("Failed to restart auth and storage containers")
				return fmt.Errorf("failed to restart auth and storage containers: %w, %s", err, string(out))
			}
		}

		fmt.Println("sleeping for 5 secs")
		time.Sleep(time.Second * 5)

		{
			fmt.Println("exporting metadata")
			export, err := compose.WrapperCmd([]string{"run", "--rm", "hasura-console", "hasura", "metadata", "export"}, compose.DataStreams{})
			if err != nil {
				return fmt.Errorf("failed to export metadata: %w", err)
			}

			out, err := export.CombinedOutput()
			if err != nil {
				status.Errorln("Failed to export metadata")
				return fmt.Errorf("failed to export migrations: %w, %s", err, string(out))
			}
		}

		fmt.Println("sleeping for 3 secs")
		time.Sleep(time.Second * 3)

		//  wait for Ctrl+C
		wg.Wait()

		//  Close the signal interruption channel
		close(stop)
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
	//	devCmd.PersistentFlags().BoolVarP(&expose, "expose", "e", false, "Expose local environment to public internet")

	//  Cobra supports local flags which will only run when this command
	//  is called directly, e.g.:
	//devCmd.Flags().BoolVarP(&background, "background", "b", false, "Run dev services in background")
}

/*
	//  Following code belongs to Nhost tunnelling service
	//  We've decided not to incorporate this feature inside the CLI until V2

	//  expose to public internet
	var exposed bool
		if expose {
			go func() {

				//  get the user's credentials
				credentials, err := nhost.LoadCredentials()
				if err != nil {
					log.WithField("component", "tunnel").Debug(err)
					log.WithField("component", "tunnel").Error("Failed to fetch authentication credentials")
					log.WithField("component", "tunnel").Info("Login again with `nhost login` and re-start the environment")
					log.WithField("component", "tunnel").Warn("We're skipping exposing your environment to the outside world")
					return
				} else {
					go func() {

						state := make(chan *tunnels.ClientState)
						client := &tunnels.Client{
							Address: "wahal.tunnel.nhost.io:443",
							Port:    port,
							Token:   credentials.Token,
							State:   state,
						}

						if err := client.Init(); err != nil {
							log.WithField("component", "tunnel").Debug(err)
							log.WithField("component", "tunnel").Error("Failed to initialize your tunnel")
							return
						}

						//  Listen for tunnel state changes
						go func() {
							for {
								change := <-state
								if *change == tunnels.Connecting {
									log.WithField("component", "tunnel").Debug("Connecting")
								} else if *change == tunnels.Connected {
									exposed = true
									log.WithField("component", "tunnel").Debug("Connected")
								} else if *change == tunnels.Disconnected {
									log.WithField("component", "tunnel").Debug("Disconnected")
								}
							}
						}()

						if err := client.Connect(); err != nil {
							log.WithField("component", "tunnel").Debug(err)
							log.WithField("component", "tunnel").Error("Failed to expose your environment to the outside world")
						}

					}()
				}
			}()
		}
*/
