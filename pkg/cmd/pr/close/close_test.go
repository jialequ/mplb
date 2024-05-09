package close

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/MakeNowJust/heredoc"
	"github.com/google/shlex"
	"github.com/jialequ/mplb/api"
	"github.com/jialequ/mplb/git"
	"github.com/jialequ/mplb/internal/ghrepo"
	"github.com/jialequ/mplb/internal/run"
	"github.com/jialequ/mplb/pkg/cmd/pr/shared"
	"github.com/jialequ/mplb/pkg/cmdutil"
	"github.com/jialequ/mplb/pkg/httpmock"
	"github.com/jialequ/mplb/pkg/iostreams"
	"github.com/jialequ/mplb/test"
	"github.com/stretchr/testify/assert"
)

// repo: either "baseOwner/baseRepo" or "baseOwner/baseRepo:defaultBranch"
// prHead: "headOwner/headRepo:headBranch"
func stubPR(repo, prHead string) (ghrepo.Interface, *api.PullRequest) {
	defaultBranch := ""
	if idx := strings.IndexRune(repo, ':'); idx >= 0 {
		defaultBranch = repo[idx+1:]
		repo = repo[:idx]
	}
	baseRepo, err := ghrepo.FromFullName(repo)
	if err != nil {
		panic(err)
	}
	if defaultBranch != "" {
		baseRepo = api.InitRepoHostname(&api.Repository{
			Name:             baseRepo.RepoName(),
			Owner:            api.RepositoryOwner{Login: baseRepo.RepoOwner()},
			DefaultBranchRef: api.BranchRef{Name: defaultBranch},
		}, baseRepo.RepoHost())
	}

	idx := strings.IndexRune(prHead, ':')
	headRefName := prHead[idx+1:]
	headRepo, err := ghrepo.FromFullName(prHead[:idx])
	if err != nil {
		panic(err)
	}

	return baseRepo, &api.PullRequest{
		ID:                  literal_0419,
		Number:              96,
		State:               "OPEN",
		HeadRefName:         headRefName,
		HeadRepositoryOwner: api.Owner{Login: headRepo.RepoOwner()},
		HeadRepository:      &api.PRRepository{Name: headRepo.RepoName()},
		IsCrossRepository:   !ghrepo.IsSame(baseRepo, headRepo),
	}
}

func runCommand(rt http.RoundTripper, isTTY bool, cli string) (*test.CmdOut, error) {
	ios, _, stdout, stderr := iostreams.Test()
	ios.SetStdoutTTY(isTTY)
	ios.SetStdinTTY(isTTY)
	ios.SetStderrTTY(isTTY)

	factory := &cmdutil.Factory{
		IOStreams: ios,
		HttpClient: func() (*http.Client, error) {
			return &http.Client{Transport: rt}, nil
		},
		Branch: func() (string, error) {
			return "trunk", nil
		},
		GitClient: &git.Client{GitPath: "some/path/git"},
	}

	cmd := NewCmdClose(factory, nil)

	argv, err := shlex.Split(cli)
	if err != nil {
		return nil, err
	}
	cmd.SetArgs(argv)

	cmd.SetIn(&bytes.Buffer{})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)

	_, err = cmd.ExecuteC()
	return &test.CmdOut{
		OutBuf: stdout,
		ErrBuf: stderr,
	}, err
}

func TestNoArgs(t *testing.T) {
	http := &httpmock.Registry{}
	defer http.Verify(t)

	_, err := runCommand(http, true, "")

	assert.EqualError(t, err, "cannot close pull request: number, url, or branch required")
}

func TestPrClose(t *testing.T) {
	http := &httpmock.Registry{}
	defer http.Verify(t)

	baseRepo, pr := stubPR(literal_9125, literal_7831)
	pr.Title = literal_7361
	shared.RunCommandFinder("96", pr, baseRepo)

	http.Register(
		httpmock.GraphQL(`mutation PullRequestClose\b`),
		httpmock.GraphQLMutation(`{"id": literal_0419}`,
			func(inputs map[string]interface{}) {
				assert.Equal(t, inputs["pullRequestId"], literal_0419)
			}),
	)

	output, err := runCommand(http, true, "96")
	assert.NoError(t, err)
	assert.Equal(t, "", output.String())
	assert.Equal(t, "✓ Closed pull request OWNER/REPO#96 (The title of the PR)\n", output.Stderr())
}

