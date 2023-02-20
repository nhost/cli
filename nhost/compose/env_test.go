package compose

import (
	"github.com/compose-spec/compose-go/types"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_env_merge(t *testing.T) {
	assert := assert.New(t)

	e := env{"A": "B"}
	e.merge(env{"C": "D", "D": "E"}, env{"FOO": "BAR", "D": "F"})
	assert.Equal(env{"A": "B", "C": "D", "D": "F", "FOO": "BAR"}, e)
}

func Test_env_dockerServiceConfigEnv(t *testing.T) {
	assert := assert.New(t)

	e := env{"A": "B", "C": "D"}
	assert.Equal(types.NewMappingWithEquals([]string{"A=B", "C=D"}), e.dockerServiceConfigEnv())
}
