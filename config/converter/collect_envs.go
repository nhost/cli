package converter

import (
	"fmt"
	"github.com/nhost/cli/nhost/envvars"
	"reflect"
	"strings"
)

func collectLegacyConfigServiceEnv(legacyEnv map[string]any) envvars.Env {
	env := envvars.New()
	for k, v := range legacyEnv {
		env[strings.ToUpper(k)] = fmt.Sprint(v)
	}
	return env
}

func flattenEnvs(data any, prefix string) envvars.Env {
	envs := envvars.New()

	if data == nil {
		return envs
	}

	value := reflect.ValueOf(data)

	if value.Kind() == reflect.Map {
		for _, mkey := range value.MapKeys() {
			var key string

			if prefix == "" {
				key = strings.ToUpper(fmt.Sprint(mkey))
			} else {
				key = strings.ToUpper(fmt.Sprintf("%s_%v", prefix, mkey))
			}

			switch value.MapIndex(mkey).Interface().(type) {
			case map[interface{}]interface{}, map[string]interface{}:
				envs.Merge(flattenEnvs(value.MapIndex(mkey).Interface(), key))
			default:
				envs[key] = fmt.Sprintf("%v", value.MapIndex(mkey).Interface())
			}
		}
	}

	return envs
}