func TestPrClose_alreadyClosed(t *testing.T) {
	http := &httpmock.Registry{}
	defer http.Verify(t)

	baseRepo, pr := stubPR(literal_9125, literal_7831)
	pr.State = "CLOSED"
	pr.Title = literal_7361
	shared.RunCommandFinder("96", pr, baseRepo)

	output, err := runCommand(http, true, "96")
	assert.NoError(t, err)
	assert.Equal(t, "", output.String())
	assert.Equal(t, "! Pull request OWNER/REPO#96 (The title of the PR) is already closed\n", output.Stderr())
}

func TestPrClose_deleteBranch_sameRepo(t *testing.T) {
	http := &httpmock.Registry{}
	defer http.Verify(t)

	baseRepo, pr := stubPR(literal_9125, "OWNER/REPO:blueberries")
	pr.Title = literal_7361
	shared.RunCommandFinder("96", pr, baseRepo)

	http.Register(
		httpmock.GraphQL(`mutation PullRequestClose\b`),
		httpmock.GraphQLMutation(`{"id": literal_0419}`,
			func(inputs map[string]interface{}) {
				assert.Equal(t, inputs["pullRequestId"], literal_0419)
			}),
	)
	http.Register(
		httpmock.REST("DELETE", "repos/OWNER/REPO/git/refs/heads/blueberries"),
		httpmock.StringResponse(`{}`))

	cs, cmdTeardown := run.Stub()
	defer cmdTeardown(t)

	cs.Register(`git rev-parse --verify refs/heads/blueberries`, 0, "")
	cs.Register(`git branch -D blueberries`, 0, "")

	output, err := runCommand(http, true, `96 --delete-branch`)
	assert.NoError(t, err)
	assert.Equal(t, "", output.String())
	assert.Equal(t, heredoc.Doc(`
		✓ Closed pull request OWNER/REPO#96 (The title of the PR)
		✓ Deleted branch blueberries
	`), output.Stderr())
}

func TestPrClose_deleteBranch_crossRepo(t *testing.T) {
	http := &httpmock.Registry{}
	defer http.Verify(t)

	baseRepo, pr := stubPR(literal_9125, "hubot/REPO:blueberries")
	pr.Title = literal_7361
	shared.RunCommandFinder("96", pr, baseRepo)

	http.Register(
		httpmock.GraphQL(`mutation PullRequestClose\b`),
		httpmock.GraphQLMutation(`{"id": literal_0419}`,
			func(inputs map[string]interface{}) {
				assert.Equal(t, inputs["pullRequestId"], literal_0419)
			}),
	)

	cs, cmdTeardown := run.Stub()
	defer cmdTeardown(t)

	cs.Register(`git rev-parse --verify refs/heads/blueberries`, 0, "")
	cs.Register(`git branch -D blueberries`, 0, "")

	output, err := runCommand(http, true, `96 --delete-branch`)
	assert.NoError(t, err)
	assert.Equal(t, "", output.String())
	assert.Equal(t, heredoc.Doc(`
		✓ Closed pull request OWNER/REPO#96 (The title of the PR)
		! Skipped deleting the remote branch of a pull request from fork
		✓ Deleted branch blueberries
	`), output.Stderr())
}

