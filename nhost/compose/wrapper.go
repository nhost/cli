package compose

import (
	"context"
	"fmt"
	"github.com/nhost/cli/util"
	"github.com/pkg/errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
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

	composeConfigFilename := filepath.Join(util.WORKING_DIR, ".nhost/docker-compose.json")

	// write data to a docker-compose.yml file
	err = os.WriteFile(composeConfigFilename, dockerComposeConfig, 0644)
	if err != nil {
		return nil, errors.Wrap(err, "could not write docker-compose.yml file")
	}

	// check that data folders exist
	paths := []string{
		filepath.Join(util.WORKING_DIR, ".nhost/data/minio"),
		filepath.Join(util.WORKING_DIR, ".nhost/data/mailhog"),
		filepath.Join(util.WORKING_DIR, ".nhost/custom/keys"),
		filepath.Join(util.WORKING_DIR, ".nhost/data/db", conf.gitBranch),
	}

	for _, folder := range paths {
		err = os.MkdirAll(folder, os.ModePerm)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("failed to create data folder '%s'", folder))
		}
	}

	dc := exec.CommandContext(ctx, "docker", append([]string{"compose", "-p", conf.composeProjectName, "-f", composeConfigFilename}, args...)...)

	// set streams
	dc.Stdout = streams.Stdout
	dc.Stderr = streams.Stderr
	dc.Stdin = os.Stdin

	return dc, nil
}
