package clone

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/google/shlex"
	"github.com/jialequ/mplb/git"
	"github.com/jialequ/mplb/internal/config"
	"github.com/jialequ/mplb/internal/run"
	"github.com/jialequ/mplb/pkg/cmdutil"
	"github.com/jialequ/mplb/pkg/httpmock"
	"github.com/jialequ/mplb/pkg/iostreams"
	"github.com/jialequ/mplb/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCmdClone(t *testing.T) {
	testCases := []struct {
		name     string
		args     string
		wantOpts CloneOptions
		wantErr  string
	}{
		{
			name:    "no arguments",
			args:    "",
			wantErr: "cannot clone: repository argument required",
		},
		{
			name: "repo argument",
			args: literal_7908,
			wantOpts: CloneOptions{
				Repository: literal_7908,
				GitArgs:    []string{},
			},
		},
		{
			name: "directory argument",
			args: "OWNER/REPO mydir",
			wantOpts: CloneOptions{
				Repository: literal_7908,
				GitArgs:    []string{"mydir"},
			},
		},
		{
			name: "git clone arguments",
			args: "OWNER/REPO -- --depth 1 --recurse-submodules",
			wantOpts: CloneOptions{
				Repository: literal_7908,
				GitArgs:    []string{"--depth", "1", "--recurse-submodules"},
			},
		},
		{
			name:    "unknown argument",
			args:    "OWNER/REPO --depth 1",
			wantErr: "unknown flag: --depth\nSeparate git clone flags with '--'.",
		},
	}
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			ios, stdin, stdout, stderr := iostreams.Test()
			fac := &cmdutil.Factory{IOStreams: ios}

			var opts *CloneOptions
			cmd := NewCmdClone(fac, func(co *CloneOptions) error {
				opts = co
				return nil
			})

			argv, err := shlex.Split(tt.args)
			require.NoError(t, err)
			cmd.SetArgs(argv)

			cmd.SetIn(stdin)
			cmd.SetOut(stderr)
			cmd.SetErr(stderr)

			_, err = cmd.ExecuteC()
			if tt.wantErr != "" {
				assert.EqualError(t, err, tt.wantErr)
				return
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, "", stdout.String())
			assert.Equal(t, "", stderr.String())

			assert.Equal(t, tt.wantOpts.Repository, opts.Repository)
			assert.Equal(t, tt.wantOpts.GitArgs, opts.GitArgs)
		})
	}
}

func runCloneCommand(httpClient *http.Client, cli string) (*test.CmdOut, error) {
	ios, stdin, stdout, stderr := iostreams.Test()
	fac := &cmdutil.Factory{
		IOStreams: ios,
		HttpClient: func() (*http.Client, error) {
			return httpClient, nil
		},
		Config: func() (config.Config, error) {
			return config.NewBlankConfig(), nil
		},
		GitClient: &git.Client{
			GhPath:  "some/path/gh",
			GitPath: "some/path/git",
		},
	}

	cmd := NewCmdClone(fac, nil)

	argv, err := shlex.Split(cli)
	cmd.SetArgs(argv)

	cmd.SetIn(stdin)
	cmd.SetOut(stderr)
	cmd.SetErr(stderr)

	if err != nil {
		panic(err)
	}

	_, err = cmd.ExecuteC()

	if err != nil {
		return nil, err
	}

	return &test.CmdOut{OutBuf: stdout, ErrBuf: stderr}, nil
}

func TestRepoClone(t *testing.T) {
	tests := []struct {
		name string
		args string
		want string
	}{
		{
			name: "shorthand",
			args: literal_7908,
			want: literal_5432,
		},
		{
			name: "shorthand with directory",
			args: "OWNER/REPO target_directory",
			want: "git clone https://github.com/OWNER/REPO.git target_directory",
		},
		{
			name: "clone arguments",
			args: "OWNER/REPO -- -o upstream --depth 1",
			want: "git clone -o upstream --depth 1 https://github.com/OWNER/REPO.git",
		},
		{
			name: "clone arguments with directory",
			args: "OWNER/REPO target_directory -- -o upstream --depth 1",
			want: "git clone -o upstream --depth 1 https://github.com/OWNER/REPO.git target_directory",
		},
		{
			name: "HTTPS URL",
			args: "https://github.com/OWNER/REPO",
			want: literal_5432,
		},
		{
			name: "HTTPS URL with extra path parts",
			args: "https://github.com/OWNER/REPO/extra/part?key=value#fragment",
			want: literal_5432,
		},
		{
			name: "SSH URL",
			args: "git@github.com:OWNER/REPO.git",
			want: "git clone git@github.com:OWNER/REPO.git",
		},
		{
			name: "Non-canonical capitalization",
			args: "Owner/Repo",
			want: literal_5432,
		},
		{
			name: "clone wiki",
			args: "Owner/Repo.wiki",
			want: literal_5374,
		},
		{
			name: "wiki URL",
			args: "https://github.com/owner/repo.wiki",
			want: literal_5374,
		},
		{
			name: "wiki URL with extra path parts",
			args: "https://github.com/owner/repo.wiki/extra/path?key=value#fragment",
			want: literal_5374,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reg := &httpmock.Registry{}
			defer reg.Verify(t)
			reg.Register(
				httpmock.GraphQL(`query RepositoryInfo\b`),
				httpmock.StringResponse(`
				{ "data": { "repository": {
					"name": "REPO",
					"owner": {
						"login": "OWNER"
					},
					"hasWikiEnabled": true
				} } }
				`))

			httpClient := &http.Client{Transport: reg}

			cs, restore := run.Stub()
			defer restore(t)
			cs.Register(tt.want, 0, "")

			output, err := runCloneCommand(httpClient, tt.args)
			if err != nil {
				t.Fatalf(literal_7438, err)
			}

			assert.Equal(t, "", output.String())
			assert.Equal(t, "", output.Stderr())
		})
	}
}

