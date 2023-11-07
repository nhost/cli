package appconfig

import (
	"strings"

	"golang.org/x/mod/semver"
)

func CompareVersions(a, b string) int {
	extractFromImage := func(s string) string {
		if strings.Contains(s, ":") {
			return strings.Split(s, ":")[1]
		}
		return s
	}

	addVPrefix := func(s string) string {
		if !strings.HasPrefix(s, "v") {
			return "v" + s
		}
		return s
	}

	a = addVPrefix(extractFromImage(a))
	b = addVPrefix(extractFromImage(b))

	return semver.Compare(a, b)
}
