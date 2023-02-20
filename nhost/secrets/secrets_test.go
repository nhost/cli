package secrets_test

import (
	"github.com/nhost/be/services/mimir/model"
	"github.com/nhost/cli/nhost/secrets"
	"reflect"
	"testing"
)

func TestInterpolate(t *testing.T) {
	tests := []struct {
		name    string
		envs    []*model.ConfigEnvironmentVariable
		secrets []byte
		want    []*model.ConfigEnvironmentVariable
		wantErr bool
	}{
		{
			name: "test with valid secrets",
			secrets: []byte(`
BAR=baz
FOO=bar
`),
			envs: []*model.ConfigEnvironmentVariable{
				{
					Name:  "FOO",
					Value: "{{ secrets.FOO }}",
				},
				{
					Name:  "BAR",
					Value: "{{ secrets.BAR }}",
				},
				{
					Name:  "FOO_BAR",
					Value: "{{ secrets.FOO }}_{{ secrets.BAR }}",
				},
			},
			want: []*model.ConfigEnvironmentVariable{
				{
					Name:  "FOO",
					Value: "bar",
				},
				{
					Name:  "BAR",
					Value: "baz",
				},
				{
					Name:  "FOO_BAR",
					Value: "bar_baz",
				},
			},
			wantErr: false,
		},
		{
			name: "test with invalid secrets",
			envs: []*model.ConfigEnvironmentVariable{
				{
					Name:  "FOO",
					Value: "{{ secrets.FOO }}",
				},
			},
			secrets: []byte(`
BLA:
BAR=baz
FOO=bar
`),
			wantErr: true,
		},
		{
			name: "test with no secrets",
			envs: []*model.ConfigEnvironmentVariable{
				{
					Name:  "FOO",
					Value: "{{ secrets.FOO }}",
				},
			},
			secrets: []byte(`                              `),
			want: []*model.ConfigEnvironmentVariable{
				{
					Name:  "FOO",
					Value: "{{ secrets.FOO }}",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := secrets.Interpolate(tt.envs, tt.secrets)
			if (err != nil) != tt.wantErr {
				t.Errorf("Compile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Compile() got = %v, want %v", got, tt.want)
			}
		})
	}
}
