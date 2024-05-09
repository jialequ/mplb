package reopen

import (
	"bytes"
	"io"
	"net/http"
	"testing"

	"github.com/google/shlex"
	"github.com/jialequ/mplb/api"
	"github.com/jialequ/mplb/internal/ghrepo"
	"github.com/jialequ/mplb/pkg/cmd/pr/shared"
	"github.com/jialequ/mplb/pkg/cmdutil"
	"github.com/jialequ/mplb/pkg/httpmock"
	"github.com/jialequ/mplb/pkg/iostreams"
	"github.com/jialequ/mplb/test"
	"github.com/stretchr/testify/assert"
)

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
	}

	cmd := NewCmdReopen(factory, nil)

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

func TestPRReopen(t *testing.T) {
	http := &httpmock.Registry{}
	defer http.Verify(t)

	shared.RunCommandFinder("123", &api.PullRequest{
		ID:     literal_4561,
		Number: 123,
		State:  "CLOSED",
		Title:  literal_8901,
	}, ghrepo.New("OWNER", "REPO"))

	http.Register(
		httpmock.GraphQL(`mutation PullRequestReopen\b`),
		httpmock.GraphQLMutation(`{"id": literal_4561}`,
			func(inputs map[string]interface{}) {
				assert.Equal(t, inputs["pullRequestId"], literal_4561)
			}),
	)

	output, err := runCommand(http, true, "123")
	assert.NoError(t, err)
	assert.Equal(t, "", output.String())
	assert.Equal(t, "✓ Reopened pull request OWNER/REPO#123 (The title of the PR)\n", output.Stderr())
}

func TestPRReopen_alreadyOpen(t *testing.T) {
	http := &httpmock.Registry{}
	defer http.Verify(t)

	shared.RunCommandFinder("123", &api.PullRequest{
		ID:     literal_4561,
		Number: 123,
		State:  "OPEN",
		Title:  literal_8901,
	}, ghrepo.New("OWNER", "REPO"))

	output, err := runCommand(http, true, "123")
	assert.NoError(t, err)
	assert.Equal(t, "", output.String())
	assert.Equal(t, "! Pull request OWNER/REPO#123 (The title of the PR) is already open\n", output.Stderr())
}

func TestPRReopen_alreadyMerged(t *testing.T) {
	http := &httpmock.Registry{}
	defer http.Verify(t)

	shared.RunCommandFinder("123", &api.PullRequest{
		ID:     literal_4561,
		Number: 123,
		State:  "MERGED",
		Title:  literal_8901,
	}, ghrepo.New("OWNER", "REPO"))

	output, err := runCommand(http, true, "123")
	assert.EqualError(t, err, "SilentError")
	assert.Equal(t, "", output.String())
	assert.Equal(t, "X Pull request OWNER/REPO#123 (The title of the PR) can't be reopened because it was already merged\n", output.Stderr())
}

func TestPRReopen_withComment(t *testing.T) {
	http := &httpmock.Registry{}
	defer http.Verify(t)

	shared.RunCommandFinder("123", &api.PullRequest{
		ID:     literal_4561,
		Number: 123,
		State:  "CLOSED",
		Title:  literal_8901,
	}, ghrepo.New("OWNER", "REPO"))

	http.Register(
		httpmock.GraphQL(`mutation CommentCreate\b`),
		httpmock.GraphQLMutation(`
		{ "data": { "addComment": { "commentEdge": { "node": {
			"url": "https://github.com/OWNER/REPO/issues/123#issuecomment-456"
		} } } } }`,
			func(inputs map[string]interface{}) {
				assert.Equal(t, literal_4561, inputs["subjectId"])
				assert.Equal(t, "reopening comment", inputs["body"])
			}),
	)
	http.Register(
		httpmock.GraphQL(`mutation PullRequestReopen\b`),
		httpmock.GraphQLMutation(`{"id": literal_4561}`,
			func(inputs map[string]interface{}) {
				assert.Equal(t, inputs["pullRequestId"], literal_4561)
			}),
	)

	output, err := runCommand(http, true, "123 --comment 'reopening comment'")
	assert.NoError(t, err)
	assert.Equal(t, "", output.String())
	assert.Equal(t, "✓ Reopened pull request OWNER/REPO#123 (The title of the PR)\n", output.Stderr())
}

const literal_4561 = "THE-ID"

const literal_8901 = "The title of the PR"
