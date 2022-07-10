package compose

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/nhost/cli/util"
	"syscall"
	"time"
)

const (
	// default ports
	serverDefaultPort        = 1337
	svcPostgresDefaultPort   = 5432
	svcHasuraDefaultPort     = 8080
	svcMailhogDefaultPort    = 1025
	hasuraConsoleDefaultPort = 9695
	hasuraConsoleApiPort     = 9693

	highestTcpPort = 65535

	portsRandomizeTimeout = time.Second * 10
)

type svc struct {
	Service    string `json:"Service"`
	Publishers []struct {
		PublishedPort uint32 `json:"PublishedPort"`
	}
}

// isPortAvailable returns true if the port is published for the service
func (s svc) isPortPublished(port uint32) bool {
	for _, p := range s.Publishers {
		if p.PublishedPort == 0 {
			continue
		}

		if p.PublishedPort == port {
			return true
		}
	}

	return false
}

type Ports map[string]uint32

func DefaultPorts() Ports {
	return Ports{
		SvcTraefik:           serverDefaultPort,
		SvcPostgres:          svcPostgresDefaultPort,
		SvcGraphqlEngine:     svcHasuraDefaultPort,
		SvcMailhog:           svcMailhogDefaultPort,
		HasuraConsole:        hasuraConsoleDefaultPort,
		HasuraConsoleApiPort: hasuraConsoleApiPort,
	}
}

func NewPorts(proxy uint32) Ports {
	p := DefaultPorts()
	p[SvcTraefik] = proxy
	return p
}

// hasura console exposes 2 ports, one for the console itself and one for the migrate api
func (p Ports) isHasuraConsole(svc string) bool {
	return svc == HasuraConsole || svc == HasuraConsoleApiPort
}

// EnsurePortsAvailable ensures that the ports are available and returns a new instance of Ports to prevent unexpected mutations
func (p Ports) EnsurePortsAvailable(ctx context.Context, config *Config) (Ports, error) {
	newPorts := make(Ports)
	// early return if context is cancelled
	if ctx.Err() == context.Canceled {
		return nil, nil
	}

	// get running services
	runningServices, err := p.getRunningServices(ctx, config)
	if err != nil {
		return nil, err
	}

	// loop through all services/ports and make sure ports are available
	for svcName, port := range p {
		if util.PortAvailable(port) {
			// port is available, skip all the checks
			newPorts[svcName] = port
			continue
		}

		// hasura console is served by hasura cli, not in docker compose
		// it exposes 2 ports, one for the console itself and one for the migrate api
		if svcName == HasuraConsole || svcName == HasuraConsoleApiPort {
			randomPort, err := p.getRandomPort(ctx, port, portsRandomizeTimeout)
			if err != nil {
				return nil, fmt.Errorf("failed to get random port for %s: %w", svcName, err)
			}

			newPorts[svcName] = randomPort
			continue
		}

		// check if port is published for the service
		if runningSvc, ok := runningServices[svcName]; ok && runningSvc.isPortPublished(port) {
			// port is published and is used in running service, no need to randomize
			newPorts[svcName] = port
			continue
		}

		// port is not available, randomize it
		randomPort, err := p.getRandomPort(ctx, port, portsRandomizeTimeout)
		if err != nil {
			return nil, fmt.Errorf("failed to get random port for %s: %w", svcName, err)
		}

		newPorts[svcName] = randomPort
	}

	return newPorts, nil
}

func (p Ports) getRunningServices(ctx context.Context, config *Config) (map[string]svc, error) {
	runningServices := make(map[string]svc)
	cmd, err := WrapperCmd(ctx, []string{"ps", "--filter", "status=running", "--format", "json"}, config, nil)
	if err != nil {
		return nil, err
	}

	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	out, err := cmd.CombinedOutput()
	if err != nil {
		if ctx.Err() == context.Canceled {
			// return without error if context is cancelled
			return nil, nil
		}

		return nil, fmt.Errorf("failed to get running docker compose services: %w", err)
	}

	var svcs []svc
	err = json.Unmarshal(out, &svcs)
	if err != nil {
		return nil, fmt.Errorf("couldn't unmarshal json result from docker compose: %w", err)
	}

	// build service mapping
	for _, service := range svcs {
		runningServices[service.Service] = service
	}

	return runningServices, nil
}

func (p Ports) getRandomPort(ctx context.Context, port uint32, timeout time.Duration) (uint32, error) {
	t := time.After(timeout)

	for {
		select {
		case <-ctx.Done():
			return 0, nil
		case <-t:
			return 0, fmt.Errorf("timed out on ports randomize")
		default:
			if port > highestTcpPort {
				return 0, fmt.Errorf("port %d is higher than the highest possible TCP port", port)
			}

			if util.PortAvailable(port) {
				return port, nil
			}

			port++
		}
	}
}
