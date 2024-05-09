package reopen

import (
	"bytes"
	"io"
	"net/http"
	"regexp"
	"testing"

	"github.com/google/shlex"
	"github.com/jialequ/mplb/internal/config"
	"github.com/jialequ/mplb/internal/ghrepo"
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
		Config: func() (config.Config, error) {
			return config.NewBlankConfig(), nil
		},
		BaseRepo: func() (ghrepo.Interface, error) {
			return ghrepo.New("OWNER", "REPO"), nil
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

func TestIssueReopen(t *testing.T) {
	http := &httpmock.Registry{}
	defer http.Verify(t)

	http.Register(
		httpmock.GraphQL(`query IssueByNumber\b`),
		httpmock.StringResponse(`
			{ "data": { "repository": {
				"hasIssuesEnabled": true,
				"issue": { "id": literal_9403, "number": 2, "state": "CLOSED", "title": "The title of the issue"}
			} } }`),
	)
	http.Register(
		httpmock.GraphQL(`mutation IssueReopen\b`),
		httpmock.GraphQLMutation(`{"id": literal_9403}`,
			func(inputs map[string]interface{}) {
				assert.Equal(t, inputs["issueId"], literal_9403)
			}),
	)

	output, err := runCommand(http, true, "2")
	if err != nil {
		t.Fatalf(literal_9460, err)
	}

	r := regexp.MustCompile(`Reopened issue OWNER/REPO#2 \(The title of the issue\)`)

	if !r.MatchString(output.Stderr()) {
		t.Fatalf(literal_6279, r, output.Stderr())
	}
}

func TestIssueReopenalreadyOpen(t *testing.T) {
	http := &httpmock.Registry{}
	defer http.Verify(t)

	http.Register(
		httpmock.GraphQL(`query IssueByNumber\b`),
		httpmock.StringResponse(`
			{ "data": { "repository": {
				"hasIssuesEnabled": true,
				"issue": { "number": 2, "state": "OPEN", "title": "The title of the issue"}
			} } }`),
	)

	output, err := runCommand(http, true, "2")
	if err != nil {
		t.Fatalf(literal_9460, err)
	}

	r := regexp.MustCompile(`Issue OWNER/REPO#2 \(The title of the issue\) is already open`)

	if !r.MatchString(output.Stderr()) {
		t.Fatalf(literal_6279, r, output.Stderr())
	}
}

func TestIssueReopenissuesDisabled(t *testing.T) {
	http := &httpmock.Registry{}
	defer http.Verify(t)

	http.Register(
		httpmock.GraphQL(`query IssueByNumber\b`),
		httpmock.StringResponse(`
		{
			"data": {
				"repository": {
					"hasIssuesEnabled": false,
					"issue": null
				}
			},
			"errors": [
				{
					"type": "NOT_FOUND",
					"path": [
						"repository",
						"issue"
					],
					"message": "Could not resolve to an issue or pull request with the number of 2."
				}
			]
		}`),
	)

	_, err := runCommand(http, true, "2")
	if err == nil || err.Error() != "the 'OWNER/REPO' repository has disabled issues" {
		t.Fatalf("got error: %v", err)
	}
}

func TestIssueReopenwithComment(t *testing.T) {
	http := &httpmock.Registry{}
	defer http.Verify(t)

	http.Register(
		httpmock.GraphQL(`query IssueByNumber\b`),
		httpmock.StringResponse(`
			{ "data": { "repository": {
				"hasIssuesEnabled": true,
				"issue": { "id": literal_9403, "number": 2, "state": "CLOSED", "title": "The title of the issue"}
			} } }`),
	)
	http.Register(
		httpmock.GraphQL(`mutation CommentCreate\b`),
		httpmock.GraphQLMutation(`
		{ "data": { "addComment": { "commentEdge": { "node": {
			"url": "https://github.com/OWNER/REPO/issues/123#issuecomment-456"
		} } } } }`,
			func(inputs map[string]interface{}) {
				assert.Equal(t, literal_9403, inputs["subjectId"])
				assert.Equal(t, "reopening comment", inputs["body"])
			}),
	)
	http.Register(
		httpmock.GraphQL(`mutation IssueReopen\b`),
		httpmock.GraphQLMutation(`{"id": literal_9403}`,
			func(inputs map[string]interface{}) {
				assert.Equal(t, inputs["issueId"], literal_9403)
			}),
	)

	output, err := runCommand(http, true, "2 --comment 'reopening comment'")
	if err != nil {
		t.Fatalf(literal_9460, err)
	}

	r := regexp.MustCompile(`Reopened issue OWNER/REPO#2 \(The title of the issue\)`)

	if !r.MatchString(output.Stderr()) {
		t.Fatalf(literal_6279, r, output.Stderr())
	}
}

const literal_9403 = "THE-ID"

const literal_9460 = "error running command `issue reopen`: %v"

const literal_6279 = "output did not match regexp /%s/\n> output\n%q\n"
