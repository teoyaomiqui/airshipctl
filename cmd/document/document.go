package document

import (
	"github.com/spf13/cobra"

	"opendev.org/airship/airshipctl/cmd/document/secret"
	"opendev.org/airship/airshipctl/pkg/environment"
)

// NewDocumentCommand creates a new command for managing airshipctl documents
func NewDocumentCommand(rootSettings *environment.AirshipCTLSettings) *cobra.Command {
	documentRootCmd := &cobra.Command{
		Use:   "document",
		Short: "manages deployment documents",
	}

	documentRootCmd.AddCommand(secret.NewSecretCommand(rootSettings))
	documentRootCmd.AddCommand(NewDocumentPullCommand(rootSettings))

	return documentRootCmd
}
