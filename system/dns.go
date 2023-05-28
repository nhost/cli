package system

import (
	"fmt"
	"io"
	"strings"
)

//nolint:lll
const dnsLine = `# Start of Nhost CLI configuration
127.0.0.1 local.auth.nhost.run local.db.nhost.run local.functions.nhost.run local.graphql.nhost.run local.hasura.nhost.run local.storage.nhost.run
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
