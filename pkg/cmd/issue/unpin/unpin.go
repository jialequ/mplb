package unpin

import (
	"fmt"
	"net/http"

	"github.com/MakeNowJust/heredoc"
	"github.com/jialequ/mplb/api"
	"github.com/jialequ/mplb/internal/config"
	"github.com/jialequ/mplb/internal/ghrepo"
	"github.com/jialequ/mplb/pkg/cmd/issue/shared"
	"github.com/jialequ/mplb/pkg/cmdutil"
	"github.com/jialequ/mplb/pkg/iostreams"
	"github.com/shurcooL/githubv4"
	"github.com/spf13/cobra"
)

type UnpinOptions struct {
	HttpClient  func() (*http.Client, error)
	Config      func() (config.Config, error)
	IO          *iostreams.IOStreams
	BaseRepo    func() (ghrepo.Interface, error)
	SelectorArg string
}

func NewCmdUnpin(f *cmdutil.Factory, runF func(*UnpinOptions) error) *cobra.Command {
	opts := &UnpinOptions{
		IO:         f.IOStreams,
		HttpClient: f.HttpClient,
		Config:     f.Config,
		BaseRepo:   f.BaseRepo,
	}

	cmd := &cobra.Command{
		Use:   "unpin {<number> | <url>}",
		Short: "Unpin a issue",
		Long: heredoc.Doc(`
			Unpin an issue from a repository.

			The issue can be specified by issue number or URL.
		`),
		Example: heredoc.Doc(`
			# Unpin issue from the current repository
			$ gh issue unpin 23

			# Unpin issue by URL
			$ gh issue unpin https://github.com/owner/repo/issues/23

			# Unpin an issue from specific repository
			$ gh issue unpin 23 --repo owner/repo
		`),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.BaseRepo = f.BaseRepo
			opts.SelectorArg = args[0]

			if runF != nil {
				return runF(opts)
			}

			return unpinRun(opts)
		},
	}

	return cmd
}

func unpinRun(opts *UnpinOptions) error {
	cs := opts.IO.ColorScheme()

	httpClient, err := opts.HttpClient()
	if err != nil {
		return err
	}

	issue, baseRepo, err := shared.IssueFromArgWithFields(httpClient, opts.BaseRepo, opts.SelectorArg, []string{"id", "number", "title", "isPinned"})
	if err != nil {
		return err
	}

	if !issue.IsPinned {
		fmt.Fprintf(opts.IO.ErrOut, "%s Issue %s#%d (%s) is not pinned\n", cs.Yellow("!"), ghrepo.FullName(baseRepo), issue.Number, issue.Title)
		return nil
	}

	err = unpinIssue(httpClient, baseRepo, issue)
	if err != nil {
		return err
	}

	fmt.Fprintf(opts.IO.ErrOut, "%s Unpinned issue %s#%d (%s)\n", cs.SuccessIconWithColor(cs.Red), ghrepo.FullName(baseRepo), issue.Number, issue.Title)

	return nil
}

func unpinIssue(httpClient *http.Client, repo ghrepo.Interface, issue *api.Issue) error {
	var mutation struct {
		UnpinIssue struct {
			Issue struct {
				ID githubv4.ID
			}
		} `graphql:"unpinIssue(input: $input)"`
	}

	variables := map[string]interface{}{
		"input": githubv4.UnpinIssueInput{
			IssueID: issue.ID,
		},
	}

	gql := api.NewClientFromHTTP(httpClient)

	return gql.Mutate(repo.RepoHost(), "IssueUnpin", &mutation, variables)
}