package environment

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	client "github.com/docker/docker/client"
	"github.com/nhost/cli/nhost"
	"github.com/nhost/cli/util"
	"github.com/nhost/cli/watcher"
	"github.com/sirupsen/logrus"
)

//  State represents current state of any structure
type State uint32

//  State enumeration
const (
	Unknown State = iota
	Initializing
	Intialized
	Executing
	HealthChecks
	Active
	ShuttingDown
	Inactive
	Failed //  keep it always last
)

func (e *Environment) UpdateState(state State) {
	e.Lock()
	e.State = state

	if e.State != HealthChecks {
		status.Reset()
	}

	switch e.State {
	case Executing:
		status.Executing("Starting your app")
	case Initializing:
		status.Executing("Initializing environment")
	case ShuttingDown:
		status.Executing("Please wait while we cleanup")
	case HealthChecks:
		status.Executing("Running quick health checks")
	case Active:
		status.Success(fmt.Sprintf("Your app is running at %shttp://localhost:%s%s %s(Ctrl+C to stop)%s", util.Blue, e.Port, util.Reset, util.Gray, util.Reset))
		if !e.Config.Services["mailhog"].NoContainer {
			fmt.Println()
			status.Info(fmt.Sprintf("%sEmails will be sent to http://localhost:%d%s", util.Gray, e.Config.Services["mailhog"].Port, util.Reset))
		}
		status.Reset()
	case Inactive:
		status.Success("See you later, grasshopper!")
	case Failed:
		status.Fatal("App has crashed")
	}
	e.Unlock()

}

func (e *Environment) Init() error {

	var err error

	log.Debug("Initializing environment")

	//  Update environment state
	e.UpdateState(Initializing)

	//  connect to docker client
	e.Context, e.Cancel = context.WithCancel(context.Background())
	e.Docker, err = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}
	defer e.Docker.Close()

	//  break execution if docker deamon is not running
	_, err = e.Docker.Info(e.Context)
	if err != nil {
		return err
	}

	//  get running containers with prefix "nhost_"
	containers, err := e.GetContainers()
	if err != nil {
		status.Errorln(util.ErrServicesNotFound)
		return err
	}

	//  wrap the fetched containers inside the environment
	if err := e.WrapContainersAsServices(containers); err != nil {
		return err
	}

	//	Initialize a new watcher for the environment
	e.Watcher = watcher.New(e.Context)

	//  If there is a local git repository,
	//  in the project root directory,
	//  then initialize watchers for keeping track
	//  of HEAD changes on git checkout/pull/merge/fetch

	if util.PathExists(nhost.GIT_DIR) {

		//  Initialize watcher for post-checkout branch changes
		e.Watcher.Register(filepath.Join(nhost.GIT_DIR, "HEAD"), e.restartAfterCheckout)

		//  Initialize watcher for post-merge commit changes
		head := getBranchHEAD(filepath.Join(nhost.GIT_DIR, "refs", "remotes", nhost.REMOTE))
		if head != "" {
			e.Watcher.Register(head, e.restartMigrations)
		}
	}

	//  Update environment state
	e.UpdateState(Intialized)

	return nil
}

//
//  Performs default migrations and metadata operations
//

func (e *Environment) Prepare() error {

	//  Send out de-activation signal before starting migrations,
	//  to inform any other resource which using Hasura
	//  e.Config.Services["hasura"].Deactivate()

	//  Enable a mutual exclusion lock,
	//  to prevent other resources from
	//  modifying Hasura's data while
	//  we're making changes

	//  This will also prevent
	//  multiple resources inside our code
	//  from concurrently making changes to the Hasura service

	//  NOTE: This mutex lock only works
	//  for resources talking to Hasura service inside this code
	//  It doesn't lock anything for external resources, obviously!
	//  e.Config.Services["hasura"].Lock()

	//  defer e.Config.Services["hasura"].Activate()
	//  defer e.Config.Services["hasura"].Unlock()

	//  If migrations directory is already mounted to nhost_hasura container,
	//  then Hasura must be auto-applying migrations
	//  hence, manually applying migrations doesn't make sense

	//  create migrations
	files, _ := os.ReadDir(nhost.MIGRATIONS_DIR)
	if len(files) > 0 {

		log.Debug("Applying migrations")

		execute := exec.CommandContext(e.ExecutionContext, e.Hasura.CLI)
		execute.Dir = nhost.NHOST_DIR

		cmdArgs := []string{e.Hasura.CLI, "migrate", "apply"}
		cmdArgs = append(cmdArgs, e.Hasura.CommonOptions...)
		execute.Args = cmdArgs

		output, err := execute.CombinedOutput()
		if err != nil {
			log.Debug(string(output))
			status.Errorln("Failed to apply migrations")
			return err
		}
	}

	metaFiles, err := os.ReadDir(nhost.METADATA_DIR)
	if err != nil {
		return err
	}

	if len(metaFiles) == 0 {

		// Export metadata
		log.Debug("Exporting metadata")

		execute := exec.CommandContext(e.ExecutionContext, e.Hasura.CLI)
		execute.Dir = nhost.NHOST_DIR

		cmdArgs := []string{e.Hasura.CLI, "metadata", "export"}
		cmdArgs = append(cmdArgs, e.Hasura.CommonOptionsWithoutDB...)
		execute.Args = cmdArgs

		output, err := execute.CombinedOutput()
		if err != nil {
			log.Debug(string(output))
			status.Errorln("Failed to export metadata")
			return err
		}
	}

	// apply metadata
	log.Debug("Applying metadata")

	execute := exec.CommandContext(e.ExecutionContext, e.Hasura.CLI)
	execute.Dir = nhost.NHOST_DIR

	cmdArgs := []string{e.Hasura.CLI, "metadata", "apply"}
	cmdArgs = append(cmdArgs, e.Hasura.CommonOptionsWithoutDB...)
	execute.Args = cmdArgs

	output, err := execute.CombinedOutput()
	if err != nil {
		log.Debug(string(output))
		status.Errorln("Failed to apply metadata")
		return err
	}

	// Reload Hasura Auth and Hasura Storage to re-apply their metadata (and migrations)
	for _, x := range []string{"auth", "storage"} {
		log.Debugf("Restarting %s container", x)
		//	Restart the container
		if err := e.Docker.ContainerRestart(e.Context, e.Config.Services[x].ID, nil); err != nil {
			return err
		}
	}

	//	Run helath-check again to wait for restarted containers to become active
	if err := e.HealthCheck(e.ExecutionContext); err != nil {
		return err
	}

	// Exporting metadata to keep local metadata in sync.
	log.Debug("Exporting metadata")
	execute = exec.CommandContext(e.ExecutionContext, e.Hasura.CLI)
	execute.Dir = nhost.NHOST_DIR
	cmdArgs = []string{e.Hasura.CLI, "metadata", "export"}
	cmdArgs = append(cmdArgs, e.Hasura.CommonOptionsWithoutDB...)
	execute.Args = cmdArgs

	output, err = execute.CombinedOutput()
	if err != nil {
		log.Debug(string(output))
		status.Errorln("Failed to export metadata")
		return err
	}

	return nil
}

