package cache

import (
	"github.com/MakeNowJust/heredoc"
	cmdDelete "github.com/jialequ/mplb/pkg/cmd/cache/delete"
	cmdList "github.com/jialequ/mplb/pkg/cmd/cache/list"
	"github.com/jialequ/mplb/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func NewCmdCache(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cache <command>",
		Short: "Manage Github Actions caches",
		Long:  "Work with Github Actions caches.",
		Example: heredoc.Doc(`
			$ gh cache list
			$ gh cache delete --all
		`),
		GroupID: "actions",
	}

	cmdutil.EnableRepoOverride(cmd, f)

	cmd.AddCommand(cmdList.NewCmdList(f, nil))
	cmd.AddCommand(cmdDelete.NewCmdDelete(f, nil))

	return cmd
}
