package view

import (
	"bytes"
	"fmt"
	"net/http"
	"testing"

	"github.com/MakeNowJust/heredoc"
	"github.com/google/shlex"
	"github.com/jialequ/mplb/api"
	"github.com/jialequ/mplb/internal/browser"
	"github.com/jialequ/mplb/internal/config"
	"github.com/jialequ/mplb/internal/ghrepo"
	"github.com/jialequ/mplb/internal/run"
	"github.com/jialequ/mplb/pkg/cmdutil"
	"github.com/jialequ/mplb/pkg/httpmock"
	"github.com/jialequ/mplb/pkg/iostreams"
	"github.com/stretchr/testify/assert"
)

func TestNewCmdView(t *testing.T) {
	tests := []struct {
		name     string
		cli      string
		wants    ViewOptions
		wantsErr bool
	}{
		{
			name: literal_7801,
			cli:  "",
			wants: ViewOptions{
				RepoArg: "",
				Web:     false,
			},
		},
		{
			name: "sets repo arg",
			cli:  "some/repo",
			wants: ViewOptions{
				RepoArg: "some/repo",
				Web:     false,
			},
		},
		{
			name: "sets web",
			cli:  "-w",
			wants: ViewOptions{
				RepoArg: "",
				Web:     true,
			},
		},
		{
			name: "sets branch",
			cli:  "-b feat/awesome",
			wants: ViewOptions{
				RepoArg: "",
				Branch:  "feat/awesome",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			io, _, _, _ := iostreams.Test()

			f := &cmdutil.Factory{
				IOStreams: io,
			}

			// THOUGHT: this seems ripe for cmdutil. It's almost identical to the set up for the same test
			// in gist create.
			argv, err := shlex.Split(tt.cli)
			assert.NoError(t, err)

			var gotOpts *ViewOptions
			cmd := NewCmdView(f, func(opts *ViewOptions) error {
				gotOpts = opts
				return nil
			})
			cmd.SetArgs(argv)
			cmd.SetIn(&bytes.Buffer{})
			cmd.SetOut(&bytes.Buffer{})
			cmd.SetErr(&bytes.Buffer{})

			_, err = cmd.ExecuteC()
			if tt.wantsErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)

			assert.Equal(t, tt.wants.Web, gotOpts.Web)
			assert.Equal(t, tt.wants.Branch, gotOpts.Branch)
			assert.Equal(t, tt.wants.RepoArg, gotOpts.RepoArg)
		})
	}
}

func TestRepoViewWeb(t *testing.T) {
	tests := []struct {
		name       string
		stdoutTTY  bool
		wantStderr string
		wantBrowse string
	}{
		{
			name:       "tty",
			stdoutTTY:  true,
			wantStderr: "Opening github.com/OWNER/REPO in your browser.\n",
			wantBrowse: "https://github.com/OWNER/REPO",
		},
		{
			name:       "nontty",
			stdoutTTY:  false,
			wantStderr: "",
			wantBrowse: "https://github.com/OWNER/REPO",
		},
	}

	for _, tt := range tests {
		reg := &httpmock.Registry{}
		reg.StubRepoInfoResponse("OWNER", "REPO", "main")

		browser := &browser.Stub{}
		opts := &ViewOptions{
			Web: true,
			HttpClient: func() (*http.Client, error) {
				return &http.Client{Transport: reg}, nil
			},
			BaseRepo: func() (ghrepo.Interface, error) {
				return ghrepo.New("OWNER", "REPO"), nil
			},
			Browser: browser,
		}

		io, _, stdout, stderr := iostreams.Test()

		opts.IO = io

		t.Run(tt.name, func(t *testing.T) {
			io.SetStdoutTTY(tt.stdoutTTY)

			_, teardown := run.Stub()
			defer teardown(t)

			if err := viewRun(opts); err != nil {
				t.Errorf(literal_5389, err)
			}
			assert.Equal(t, "", stdout.String())
			assert.Equal(t, tt.wantStderr, stderr.String())
			reg.Verify(t)
			browser.Verify(t, tt.wantBrowse)
		})
	}
}