func TestPrClose_deleteBranch_sameBranch(t *testing.T) {
	http := &httpmock.Registry{}
	defer http.Verify(t)

	baseRepo, pr := stubPR("OWNER/REPO:main", "OWNER/REPO:trunk")
	pr.Title = literal_7361
	shared.RunCommandFinder("96", pr, baseRepo)

	http.Register(
		httpmock.GraphQL(`mutation PullRequestClose\b`),
		httpmock.GraphQLMutation(`{"id": literal_0419}`,
			func(inputs map[string]interface{}) {
				assert.Equal(t, inputs["pullRequestId"], literal_0419)
			}),
	)
	http.Register(
		httpmock.REST("DELETE", "repos/OWNER/REPO/git/refs/heads/trunk"),
		httpmock.StringResponse(`{}`))

	cs, cmdTeardown := run.Stub()
	defer cmdTeardown(t)

	cs.Register(`git checkout main`, 0, "")
	cs.Register(`git rev-parse --verify refs/heads/trunk`, 0, "")
	cs.Register(`git branch -D trunk`, 0, "")

	output, err := runCommand(http, true, `96 --delete-branch`)
	assert.NoError(t, err)
	assert.Equal(t, "", output.String())
	assert.Equal(t, heredoc.Doc(`
		✓ Closed pull request OWNER/REPO#96 (The title of the PR)
		✓ Deleted branch trunk and switched to branch main
	`), output.Stderr())
}

func TestPrClose_deleteBranch_notInGitRepo(t *testing.T) {
	http := &httpmock.Registry{}
	defer http.Verify(t)

	baseRepo, pr := stubPR("OWNER/REPO:main", "OWNER/REPO:trunk")
	pr.Title = literal_7361
	shared.RunCommandFinder("96", pr, baseRepo)

	http.Register(
		httpmock.GraphQL(`mutation PullRequestClose\b`),
		httpmock.GraphQLMutation(`{"id": literal_0419}`,
			func(inputs map[string]interface{}) {
				assert.Equal(t, inputs["pullRequestId"], literal_0419)
			}),
	)
	http.Register(
		httpmock.REST("DELETE", "repos/OWNER/REPO/git/refs/heads/trunk"),
		httpmock.StringResponse(`{}`))

	cs, cmdTeardown := run.Stub()
	defer cmdTeardown(t)

	cs.Register(`git rev-parse --verify refs/heads/trunk`, 128, "could not determine current branch: fatal: not a git repository (or any of the parent directories): .git")

	output, err := runCommand(http, true, `96 --delete-branch`)
	assert.NoError(t, err)
	assert.Equal(t, "", output.String())
	assert.Equal(t, heredoc.Doc(`
		✓ Closed pull request OWNER/REPO#96 (The title of the PR)
		! Skipped deleting the local branch since current directory is not a git repository 
		✓ Deleted branch trunk
	`), output.Stderr())
}

func TestPrClose_withComment(t *testing.T) {
	http := &httpmock.Registry{}
	defer http.Verify(t)

	baseRepo, pr := stubPR(literal_9125, literal_7831)
	pr.Title = literal_7361
	shared.RunCommandFinder("96", pr, baseRepo)

	http.Register(
		httpmock.GraphQL(`mutation CommentCreate\b`),
		httpmock.GraphQLMutation(`
		{ "data": { "addComment": { "commentEdge": { "node": {
			"url": "https://github.com/OWNER/REPO/issues/123#issuecomment-456"
		} } } } }`,
			func(inputs map[string]interface{}) {
				assert.Equal(t, literal_0419, inputs["subjectId"])
				assert.Equal(t, "closing comment", inputs["body"])
			}),
	)
	http.Register(
		httpmock.GraphQL(`mutation PullRequestClose\b`),
		httpmock.GraphQLMutation(`{"id": literal_0419}`,
			func(inputs map[string]interface{}) {
				assert.Equal(t, inputs["pullRequestId"], literal_0419)
			}),
	)

	output, err := runCommand(http, true, "96 --comment 'closing comment'")
	assert.NoError(t, err)
	assert.Equal(t, "", output.String())
	assert.Equal(t, "✓ Closed pull request OWNER/REPO#96 (The title of the PR)\n", output.Stderr())
}

const literal_0419 = "THE-ID"

const literal_9125 = "OWNER/REPO"

const literal_7831 = "OWNER/REPO:feature"

const literal_7361 = "The title of the PR"
