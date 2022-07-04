package nhost

import (
	"fmt"
	"github.com/nhost/cli/util"
	"io/ioutil"
	"path/filepath"
	"strings"
)

func ParseEnvVarsFromConfig(payload map[interface{}]interface{}, prefix string) []string {
	var response []string
	for key, item := range payload {
		switch item := item.(type) {
		case map[interface{}]interface{}:
			response = append(response, ParseEnvVarsFromConfig(item, strings.ToUpper(strings.Join([]string{prefix, fmt.Sprint(key)}, "_")))...)
		case interface{}:
			if item != "" {
				response = append(response, fmt.Sprintf("%s_%v=%v", prefix, strings.ToUpper(fmt.Sprint(key)), item))
			}
		}
	}
	return response
}

func GetContainerName(name string) string {
	return fmt.Sprintf("%s_%s", PREFIX, name)
}

func GetCurrentBranch() string {

	log.Debug("Fetching local git branch")
	data, err := ioutil.ReadFile(filepath.Join(GIT_DIR, "HEAD"))
	if err != nil {
		return ""
	}
	payload := strings.Split(string(data), " ")
	return strings.TrimSpace(filepath.Base(payload[1]))
}

func GetHeadBranchRef(branch string) (string, error) {
	refPath := filepath.Join(GIT_DIR, "refs/heads", branch)
	if !util.PathExists(refPath) {
		return "", fmt.Errorf("Branch '%s' not found", branch)
	}

	data, err := ioutil.ReadFile(refPath)
	return strings.TrimSpace(string(data)), err
}

func GetRemoteBranchRef(branch string) (string, error) {
	// TODO: the "origin" remote is hardcoded here, make it configurable
	refPath := filepath.Join(GIT_DIR, "refs/remotes/origin", branch)
	if !util.PathExists(refPath) {
		return "", fmt.Errorf("Branch '%s' not found", branch)
	}

	data, err := ioutil.ReadFile(refPath)
	return strings.TrimSpace(string(data)), err
}
