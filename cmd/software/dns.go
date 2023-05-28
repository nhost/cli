package software

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/nhost/cli/clienv"
	"github.com/nhost/cli/system"
	"github.com/urfave/cli/v2"
)

func CommandDNS() *cli.Command {
	return &cli.Command{ //nolint:exhaustruct
		Name:    "fix-dns",
		Aliases: []string{},
		Usage:   "Add DNS entries to /etc/hosts to avoid issues when using the cli without internet connection",
		Action:  commandDNS,
	}
}

func commandDNS(cCtx *cli.Context) error {
	ce := clienv.New(cCtx)

	etcHosts := filepath.Join("/", "etc", "hosts")
	f, err := os.OpenFile(etcHosts, os.O_APPEND|os.O_RDWR, 0o644) //nolint:gomnd
	if os.IsPermission(err) {
		ce.Warnln("Permission denied. Please `sudo nhost sw fix-dns`")
	}
	if err != nil {
		return fmt.Errorf("failed to open %s: %w", etcHosts, err)
	}
	defer f.Close()

	if system.DNSPresent(f) {
		ce.Infoln("DNS already configured")
		return nil
	}

	if err := system.DNSAdd(f); err != nil {
		return fmt.Errorf("failed to add dns configuration: %w", err)
	}
	ce.Infoln("DNS updated successfully")

	return nil
}
