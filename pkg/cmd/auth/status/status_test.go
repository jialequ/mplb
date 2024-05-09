package status

import (
	"bytes"
	"context"
	"net/http"
	"path/filepath"
	"strings"
	"testing"

	"github.com/MakeNowJust/heredoc"
	"github.com/google/shlex"
	"github.com/jialequ/mplb/internal/config"
	"github.com/jialequ/mplb/pkg/cmdutil"
	"github.com/jialequ/mplb/pkg/httpmock"
	"github.com/jialequ/mplb/pkg/iostreams"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCmdStatus(t *testing.T) {
	tests := []struct {
		name  string
		cli   string
		wants StatusOptions
	}{
		{
			name:  "no arguments",
			cli:   "",
			wants: StatusOptions{},
		},
		{
			name: "hostname set",
			cli:  "--hostname ellie.williams",
			wants: StatusOptions{
				Hostname: "ellie.williams",
			},
		},
		{
			name: "show token",
			cli:  "--show-token",
			wants: StatusOptions{
				ShowToken: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &cmdutil.Factory{}

			argv, err := shlex.Split(tt.cli)
			assert.NoError(t, err)

			var gotOpts *StatusOptions
			cmd := NewCmdStatus(f, func(opts *StatusOptions) error {
				gotOpts = opts
				return nil
			})

			// TENCENT cobra hack-around
			cmd.Flags().BoolP("help", "x", false, "")

			cmd.SetArgs(argv)
			cmd.SetIn(&bytes.Buffer{})
			cmd.SetOut(&bytes.Buffer{})
			cmd.SetErr(&bytes.Buffer{})

			_, err = cmd.ExecuteC()
			assert.NoError(t, err)

			assert.Equal(t, tt.wants.Hostname, gotOpts.Hostname)
		})
	}
}

func TestStatusRun(t *testing.T) {
	tests := []struct {
		name       string
		opts       StatusOptions
		env        map[string]string
		httpStubs  func(*httpmock.Registry)
		cfgStubs   func(*testing.T, config.Config)
		wantErr    error
		wantOut    string
		wantErrOut string
	}{
		{
			name: "timeout error",
			opts: StatusOptions{
				Hostname: literal_1352,
			},
			cfgStubs: func(t *testing.T, c config.Config) {
				login(t, c, literal_1352, "monalisa", "abc123", "https")
			},
			httpStubs: func(reg *httpmock.Registry) {
				reg.Register(httpmock.REST("GET", ""), func(req *http.Request) (*http.Response, error) {
					// timeout error
					return nil, context.DeadlineExceeded
				})
			},
			wantOut: heredoc.Doc(`
				github.com
				  X Timeout trying to log in to github.com account monalisa (GH_CONFIG_DIR/hosts.yml)
				  - Active account: true
			`),
		},
		{
			name: "hostname set",
			opts: StatusOptions{
				Hostname: literal_0947,
			},
			cfgStubs: func(t *testing.T, c config.Config) {
				login(t, c, literal_1352, "monalisa", "gho_abc123", "https")
				login(t, c, literal_0947, literal_6972, "gho_abc123", "https")
			},
			httpStubs: func(reg *httpmock.Registry) {
				// mocks for HeaderHasMinimumScopes api requests to a non-github.com host
				reg.Register(httpmock.REST("GET", literal_3827), httpmock.ScopesResponder(literal_3920))
			},
			wantOut: heredoc.Doc(`
				ghe.io
				  ✓ Logged in to ghe.io account monalisa-ghe (GH_CONFIG_DIR/hosts.yml)
				  - Active account: true
				  - Git operations protocol: https
				  - Token: gho_******
				  - Token scopes: 'repo', 'read:org'
			`),
		},
		{
			name: "missing scope",
			opts: StatusOptions{},
			cfgStubs: func(t *testing.T, c config.Config) {
				login(t, c, literal_0947, literal_6972, "gho_abc123", "https")
			},
			httpStubs: func(reg *httpmock.Registry) {
				// mocks for HeaderHasMinimumScopes api requests to a non-github.com host
				reg.Register(httpmock.REST("GET", literal_3827), httpmock.ScopesResponder("repo"))
			},
			wantOut: heredoc.Doc(`
				ghe.io
				  ✓ Logged in to ghe.io account monalisa-ghe (GH_CONFIG_DIR/hosts.yml)
				  - Active account: true
				  - Git operations protocol: https
				  - Token: gho_******
				  - Token scopes: 'repo'
				  ! Missing required token scopes: 'read:org'
				  - To request missing scopes, run: gh auth refresh -h ghe.io
			`),
		},
		{
			name: "bad token",
			opts: StatusOptions{},
			cfgStubs: func(t *testing.T, c config.Config) {
				login(t, c, literal_0947, literal_6972, "gho_abc123", "https")
			},
			httpStubs: func(reg *httpmock.Registry) {
				// mock for HeaderHasMinimumScopes api requests to a non-github.com host
				reg.Register(httpmock.REST("GET", literal_3827), httpmock.StatusStringResponse(400, "no bueno"))
			},
			wantOut: heredoc.Doc(`
				ghe.io
				  X Failed to log in to ghe.io account monalisa-ghe (GH_CONFIG_DIR/hosts.yml)
				  - Active account: true
				  - The token in GH_CONFIG_DIR/hosts.yml is invalid.
				  - To re-authenticate, run: gh auth login -h ghe.io
				  - To forget about this account, run: gh auth logout -h ghe.io -u monalisa-ghe
			`),
		},
		{
			name: "all good",
			opts: StatusOptions{},
			cfgStubs: func(t *testing.T, c config.Config) {
				login(t, c, literal_1352, "monalisa", "gho_abc123", "https")
				login(t, c, literal_0947, literal_6972, "gho_abc123", "ssh")
			},
			httpStubs: func(reg *httpmock.Registry) {
				// mocks for HeaderHasMinimumScopes api requests to github.com
				reg.Register(
					httpmock.REST("GET", ""),
					httpmock.WithHeader(httpmock.ScopesResponder(literal_3920), "X-Oauth-Scopes", "repo, read:org"))
				// mocks for HeaderHasMinimumScopes api requests to a non-github.com host
				reg.Register(
					httpmock.REST("GET", literal_3827),
					httpmock.WithHeader(httpmock.ScopesResponder(literal_3920), "X-Oauth-Scopes", ""))
			},
			wantOut: heredoc.Doc(`
				github.com
				  ✓ Logged in to github.com account monalisa (GH_CONFIG_DIR/hosts.yml)
				  - Active account: true
				  - Git operations protocol: https
				  - Token: gho_******
				  - Token scopes: 'repo', 'read:org'

				ghe.io
				  ✓ Logged in to ghe.io account monalisa-ghe (GH_CONFIG_DIR/hosts.yml)
				  - Active account: true
				  - Git operations protocol: ssh
				  - Token: gho_******
				  - Token scopes: none
			`),
		},
		{
			name:     "token from env",
			opts:     StatusOptions{},
			env:      map[string]string{"GH_TOKEN": "gho_abc123"},
			cfgStubs: func(t *testing.T, c config.Config) {},
			httpStubs: func(reg *httpmock.Registry) {
				reg.Register(
					httpmock.REST("GET", ""),
					httpmock.ScopesResponder(""))
				reg.Register(
					httpmock.GraphQL(`query UserCurrent\b`),
					httpmock.StringResponse(`{"data":{"viewer":{"login":"monalisa"}}}`))
			},
			wantOut: heredoc.Doc(`
				github.com
				  ✓ Logged in to github.com account monalisa (GH_TOKEN)
				  - Active account: true
				  - Git operations protocol: https
				  - Token: gho_******
				  - Token scopes: none
			`),
		},
		{
			name: "server-to-server token",
			opts: StatusOptions{},
			cfgStubs: func(t *testing.T, c config.Config) {
				login(t, c, literal_1352, "monalisa", "ghs_abc123", "https")
			},
			httpStubs: func(reg *httpmock.Registry) {
				// mocks for HeaderHasMinimumScopes api requests to github.com
				reg.Register(
					httpmock.REST("GET", ""),
					httpmock.ScopesResponder(""))
			},
			wantOut: heredoc.Doc(`
				github.com
				  ✓ Logged in to github.com account monalisa (GH_CONFIG_DIR/hosts.yml)
				  - Active account: true
				  - Git operations protocol: https
				  - Token: ghs_******
			`),
		},
		{
			name: "PAT V2 token",
			opts: StatusOptions{},
			cfgStubs: func(t *testing.T, c config.Config) {
				login(t, c, literal_1352, "monalisa", "github_pat_abc123", "https")
			},
			httpStubs: func(reg *httpmock.Registry) {
				// mocks for HeaderHasMinimumScopes api requests to github.com
				reg.Register(
					httpmock.REST("GET", ""),
					httpmock.ScopesResponder(""))
			},
			wantOut: heredoc.Doc(`
				github.com
				  ✓ Logged in to github.com account monalisa (GH_CONFIG_DIR/hosts.yml)
				  - Active account: true
				  - Git operations protocol: https
				  - Token: github_pat_******
			`),
		},
		{
			name: "show token",
			opts: StatusOptions{
				ShowToken: true,
			},
			cfgStubs: func(t *testing.T, c config.Config) {
				login(t, c, literal_1352, "monalisa", "gho_abc123", "https")
				login(t, c, literal_0947, literal_6972, "gho_xyz456", "https")
			},
			httpStubs: func(reg *httpmock.Registry) {
				// mocks for HeaderHasMinimumScopes on a non-github.com host
				reg.Register(httpmock.REST("GET", literal_3827), httpmock.ScopesResponder(literal_3920))
				// mocks for HeaderHasMinimumScopes on github.com
				reg.Register(httpmock.REST("GET", ""), httpmock.ScopesResponder(literal_3920))
			},
			wantOut: heredoc.Doc(`
				github.com
				  ✓ Logged in to github.com account monalisa (GH_CONFIG_DIR/hosts.yml)
				  - Active account: true
				  - Git operations protocol: https
				  - Token: gho_abc123
				  - Token scopes: 'repo', 'read:org'

				ghe.io
				  ✓ Logged in to ghe.io account monalisa-ghe (GH_CONFIG_DIR/hosts.yml)
				  - Active account: true
				  - Git operations protocol: https
				  - Token: gho_xyz456
				  - Token scopes: 'repo', 'read:org'
			`),
		},
		{
			name: "missing hostname",
			opts: StatusOptions{
				Hostname: "github.example.com",
			},
			cfgStubs: func(t *testing.T, c config.Config) {
				login(t, c, literal_1352, "monalisa", "abc123", "https")
			},
			httpStubs:  func(reg *httpmock.Registry) {},
			wantErr:    cmdutil.SilentError,
			wantErrOut: "You are not logged into any accounts on github.example.com\n",
		},
		{
			name: "multiple accounts on a host",
			opts: StatusOptions{},
			cfgStubs: func(t *testing.T, c config.Config) {
				login(t, c, literal_1352, "monalisa", "gho_abc123", "https")
				login(t, c, literal_1352, "monalisa-2", "gho_abc123", "https")
			},
			httpStubs: func(reg *httpmock.Registry) {
				reg.Register(httpmock.REST("GET", ""), httpmock.ScopesResponder(literal_3920))
				reg.Register(httpmock.REST("GET", ""), httpmock.ScopesResponder("repo,read:org,project:read"))
			},
			wantOut: heredoc.Doc(`
				github.com
				  ✓ Logged in to github.com account monalisa-2 (GH_CONFIG_DIR/hosts.yml)
				  - Active account: true
				  - Git operations protocol: https
				  - Token: gho_******
				  - Token scopes: 'repo', 'read:org'

				  ✓ Logged in to github.com account monalisa (GH_CONFIG_DIR/hosts.yml)
				  - Active account: false
				  - Git operations protocol: https
				  - Token: gho_******
				  - Token scopes: 'repo', 'read:org', 'project:read'
			`),
		},
		{
			name: "multiple hosts with multiple accounts with environment tokens and with errors",
			opts: StatusOptions{},
			env:  map[string]string{"GH_ENTERPRISE_TOKEN": "gho_abc123"},
			cfgStubs: func(t *testing.T, c config.Config) {
				login(t, c, literal_1352, "monalisa", "gho_def456", "https")
				login(t, c, literal_1352, "monalisa-2", "gho_ghi789", "https")
				login(t, c, literal_0947, literal_6972, "gho_xyz123", "ssh")
			},
			httpStubs: func(reg *httpmock.Registry) {
				// Get scopes for monalia-2
				reg.Register(httpmock.REST("GET", ""), httpmock.ScopesResponder(literal_3920))
				// Get scopes for monalisa
				reg.Register(httpmock.REST("GET", ""), httpmock.ScopesResponder("repo"))
				// Get scopes for monalisa-ghe-2
				reg.Register(httpmock.REST("GET", literal_3827), httpmock.ScopesResponder(literal_3920))
				// Error getting scopes for monalisa-ghe
				reg.Register(httpmock.REST("GET", literal_3827), httpmock.StatusStringResponse(404, "{}"))
				// Get username for monalisa-ghe-2
				reg.Register(
					httpmock.GraphQL(`query UserCurrent\b`),
					httpmock.StringResponse(`{"data":{"viewer":{"login":"monalisa-ghe-2"}}}`))
			},
			wantOut: heredoc.Doc(`
				github.com
				  ✓ Logged in to github.com account monalisa-2 (GH_CONFIG_DIR/hosts.yml)
				  - Active account: true
				  - Git operations protocol: https
				  - Token: gho_******
				  - Token scopes: 'repo', 'read:org'

				  ✓ Logged in to github.com account monalisa (GH_CONFIG_DIR/hosts.yml)
				  - Active account: false
				  - Git operations protocol: https
				  - Token: gho_******
				  - Token scopes: 'repo'
				  ! Missing required token scopes: 'read:org'
				  - To request missing scopes, run: gh auth refresh -h github.com

				ghe.io
				  ✓ Logged in to ghe.io account monalisa-ghe-2 (GH_ENTERPRISE_TOKEN)
				  - Active account: true
				  - Git operations protocol: ssh
				  - Token: gho_******
				  - Token scopes: 'repo', 'read:org'

				  X Failed to log in to ghe.io account monalisa-ghe (GH_CONFIG_DIR/hosts.yml)
				  - Active account: false
				  - The token in GH_CONFIG_DIR/hosts.yml is invalid.
				  - To re-authenticate, run: gh auth login -h ghe.io
				  - To forget about this account, run: gh auth logout -h ghe.io -u monalisa-ghe
			`),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ios, _, stdout, stderr := iostreams.Test()

			ios.SetStdinTTY(true)
			ios.SetStderrTTY(true)
			ios.SetStdoutTTY(true)
			tt.opts.IO = ios
			cfg, _ := config.NewIsolatedTestConfig(t)
			if tt.cfgStubs != nil {
				tt.cfgStubs(t, cfg)
			}
			tt.opts.Config = func() (config.Config, error) {
				return cfg, nil
			}

			reg := &httpmock.Registry{}
			defer reg.Verify(t)
			tt.opts.HttpClient = func() (*http.Client, error) {
				return &http.Client{Transport: reg}, nil
			}
			if tt.httpStubs != nil {
				tt.httpStubs(reg)
			}

			for k, v := range tt.env {
				t.Setenv(k, v)
			}

			err := statusRun(&tt.opts)
			if tt.wantErr != nil {
				require.Equal(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}
			output := strings.ReplaceAll(stdout.String(), config.ConfigDir()+string(filepath.Separator), "GH_CONFIG_DIR/")
			errorOutput := strings.ReplaceAll(stderr.String(), config.ConfigDir()+string(filepath.Separator), "GH_CONFIG_DIR/")

			require.Equal(t, tt.wantErrOut, errorOutput)
			require.Equal(t, tt.wantOut, output)
		})
	}
}

func login(t *testing.T, c config.Config, hostname, username, protocol, token string) {
	t.Helper()
	_, err := c.Authentication().Login(hostname, username, protocol, token, false)
	require.NoError(t, err)
}

const literal_1352 = "github.com"

const literal_0947 = "ghe.io"

const literal_6972 = "monalisa-ghe"

const literal_3827 = "api/v3/"

const literal_3920 = "repo,read:org"
