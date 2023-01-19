package nhost

import (
	"fmt"
	"github.com/nhost/be/services/mimir/model"
	"github.com/nhost/be/services/mimir/schema"
	"github.com/nhost/cli/config"
	"github.com/pelletier/go-toml/v2"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/docker/docker/pkg/namesgenerator"
	"github.com/nhost/cli/util"
)

const (
	projectNameFile = "project_name"
)

var (
	projectNameIgnoreRegex = regexp.MustCompile(`([^a-z0-9-_])+`)
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

func GetDockerComposeProjectName() (string, error) {
	projectNameFilename := filepath.Join(DOT_NHOST_DIR, projectNameFile)

	data, err := ioutil.ReadFile(projectNameFilename)
	if err != nil {
		return "", fmt.Errorf("can't read file '%s' %v", projectNameFile, err)
	}

	return strings.TrimSpace(string(data)), nil
}

func GetConfiguration() (*config.Config, error) {
	var c model.ConfigConfig

	data, err := os.ReadFile(CONFIG_PATH)
	if err != nil {
		return nil, err
	}

	err = toml.Unmarshal(data, &c)
	if err != nil {
		return nil, err
	}

	s, err := schema.New()
	if err != nil {
		return nil, err
	}

	if err = s.ValidateConfig(c); err != nil {
		return nil, err
	}

	return config.NewConfig(&c), nil
}

func EnsureProjectNameFileExists() error {
	projectNameFilename := filepath.Join(DOT_NHOST_DIR, projectNameFile)

	if !util.PathExists(projectNameFilename) {
		randomName := randomProjectName(filepath.Base(util.WORKING_DIR))

		if err := os.MkdirAll(DOT_NHOST_DIR, os.ModePerm); err != nil {
			return err
		}

		return ioutil.WriteFile(projectNameFilename, []byte(randomName), 0600)
	}

	return nil
}

func randomProjectName(base string) string {
	base = strings.ToLower(base)
	base = strings.TrimLeft(base, "_")
	base = strings.TrimRight(base, "_")
	base = projectNameIgnoreRegex.ReplaceAllString(base, "-")
	base = strings.TrimSuffix(base, "-")

	rand.Seed(time.Now().UnixNano())
	return strings.ToLower(strings.Join([]string{base, namesgenerator.GetRandomName(0)}, "-"))
}
