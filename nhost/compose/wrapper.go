package compose

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"io"
	"os"
	"os/exec"
)

type DataStreams struct {
	Stdout io.Writer
	Stderr io.Writer
}

func WrapperCmd(ctx context.Context, args []string, conf *Config, streams DataStreams) (*exec.Cmd, error) {
	dockerComposeConfig, err := conf.BuildJSON()
	if err != nil {
		return nil, err
	}

	composeConfigFilename := "docker-compose.json"

	// write data to a docker-compose.yml file
	err = os.WriteFile(composeConfigFilename, dockerComposeConfig, 0644)
	if err != nil {
		return nil, errors.Wrap(err, "could not write docker-compose.yml file")
	}

	// check that data folders exist
	hostDataFolders, err := conf.HostMountedDataPaths()
	if err != nil {
		return nil, err
	}

	for _, folder := range hostDataFolders {
		err = os.MkdirAll(folder, os.ModePerm)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("failed to create data folder '%s'", folder))
		}
	}

	dc := exec.CommandContext(ctx, "docker", append([]string{"compose", "-f", composeConfigFilename}, args...)...)

	// set streams
	dc.Stdout = streams.Stdout
	dc.Stderr = streams.Stderr
	dc.Stdin = os.Stdin
	//dc.Stdin = bytes.NewReader(Config)

	return dc, nil
}