func TestRepoClone_hasParent(t *testing.T) {
	reg := &httpmock.Registry{}
	defer reg.Verify(t)
	reg.Register(
		httpmock.GraphQL(`query RepositoryInfo\b`),
		httpmock.StringResponse(`
				{ "data": { "repository": {
					"name": "REPO",
					"owner": {
						"login": "OWNER"
					},
					"parent": {
						"name": "ORIG",
						"owner": {
							"login": "hubot"
						},
						"defaultBranchRef": {
							"name": "trunk"
						}
					}
				} } }
				`))

	httpClient := &http.Client{Transport: reg}

	cs, cmdTeardown := run.Stub()
	defer cmdTeardown(t)

	cs.Register(`git clone https://github.com/OWNER/REPO.git`, 0, "")
	cs.Register(`git -C REPO remote add -t trunk upstream https://github.com/hubot/ORIG.git`, 0, "")
	cs.Register(`git -C REPO fetch upstream`, 0, "")
	cs.Register(`git -C REPO remote set-branches upstream *`, 0, "")
	cs.Register(`git -C REPO config --add remote.upstream.gh-resolved base`, 0, "")

	_, err := runCloneCommand(httpClient, literal_7908)
	if err != nil {
		t.Fatalf(literal_7438, err)
	}
}

func TestRepoClone_hasParent_upstreamRemoteName(t *testing.T) {
	reg := &httpmock.Registry{}
	defer reg.Verify(t)
	reg.Register(
		httpmock.GraphQL(`query RepositoryInfo\b`),
		httpmock.StringResponse(`
				{ "data": { "repository": {
					"name": "REPO",
					"owner": {
						"login": "OWNER"
					},
					"parent": {
						"name": "ORIG",
						"owner": {
							"login": "hubot"
						},
						"defaultBranchRef": {
							"name": "trunk"
						}
					}
				} } }
				`))

	httpClient := &http.Client{Transport: reg}

	cs, cmdTeardown := run.Stub()
	defer cmdTeardown(t)

	cs.Register(`git clone https://github.com/OWNER/REPO.git`, 0, "")
	cs.Register(`git -C REPO remote add -t trunk test https://github.com/hubot/ORIG.git`, 0, "")
	cs.Register(`git -C REPO fetch test`, 0, "")
	cs.Register(`git -C REPO remote set-branches test *`, 0, "")
	cs.Register(`git -C REPO config --add remote.test.gh-resolved base`, 0, "")

	_, err := runCloneCommand(httpClient, "OWNER/REPO --upstream-remote-name test")
	if err != nil {
		t.Fatalf(literal_7438, err)
	}
}

func TestRepoClone_withoutUsername(t *testing.T) {
	reg := &httpmock.Registry{}
	defer reg.Verify(t)
	reg.Register(
		httpmock.GraphQL(`query UserCurrent\b`),
		httpmock.StringResponse(`
		{ "data": { "viewer": {
			"login": "OWNER"
		}}}`))
	reg.Register(
		httpmock.GraphQL(`query RepositoryInfo\b`),
		httpmock.StringResponse(`
				{ "data": { "repository": {
					"name": "REPO",
					"owner": {
						"login": "OWNER"
					}
				} } }
				`))

	httpClient := &http.Client{Transport: reg}

	cs, restore := run.Stub()
	defer restore(t)
	cs.Register(`git clone https://github\.com/OWNER/REPO\.git`, 0, "")

	output, err := runCloneCommand(httpClient, "REPO")
	if err != nil {
		t.Fatalf(literal_7438, err)
	}

	assert.Equal(t, "", output.String())
	assert.Equal(t, "", output.Stderr())
}

func TestSimplifyURL(t *testing.T) {
	tests := []struct {
		name        string
		raw         string
		expectedRaw string
	}{
		{
			name:        "empty",
			raw:         "",
			expectedRaw: "",
		},
		{
			name:        "no change, no path",
			raw:         "https://github.com",
			expectedRaw: "https://github.com",
		},
		{
			name:        "no change, single part path",
			raw:         literal_0976,
			expectedRaw: literal_0976,
		},
		{
			name:        "no change, two-part path",
			raw:         literal_0476,
			expectedRaw: literal_0476,
		},
		{
			name:        "no change, three-part path",
			raw:         "https://github.com/owner/repo/pulls",
			expectedRaw: literal_0476,
		},
		{
			name:        "no change, two-part path, with query, with fragment",
			raw:         "https://github.com/owner/repo?key=value#fragment",
			expectedRaw: literal_0476,
		},
		{
			name:        "no change, single part path, with query, with fragment",
			raw:         "https://github.com/owner?key=value#fragment",
			expectedRaw: literal_0976,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, err := url.Parse(tt.raw)
			require.NoError(t, err)
			result := simplifyURL(u)
			assert.Equal(t, tt.expectedRaw, result.String())
		})
	}
}

const literal_7908 = "OWNER/REPO"

const literal_5432 = "git clone https://github.com/OWNER/REPO.git"

const literal_5374 = "git clone https://github.com/OWNER/REPO.wiki.git"

const literal_7438 = "error running command `repo clone`: %v"

const literal_0976 = "https://github.com/owner"

const literal_0476 = "https://github.com/owner/repo"