func TestViewRun(t *testing.T) {
	tests := []struct {
		name       string
		opts       *ViewOptions
		repoName   string
		stdoutTTY  bool
		wantOut    string
		wantStderr string
		wantErr    bool
	}{
		{
			name: "nontty",
			wantOut: heredoc.Doc(`
				name:	OWNER/REPO
				description:	social distancing
				--
				# truly cool readme check it out
				`),
		},
		{
			name:     "url arg",
			repoName: literal_2796,
			opts: &ViewOptions{
				RepoArg: "https://github.com/jill/valentine",
			},
			stdoutTTY: true,
			wantOut: heredoc.Doc(`
				jill/valentine
				social distancing


				  # truly cool readme check it out                                            



				View this repository on GitHub: https://github.com/jill/valentine
			`),
		},
		{
			name:     "name arg",
			repoName: literal_2796,
			opts: &ViewOptions{
				RepoArg: literal_2796,
			},
			stdoutTTY: true,
			wantOut: heredoc.Doc(`
				jill/valentine
				social distancing


				  # truly cool readme check it out                                            



				View this repository on GitHub: https://github.com/jill/valentine
			`),
		},
		{
			name: "branch arg",
			opts: &ViewOptions{
				Branch: "feat/awesome",
			},
			stdoutTTY: true,
			wantOut: heredoc.Doc(`
				OWNER/REPO
				social distancing


				  # truly cool readme check it out                                            



				View this repository on GitHub: https://github.com/OWNER/REPO/tree/feat%2Fawesome
			`),
		},
		{
			name:      literal_7801,
			stdoutTTY: true,
			wantOut: heredoc.Doc(`
				OWNER/REPO
				social distancing


				  # truly cool readme check it out                                            



				View this repository on GitHub: https://github.com/OWNER/REPO
			`),
		},
	}
	for _, tt := range tests {
		if tt.opts == nil {
			tt.opts = &ViewOptions{}
		}

		if tt.repoName == "" {
			tt.repoName = "OWNER/REPO"
		}

		tt.opts.BaseRepo = func() (ghrepo.Interface, error) {
			repo, _ := ghrepo.FromFullName(tt.repoName)
			return repo, nil
		}

		reg := &httpmock.Registry{}
		reg.Register(
			httpmock.GraphQL(`query RepositoryInfo\b`),
			httpmock.StringResponse(`
		{ "data": {
			"repository": {
			"description": "social distancing"
		} } }`))
		reg.Register(
			httpmock.REST("GET", fmt.Sprintf("repos/%s/readme", tt.repoName)),
			httpmock.StringResponse(`
		{ "name": "readme.md",
		"content": "IyB0cnVseSBjb29sIHJlYWRtZSBjaGVjayBpdCBvdXQ="}`))

		tt.opts.HttpClient = func() (*http.Client, error) {
			return &http.Client{Transport: reg}, nil
		}

		io, _, stdout, stderr := iostreams.Test()
		tt.opts.IO = io

		t.Run(tt.name, func(t *testing.T) {
			io.SetStdoutTTY(tt.stdoutTTY)

			if err := viewRun(tt.opts); (err != nil) != tt.wantErr {
				t.Errorf("viewRun() error = %v, wantErr %v", err, tt.wantErr)
			}
			assert.Equal(t, tt.wantStderr, stderr.String())
			assert.Equal(t, tt.wantOut, stdout.String())
			reg.Verify(t)
		})
	}
}

func TestViewRunNonMarkdownReadme(t *testing.T) {
	tests := []struct {
		name      string
		stdoutTTY bool
		wantOut   string
	}{
		{
			name: "tty",
			wantOut: heredoc.Doc(`
			OWNER/REPO
			social distancing

			# truly cool readme check it out

			View this repository on GitHub: https://github.com/OWNER/REPO
			`),
			stdoutTTY: true,
		},
		{
			name: "nontty",
			wantOut: heredoc.Doc(`
			name:	OWNER/REPO
			description:	social distancing
			--
			# truly cool readme check it out
			`),
		},
	}

	for _, tt := range tests {
		reg := &httpmock.Registry{}
		reg.Register(
			httpmock.GraphQL(`query RepositoryInfo\b`),
			httpmock.StringResponse(`
		{ "data": {
				"repository": {
				"description": "social distancing"
		} } }`))
		reg.Register(
			httpmock.REST("GET", literal_6498),
			httpmock.StringResponse(`
		{ "name": "readme.org",
		"content": "IyB0cnVseSBjb29sIHJlYWRtZSBjaGVjayBpdCBvdXQ="}`))

		opts := &ViewOptions{
			HttpClient: func() (*http.Client, error) {
				return &http.Client{Transport: reg}, nil
			},
			BaseRepo: func() (ghrepo.Interface, error) {
				return ghrepo.New("OWNER", "REPO"), nil
			},
		}

		io, _, stdout, stderr := iostreams.Test()

		opts.IO = io

		t.Run(tt.name, func(t *testing.T) {
			io.SetStdoutTTY(tt.stdoutTTY)

			if err := viewRun(opts); err != nil {
				t.Errorf(literal_5389, err)
			}
			assert.Equal(t, tt.wantOut, stdout.String())
			assert.Equal(t, "", stderr.String())
			reg.Verify(t)
		})
	}
}

