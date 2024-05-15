package software

import (
	"context"
	"fmt"
	"runtime"
	"strings"

	"github.com/nhost/be/services/mimir/model"
	"github.com/nhost/cli/clienv"
	"github.com/nhost/cli/cmd/config"
	"github.com/nhost/cli/nhostclient/graphql"
	"github.com/nhost/cli/project/env"
	"github.com/nhost/cli/software"
	"github.com/urfave/cli/v2"
)

func CommandVersion() *cli.Command {
	return &cli.Command{ //nolint:exhaustruct
		Name:    "version",
		Aliases: []string{},
		Usage:   "Show the current version of Nhost CLI you have installed",
		Action:  commandVersion,
	}
}

func checkCLIVersion(
	ctx context.Context,
	ce *clienv.CliEnv,
	curVersion string,
) error {
	mgr := software.NewManager()
	releases, err := mgr.GetReleases(ctx, curVersion)
	if err != nil {
		return fmt.Errorf("failed to get releases: %w", err)
	}

	if len(releases) == 0 {
		ce.Infoln(
			"✅ Nhost CLI %s for %s-%s is already on the latest version",
			curVersion, runtime.GOOS, runtime.GOARCH,
		)
		return nil
	}

	latest := releases[0]
	if latest.TagName == curVersion {
		return nil
	}

	ce.Warnln("🟡 A new version of Nhost CLI is available: %s", latest.TagName)
	ce.Println("   You can upgrade the CLI by running `nhost sw upgrade`")
	ce.Println("   For details on what's new, visit https://github.com/nhost/cli/releases")

	return nil
}

func checkServiceVersion(
	ce *clienv.CliEnv,
	software graphql.SoftwareTypeEnum,
	curVersion string,
	availableVersions *graphql.GetSoftwareVersions,
	changelog string,
) {
	recommendedVersions := make([]string, 0, 5) //nolint:gomnd
	for _, v := range availableVersions.GetSoftwareVersions() {
		if *v.GetSoftware() == software && v.GetVersion() == curVersion {
			ce.Infoln("✅ %s is already on a recommended version: %s", software, curVersion)
			return
		} else if *v.GetSoftware() == software {
			recommendedVersions = append(recommendedVersions, v.GetVersion())
		}
	}
	ce.Warnln("🟡 %s is not on a recommended version: %s", software, curVersion)
	ce.Println("   Recommended version(s) are: %s", strings.Join(recommendedVersions, ", "))
	if changelog != "" {
		ce.Println("   For details on what's new, visit %s", changelog)
	}
}

func CheckVersions(
	ctx context.Context,
	ce *clienv.CliEnv,
	cfg *model.ConfigConfig,
	appVersion string,
) error {
	var secrets model.Secrets
	if err := clienv.UnmarshalFile(ce.Path.Secrets(), &secrets, env.Unmarshal); err != nil {
		return fmt.Errorf(
			"failed to parse secrets, make sure secret values are between quotes: %w",
			err,
		)
	}

	cl, err := ce.GetNhostClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to get nhost client: %w", err)
	}

	swv, err := cl.GetSoftwareVersions(ctx)
	if err != nil {
		return fmt.Errorf("failed to get software versions: %w", err)
	}

	checkServiceVersion(
		ce, graphql.SoftwareTypeEnumAuth, *cfg.GetAuth().GetVersion(), swv,
		"https://github.com/nhost/hasura-auth/releases",
	)
	checkServiceVersion(
		ce, graphql.SoftwareTypeEnumStorage, *cfg.GetStorage().GetVersion(), swv,
		"https://github.com/nhost/hasura-storage/releases",
	)
	checkServiceVersion(
		ce, graphql.SoftwareTypeEnumPostgreSQL, *cfg.GetPostgres().GetVersion(), swv,
		"https://hub.docker.com/r/nhost/postgres",
	)
	checkServiceVersion(ce, graphql.SoftwareTypeEnumHasura, *cfg.GetHasura().GetVersion(), swv, "")

	if cfg.GetAi() != nil {
		checkServiceVersion(ce, graphql.SoftwareTypeEnumGraphite, *cfg.GetAi().GetVersion(), swv, "")
	}

	return checkCLIVersion(ctx, ce, appVersion)
}

func CheckVersions2(
	ce *clienv.CliEnv,
) {
	check := clienv.Column{
		Header: "",
		Rows:   []string{"🟡", "✅", "🟡", "✅", "🟡", "🟡"},
	}
	num := clienv.Column{
		Header: "Software",
		Rows:   []string{"Nhost CLI", "Hasura", "Auth", "Storage", "Postgres", "Graphite"},
	}
	subdomain := clienv.Column{
		Header: "Running Version",
		Rows:   []string{"v1.16.1", "v2.33.4-ce", "0.21.2", "0.4.0", "14.6-20231218-1", "0.3.1"},
	}
	project := clienv.Column{
		Header: "Recommended Versions",
		Rows:   []string{"v1.16.4", "v2.38.0-ce", "0.24.1, 0.25.0, 0.26.0, 0.28.1, 0.29.4", "0.6.0", "14.11-20240429-1, 15.2-20240429-1, 16.2-20240429-1", "0.5.2"},
	}
	changelog := clienv.Column{
		Header: "Changelog",
		Rows: []string{
			"https://github.com/nhost/cli/releases",
			"",
			"https://github.com/nhost/hasura-auth/releases",
			"https://github.com/nhost/hasura-storage/releases",
			"https://hub.docker.com/r/nhost/postgres",
			"",
		},
	}

	ce.Println(clienv.Table(check, num, subdomain, project, changelog))
}

func commandVersion(cCtx *cli.Context) error {
	ce := clienv.FromCLI(cCtx)

	var cfg *model.ConfigConfig
	var err error
	if clienv.PathExists(ce.Path.NhostToml()) && clienv.PathExists(ce.Path.Secrets()) {
		var secrets model.Secrets
		if err := clienv.UnmarshalFile(ce.Path.Secrets(), &secrets, env.Unmarshal); err != nil {
			return fmt.Errorf(
				"failed to parse secrets, make sure secret values are between quotes: %w",
				err,
			)
		}

		cfg, err = config.Validate(ce, "local", secrets)
		if err != nil {
			return fmt.Errorf("failed to validate config: %w", err)
		}
	} else {
		ce.Warnln("🟡 No Nhost project found")
	}

	CheckVersions2(ce)

	return CheckVersions(cCtx.Context, ce, cfg, cCtx.App.Version)
}
