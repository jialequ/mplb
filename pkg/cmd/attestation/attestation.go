package attestation

import (
	"github.com/MakeNowJust/heredoc"
	"github.com/jialequ/mplb/pkg/cmd/attestation/download"
	"github.com/jialequ/mplb/pkg/cmd/attestation/inspect"
	"github.com/jialequ/mplb/pkg/cmd/attestation/tufrootverify"
	"github.com/jialequ/mplb/pkg/cmd/attestation/verify"
	"github.com/jialequ/mplb/pkg/cmdutil"

	"github.com/spf13/cobra"
)

func NewCmdAttestation(f *cmdutil.Factory) *cobra.Command {
	root := &cobra.Command{
		Use:     "attestation [subcommand]",
		Short:   "Work with artifact attestations",
		Aliases: []string{"at"},
		Long: heredoc.Doc(`
			### NOTE: This feature is currently in beta, and subject to change.

			Download and verify artifact attestations.
			`),
	}

	root.AddCommand(download.NewDownloadCmd(f, nil))
	root.AddCommand(inspect.NewInspectCmd(f, nil))
	root.AddCommand(verify.NewVerifyCmd(f, nil))
	root.AddCommand(tufrootverify.NewTUFRootVerifyCmd(f, nil))

	return root
}