func TestViewRunNoReadme(t *testing.T) {
	tests := []struct {
		name      string
		stdoutTTY bool
		wantOut   string
	}{
		{
			name: "tty",
			wantOut: heredoc.Doc(`
			OWNER/REPO
			social distancing

			This repository does not have a README

			View this repository on GitHub: https://github.com/OWNER/REPO
			`),
			stdoutTTY: true,
		},
		{
			name: "nontty",
			wantOut: heredoc.Doc(`
			name:	OWNER/REPO
			description:	social distancing
			`),
		},
	}

	for _, tt := range tests {
		reg := &httpmock.Registry{}
		reg.Register(
			httpmock.GraphQL(`query RepositoryInfo\b`),
			httpmock.StringResponse(`
		{ "data": {
				"repository": {
				"description": "social distancing"
		} } }`))
		reg.Register(
			httpmock.REST("GET", literal_6498),
			httpmock.StatusStringResponse(404, `{}`))

		opts := &ViewOptions{
			HttpClient: func() (*http.Client, error) {
				return &http.Client{Transport: reg}, nil
			},
			BaseRepo: func() (ghrepo.Interface, error) {
				return ghrepo.New("OWNER", "REPO"), nil
			},
		}

		io, _, stdout, stderr := iostreams.Test()

		opts.IO = io

		t.Run(tt.name, func(t *testing.T) {
			io.SetStdoutTTY(tt.stdoutTTY)

			if err := viewRun(opts); err != nil {
				t.Errorf(literal_5389, err)
			}
			assert.Equal(t, tt.wantOut, stdout.String())
			assert.Equal(t, "", stderr.String())
			reg.Verify(t)
		})
	}
}

func TestViewRunNoDescription(t *testing.T) {
	tests := []struct {
		name      string
		stdoutTTY bool
		wantOut   string
	}{
		{
			name: "tty",
			wantOut: heredoc.Doc(`
			OWNER/REPO
			No description provided

			# truly cool readme check it out

			View this repository on GitHub: https://github.com/OWNER/REPO
			`),
			stdoutTTY: true,
		},
		{
			name: "nontty",
			wantOut: heredoc.Doc(`
			name:	OWNER/REPO
			description:	
			--
			# truly cool readme check it out
			`),
		},
	}

	for _, tt := range tests {
		reg := &httpmock.Registry{}
		reg.Register(
			httpmock.GraphQL(`query RepositoryInfo\b`),
			httpmock.StringResponse(`
		{ "data": {
				"repository": {
				"description": ""
		} } }`))
		reg.Register(
			httpmock.REST("GET", literal_6498),
			httpmock.StringResponse(`
		{ "name": "readme.org",
		"content": "IyB0cnVseSBjb29sIHJlYWRtZSBjaGVjayBpdCBvdXQ="}`))

		opts := &ViewOptions{
			HttpClient: func() (*http.Client, error) {
				return &http.Client{Transport: reg}, nil
			},
			BaseRepo: func() (ghrepo.Interface, error) {
				return ghrepo.New("OWNER", "REPO"), nil
			},
		}

		io, _, stdout, stderr := iostreams.Test()

		opts.IO = io

		t.Run(tt.name, func(t *testing.T) {
			io.SetStdoutTTY(tt.stdoutTTY)

			if err := viewRun(opts); err != nil {
				t.Errorf(literal_5389, err)
			}
			assert.Equal(t, tt.wantOut, stdout.String())
			assert.Equal(t, "", stderr.String())
			reg.Verify(t)
		})
	}
}

func TestViewRunWithoutUsername(t *testing.T) {
	reg := &httpmock.Registry{}
	reg.Register(
		httpmock.GraphQL(`query UserCurrent\b`),
		httpmock.StringResponse(`
		{ "data": { "viewer": {
			"login": "OWNER"
		}}}`))
	reg.Register(
		httpmock.GraphQL(`query RepositoryInfo\b`),
		httpmock.StringResponse(`
	{ "data": {
		"repository": {
		"description": "social distancing"
	} } }`))
	reg.Register(
		httpmock.REST("GET", literal_6498),
		httpmock.StringResponse(`
	{ "name": "readme.md",
	"content": "IyB0cnVseSBjb29sIHJlYWRtZSBjaGVjayBpdCBvdXQ="}`))

	io, _, stdout, stderr := iostreams.Test()
	io.SetStdoutTTY(false)

	opts := &ViewOptions{
		RepoArg: "REPO",
		HttpClient: func() (*http.Client, error) {
			return &http.Client{Transport: reg}, nil
		},
		IO: io,
		Config: func() (config.Config, error) {
			return config.NewBlankConfig(), nil
		},
	}

	if err := viewRun(opts); err != nil {
		t.Errorf(literal_5389, err)
	}

	assert.Equal(t, heredoc.Doc(`
			name:	OWNER/REPO
			description:	social distancing
			--
			# truly cool readme check it out
			`), stdout.String())
	assert.Equal(t, "", stderr.String())
	reg.Verify(t)
}

