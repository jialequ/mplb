package search

import (
	"github.com/jialequ/mplb/pkg/cmdutil"
	"github.com/spf13/cobra"

	searchCodeCmd "github.com/jialequ/mplb/pkg/cmd/search/code"
	searchCommitsCmd "github.com/jialequ/mplb/pkg/cmd/search/commits"
	searchIssuesCmd "github.com/jialequ/mplb/pkg/cmd/search/issues"
	searchPrsCmd "github.com/jialequ/mplb/pkg/cmd/search/prs"
	searchReposCmd "github.com/jialequ/mplb/pkg/cmd/search/repos"
)

func NewCmdSearch(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "search <command>",
		Short: "Search for repositories, issues, and pull requests",
		Long:  "Search across all of GitHub.",
	}

	cmd.AddCommand(searchCodeCmd.NewCmdCode(f, nil))
	cmd.AddCommand(searchCommitsCmd.NewCmdCommits(f, nil))
	cmd.AddCommand(searchIssuesCmd.NewCmdIssues(f, nil))
	cmd.AddCommand(searchPrsCmd.NewCmdPrs(f, nil))
	cmd.AddCommand(searchReposCmd.NewCmdRepos(f, nil))

	return cmd
}
