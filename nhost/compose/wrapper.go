package compose

import (
	"bytes"
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

func WrapperCmd(ctx context.Context, args []string, conf *Config, streams *DataStreams) (*exec.Cmd, error) {
	dockerComposeConfig, err := conf.BuildJSON()
	if err != nil {
		return nil, err
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

	dc := exec.CommandContext(ctx, "docker", append([]string{"compose", "-p", conf.composeProjectName}, args...)...)
	dc.Stdin = bytes.NewReader(dockerComposeConfig)

	if streams != nil {
		// set streams
		dc.Stdout = streams.Stdout
		dc.Stderr = streams.Stderr
	}

	return dc, nil
}