func TestViewRunHandlesSpecialCharacters(t *testing.T) {
	tests := []struct {
		name       string
		opts       *ViewOptions
		repoName   string
		stdoutTTY  bool
		wantOut    string
		wantStderr string
		wantErr    bool
	}{
		{
			name: "nontty",
			wantOut: heredoc.Doc(`
				name:	OWNER/REPO
				description:	Some basic special characters " & / < > '
				--
				# < is always > than & ' and "
				`),
		},
		{
			name:      literal_7801,
			stdoutTTY: true,
			wantOut: heredoc.Doc(`
				OWNER/REPO
				Some basic special characters " & / < > '


				  # < is always > than & ' and "                                              



				View this repository on GitHub: https://github.com/OWNER/REPO
			`),
		},
	}
	for _, tt := range tests {
		if tt.opts == nil {
			tt.opts = &ViewOptions{}
		}

		if tt.repoName == "" {
			tt.repoName = "OWNER/REPO"
		}

		tt.opts.BaseRepo = func() (ghrepo.Interface, error) {
			repo, _ := ghrepo.FromFullName(tt.repoName)
			return repo, nil
		}

		reg := &httpmock.Registry{}
		reg.Register(
			httpmock.GraphQL(`query RepositoryInfo\b`),
			httpmock.StringResponse(`
		{ "data": {
			"repository": {
			"description": "Some basic special characters \" & / < > '"
		} } }`))
		reg.Register(
			httpmock.REST("GET", fmt.Sprintf("repos/%s/readme", tt.repoName)),
			httpmock.StringResponse(`
		{ "name": "readme.md",
		"content": "IyA8IGlzIGFsd2F5cyA+IHRoYW4gJiAnIGFuZCAi"}`))

		tt.opts.HttpClient = func() (*http.Client, error) {
			return &http.Client{Transport: reg}, nil
		}

		io, _, stdout, stderr := iostreams.Test()
		tt.opts.IO = io

		t.Run(tt.name, func(t *testing.T) {
			io.SetStdoutTTY(tt.stdoutTTY)

			if err := viewRun(tt.opts); (err != nil) != tt.wantErr {
				t.Errorf("viewRun() error = %v, wantErr %v", err, tt.wantErr)
			}
			assert.Equal(t, tt.wantStderr, stderr.String())
			assert.Equal(t, tt.wantOut, stdout.String())
			reg.Verify(t)
		})
	}
}

func TestViewRunjson(t *testing.T) {
	io, _, stdout, stderr := iostreams.Test()
	io.SetStdoutTTY(false)

	reg := &httpmock.Registry{}
	defer reg.Verify(t)
	reg.StubRepoInfoResponse("OWNER", "REPO", "main")

	opts := &ViewOptions{
		IO: io,
		HttpClient: func() (*http.Client, error) {
			return &http.Client{Transport: reg}, nil
		},
		BaseRepo: func() (ghrepo.Interface, error) {
			return ghrepo.New("OWNER", "REPO"), nil
		},
		Exporter: &testExporter{
			fields: []string{"name", "defaultBranchRef"},
		},
	}

	_, teardown := run.Stub()
	defer teardown(t)

	err := viewRun(opts)
	assert.NoError(t, err)
	assert.Equal(t, heredoc.Doc(`
		name: REPO
		defaultBranchRef: main
	`), stdout.String())
	assert.Equal(t, "", stderr.String())
}

type testExporter struct {
	fields []string
}

func (e *testExporter) Fields() []string {
	return e.fields
}

func (e *testExporter) Write(io *iostreams.IOStreams, data interface{}) error {
	r := data.(*api.Repository)
	fmt.Fprintf(io.Out, "name: %s\n", r.Name)
	fmt.Fprintf(io.Out, "defaultBranchRef: %s\n", r.DefaultBranchRef.Name)
	return nil
}

const literal_7801 = "no args"

const literal_5389 = "viewRun() error = %v"

const literal_2796 = "jill/valentine"

const literal_6498 = "repos/OWNER/REPO/readme"
