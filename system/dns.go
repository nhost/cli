package system

import (
	"fmt"
	"io"
	"strings"
)

const dnsLine = `# Start of Nhost CLI configuration
127.0.0.1 local.auth.nhost.run
127.0.0.1 local.db.nhost.run
127.0.0.1 local.functions.nhost.run
127.0.0.1 local.graphql.nhost.run
127.0.0.1 local.hasura.nhost.run
127.0.0.1 local.storage.nhost.run
# End of Nhost CLI configuration
`

func DNSPresent(r io.Reader) bool {
	b, err := io.ReadAll(r)
	if err != nil {
		return false
	}

	if strings.Contains(string(b), dnsLine) {
		return true
	}

	return false
}

func DNSAdd(w io.Writer) error {
	if _, err := w.Write([]byte(dnsLine)); err != nil {
		return fmt.Errorf("failed to write: %w", err)
	}

	return nil
}
