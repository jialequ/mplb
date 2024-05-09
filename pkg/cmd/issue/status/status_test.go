package status

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

	cmd := NewCmdStatus(factory, nil)

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

func TestIssueStatus(t *testing.T) {
	http := &httpmock.Registry{}
	defer http.Verify(t)

	http.Register(
		httpmock.GraphQL(`query UserCurrent\b`),
		httpmock.StringResponse(`{"data":{"viewer":{"login":"octocat"}}}`))
	http.Register(
		httpmock.GraphQL(`query IssueStatus\b`),
		httpmock.FileResponse("./fixtures/issueStatus.json"))

	output, err := runCommand(http, true, "")
	if err != nil {
		t.Errorf(literal_1572, err)
	}

	expectedIssues := []*regexp.Regexp{
		regexp.MustCompile(`(?m)8.*carrots.*about.*ago`),
		regexp.MustCompile(`(?m)9.*squash.*about.*ago`),
		regexp.MustCompile(`(?m)10.*broccoli.*about.*ago`),
		regexp.MustCompile(`(?m)11.*swiss chard.*about.*ago`),
	}

	for _, r := range expectedIssues {
		if !r.MatchString(output.String()) {
			t.Errorf("output did not match regexp /%s/\n> output\n%s\n", r, output)
			return
		}
	}
}

func TestIssueStatusblankSlate(t *testing.T) {
	http := &httpmock.Registry{}
	defer http.Verify(t)

	http.Register(
		httpmock.GraphQL(`query UserCurrent\b`),
		httpmock.StringResponse(`{"data":{"viewer":{"login":"octocat"}}}`))
	http.Register(
		httpmock.GraphQL(`query IssueStatus\b`),
		httpmock.StringResponse(`
		{ "data": { "repository": {
			"hasIssuesEnabled": true,
			"assigned": { "nodes": [] },
			"mentioned": { "nodes": [] },
			"authored": { "nodes": [] }
		} } }`))

	output, err := runCommand(http, true, "")
	if err != nil {
		t.Errorf(literal_1572, err)
	}

	expectedOutput := `
Relevant issues in OWNER/REPO

Issues assigned to you
  There are no issues assigned to you

Issues mentioning you
  There are no issues mentioning you

Issues opened by you
  There are no issues opened by you

`
	if output.String() != expectedOutput {
		t.Errorf("expected %q, got %q", expectedOutput, output)
	}
}

func TestIssueStatusdisabledIssues(t *testing.T) {
	http := &httpmock.Registry{}
	defer http.Verify(t)

	http.Register(
		httpmock.GraphQL(`query UserCurrent\b`),
		httpmock.StringResponse(`{"data":{"viewer":{"login":"octocat"}}}`))
	http.Register(
		httpmock.GraphQL(`query IssueStatus\b`),
		httpmock.StringResponse(`
		{ "data": { "repository": {
			"hasIssuesEnabled": false
		} } }`))

	_, err := runCommand(http, true, "")
	if err == nil || err.Error() != "the 'OWNER/REPO' repository has disabled issues" {
		t.Errorf(literal_1572, err)
	}
}

const literal_1572 = "error running command `issue status`: %v"