//  Runs concurrent healthchecks on all Nhost services,
//  which have a health check endpoint.
//
//  Also, supports process cancellation from contexts.
func (e *Environment) HealthCheck(ctx context.Context) error {

	e.UpdateState(HealthChecks)
	log.Debug("Starting health check")

	var err error
	var health_waiter sync.WaitGroup
	for _, service := range e.Config.Services {
		if service.HealthEndpoint != "" {
			status.Update(1)
			health_waiter.Add(1)
			go func(service *nhost.Service) {

				for counter := 1; counter <= 240; counter++ {
					select {
					case <-ctx.Done():
						log.WithFields(logrus.Fields{
							"type":      "service",
							"container": service.Name,
						}).Debug("Health check cancelled")
						return
					default:
						if healthy := service.Healthz(); healthy {
							status.Increment(1)
							log.WithFields(logrus.Fields{
								"type":      "service",
								"container": service.Name,
							}).Debug("Health check successful")

							//  Activate the service
							service.Activate()

							health_waiter.Done()

							return
						}
						time.Sleep(1 * time.Second)
						log.WithFields(logrus.Fields{
							"type":      "container",
							"component": service.Name,
						}).Debugf("Health check attempt #%v unsuccessful", counter)
					}
				}

				status.Error("Health check failed for " + service.Name)
				err = errors.New("health check of at least 1 service has failed")

			}(service)
		} else {

			//
			//	If any service doesn't have a health check endpoint,
			//	then, by default, declare it active.
			//
			//	This is being done to prevent the enviroment from failing activation checks.
			//	Because we are not performing health checks on postgres and mailhog,
			//	they end up making the entire environment fail activation check.
			service.Active = true
		}
	}

	//  wait for all healthchecks to pass
	health_waiter.Wait()
	return err
}

func (e *Environment) Seed(path string) error {

	seed_files, err := ioutil.ReadDir(path)
	if err != nil {
		return err
	}

	if len(seed_files) > 0 {
		log.Debug("Applying seeds")
	}

	//  if there are more seeds than just enum tables,
	//  apply them too
	for _, item := range seed_files {

		//  read seed file
		data, err := ioutil.ReadFile(filepath.Join(path, item.Name()))
		if err != nil {
			status.Errorln("Failed to open:" + item.Name())
			return err
		}

		//  apply seed data
		if err := e.Hasura.Seed(string(data)); err != nil {
			status.Errorln("Failed to apply:" + item.Name())
			return err
		}
		/*
			cmdArgs = []string{hasuraCLI, "seed", "apply", "--database-name", "default"}
			cmdArgs = append(cmdArgs, commandConfiguration...)
			execute.Args = cmdArgs

			if err = execute.Run(); err != nil {
				status.Errorln("Failed to apply seeds")
				return err
			}
		*/
	}
	return nil
}

func (e *Environment) Cleanup() {

	e.UpdateState(ShuttingDown)

	//  Gracefully shut down all registered servers of the environment
	for _, server := range e.Servers {
		server.Shutdown(e.Context)
	}

	//	Close the watcher
	e.Watcher.Close()

	if e.State >= Executing {

		//	Pass the parent context of the environment,
		//	because this is the final cleanup procedure
		//	and we are going to cancel this context shortly after
		if err := e.Shutdown(true, e.Context); err != nil {
			log.Debug(err)
			status.Error("Failed to stop running services")
		}
	}

	e.UpdateState(Inactive)

	//  Don't cancel the contexts before shutting down the containers
	e.Cancel()
}
