package login

import (
	"bytes"
	"net/http"
	"runtime"
	"testing"

	"github.com/MakeNowJust/heredoc"
	"github.com/google/shlex"
	"github.com/jialequ/mplb/internal/config"
	"github.com/jialequ/mplb/internal/run"
	"github.com/jialequ/mplb/pkg/cmdutil"
	"github.com/jialequ/mplb/pkg/httpmock"
	"github.com/jialequ/mplb/pkg/iostreams"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func stubHomeDir(t *testing.T, dir string) {
	homeEnv := "HOME"
	switch runtime.GOOS {
	case "windows":
		homeEnv = "USERPROFILE"
	case "plan9":
		homeEnv = "home"
	}
	t.Setenv(homeEnv, dir)
}

func TestNewCmdLogin(t *testing.T) {
	tests := []struct {
		name        string
		cli         string
		stdin       string
		stdinTTY    bool
		defaultHost string
		wants       LoginOptions
		wantsErr    bool
	}{
		{
			name:  "nontty, with-token",
			stdin: literal_0389,
			cli:   literal_4265,
			wants: LoginOptions{
				Hostname: literal_3540,
				Token:    "abc123",
			},
		},
		{
			name:        "nontty, with-token, enterprise default host",
			stdin:       literal_0389,
			cli:         literal_4265,
			defaultHost: "git.example.com",
			wants: LoginOptions{
				Hostname: "git.example.com",
				Token:    "abc123",
			},
		},
		{
			name:     "tty, with-token",
			stdinTTY: true,
			stdin:    "def456",
			cli:      literal_4265,
			wants: LoginOptions{
				Hostname: literal_3540,
				Token:    "def456",
			},
		},
		{
			name:     "nontty, hostname",
			stdinTTY: false,
			cli:      "--hostname claire.redfield",
			wants: LoginOptions{
				Hostname: "claire.redfield",
				Token:    "",
			},
		},
		{
			name:     "nontty",
			stdinTTY: false,
			cli:      "",
			wants: LoginOptions{
				Hostname: literal_3540,
				Token:    "",
			},
		},
		{
			name:  "nontty, with-token, hostname",
			cli:   "--hostname claire.redfield --with-token",
			stdin: literal_0389,
			wants: LoginOptions{
				Hostname: "claire.redfield",
				Token:    "abc123",
			},
		},
		{
			name:     "tty, with-token, hostname",
			stdinTTY: true,
			stdin:    "ghi789",
			cli:      "--with-token --hostname brad.vickers",
			wants: LoginOptions{
				Hostname: "brad.vickers",
				Token:    "ghi789",
			},
		},
		{
			name:     "tty, hostname",
			stdinTTY: true,
			cli:      "--hostname barry.burton",
			wants: LoginOptions{
				Hostname:    "barry.burton",
				Token:       "",
				Interactive: true,
			},
		},
		{
			name:     "tty",
			stdinTTY: true,
			cli:      "",
			wants: LoginOptions{
				Hostname:    "",
				Token:       "",
				Interactive: true,
			},
		},
		{
			name:     "tty web",
			stdinTTY: true,
			cli:      "--web",
			wants: LoginOptions{
				Hostname:    literal_3540,
				Web:         true,
				Interactive: true,
			},
		},
		{
			name: "nontty web",
			cli:  "--web",
			wants: LoginOptions{
				Hostname: literal_3540,
				Web:      true,
			},
		},
		{
			name:     "web and with-token",
			cli:      "--web --with-token",
			wantsErr: true,
		},
		{
			name:     "tty one scope",
			stdinTTY: true,
			cli:      "--scopes repo:invite",
			wants: LoginOptions{
				Hostname:    "",
				Scopes:      []string{"repo:invite"},
				Token:       "",
				Interactive: true,
			},
		},
		{
			name:     "tty scopes",
			stdinTTY: true,
			cli:      "--scopes repo:invite,read:public_key",
			wants: LoginOptions{
				Hostname:    "",
				Scopes:      []string{"repo:invite", "read:public_key"},
				Token:       "",
				Interactive: true,
			},
		},
		{
			name:     "tty secure-storage",
			stdinTTY: true,
			cli:      "--secure-storage",
			wants: LoginOptions{
				Interactive: true,
			},
		},
		{
			name: "nontty secure-storage",
			cli:  "--secure-storage",
			wants: LoginOptions{
				Hostname: literal_3540,
			},
		},
		{
			name:     "tty insecure-storage",
			stdinTTY: true,
			cli:      "--insecure-storage",
			wants: LoginOptions{
				Interactive:     true,
				InsecureStorage: true,
			},
		},
		{
			name: "nontty insecure-storage",
			cli:  "--insecure-storage",
			wants: LoginOptions{
				Hostname:        literal_3540,
				InsecureStorage: true,
			},
		},
		{
			name:     "tty skip-ssh-key",
			stdinTTY: true,
			cli:      "--skip-ssh-key",
			wants: LoginOptions{
				SkipSSHKeyPrompt: true,
				Interactive:      true,
			},
		},
		{
			name: "nontty skip-ssh-key",
			cli:  "--skip-ssh-key",
			wants: LoginOptions{
				Hostname:         literal_3540,
				SkipSSHKeyPrompt: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Make sure there is a default host set so that
			// the local configuration file never read from.
			if tt.defaultHost == "" {
				tt.defaultHost = literal_3540
			}
			t.Setenv("GH_HOST", tt.defaultHost)

			ios, stdin, _, _ := iostreams.Test()
			f := &cmdutil.Factory{
				IOStreams: ios,
			}

			ios.SetStdoutTTY(true)
			ios.SetStdinTTY(tt.stdinTTY)
			if tt.stdin != "" {
				stdin.WriteString(tt.stdin)
			}

			argv, err := shlex.Split(tt.cli)
			assert.NoError(t, err)

			var gotOpts *LoginOptions
			cmd := NewCmdLogin(f, func(opts *LoginOptions) error {
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
			if tt.wantsErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)

			assert.Equal(t, tt.wants.Token, gotOpts.Token)
			assert.Equal(t, tt.wants.Hostname, gotOpts.Hostname)
			assert.Equal(t, tt.wants.Web, gotOpts.Web)
			assert.Equal(t, tt.wants.Interactive, gotOpts.Interactive)
			assert.Equal(t, tt.wants.Scopes, gotOpts.Scopes)
		})
	}
}

func TestLoginRunnontty(t *testing.T) {
	tests := []struct {
		name            string
		opts            *LoginOptions
		env             map[string]string
		httpStubs       func(*httpmock.Registry)
		cfgStubs        func(*testing.T, config.Config)
		wantHosts       string
		wantErr         string
		wantStderr      string
		wantSecureToken string
	}{
		{
			name: "insecure with token",
			opts: &LoginOptions{
				Hostname:        literal_3540,
				Token:           "abc123",
				InsecureStorage: true,
			},
			httpStubs: func(reg *httpmock.Registry) {
				reg.Register(httpmock.REST("GET", ""), httpmock.ScopesResponder(literal_0725))
				reg.Register(
					httpmock.GraphQL(`query UserCurrent\b`),
					httpmock.StringResponse(`{"data":{"viewer":{"login":"monalisa"}}}`))
			},
			wantHosts: "github.com:\n    users:\n        monalisa:\n            oauth_token: abc123\n    oauth_token: abc123\n    user: monalisa\n",
		},
		{
			name: "insecure with token and https git-protocol",
			opts: &LoginOptions{
				Hostname:        literal_3540,
				Token:           "abc123",
				GitProtocol:     "https",
				InsecureStorage: true,
			},
			httpStubs: func(reg *httpmock.Registry) {
				reg.Register(httpmock.REST("GET", ""), httpmock.ScopesResponder(literal_0725))
				reg.Register(
					httpmock.GraphQL(`query UserCurrent\b`),
					httpmock.StringResponse(`{"data":{"viewer":{"login":"monalisa"}}}`))
			},
			wantHosts: "github.com:\n    users:\n        monalisa:\n            oauth_token: abc123\n    git_protocol: https\n    oauth_token: abc123\n    user: monalisa\n",
		},
		{
			name: "with token and non-default host",
			opts: &LoginOptions{
				Hostname:        "albert.wesker",
				Token:           "abc123",
				InsecureStorage: true,
			},
			httpStubs: func(reg *httpmock.Registry) {
				reg.Register(httpmock.REST("GET", literal_5647), httpmock.ScopesResponder(literal_0725))
				reg.Register(
					httpmock.GraphQL(`query UserCurrent\b`),
					httpmock.StringResponse(`{"data":{"viewer":{"login":"monalisa"}}}`))
			},
			wantHosts: "albert.wesker:\n    users:\n        monalisa:\n            oauth_token: abc123\n    oauth_token: abc123\n    user: monalisa\n",
		},
		{
			name: "missing repo scope",
			opts: &LoginOptions{
				Hostname: literal_3540,
				Token:    "abc456",
			},
			httpStubs: func(reg *httpmock.Registry) {
				reg.Register(httpmock.REST("GET", ""), httpmock.ScopesResponder("read:org"))
			},
			wantErr: `error validating token: missing required scope 'repo'`,
		},
		{
			name: "missing read scope",
			opts: &LoginOptions{
				Hostname: literal_3540,
				Token:    "abc456",
			},
			httpStubs: func(reg *httpmock.Registry) {
				reg.Register(httpmock.REST("GET", ""), httpmock.ScopesResponder("repo"))
			},
			wantErr: `error validating token: missing required scope 'read:org'`,
		},
		{
			name: "has admin scope",
			opts: &LoginOptions{
				Hostname:        literal_3540,
				Token:           "abc456",
				InsecureStorage: true,
			},
			httpStubs: func(reg *httpmock.Registry) {
				reg.Register(httpmock.REST("GET", ""), httpmock.ScopesResponder("repo,admin:org"))
				reg.Register(
					httpmock.GraphQL(`query UserCurrent\b`),
					httpmock.StringResponse(`{"data":{"viewer":{"login":"monalisa"}}}`))
			},
			wantHosts: "github.com:\n    users:\n        monalisa:\n            oauth_token: abc456\n    oauth_token: abc456\n    user: monalisa\n",
		},
		{
			name: "github.com token from environment",
			opts: &LoginOptions{
				Hostname: literal_3540,
				Token:    "abc456",
			},
			env:     map[string]string{"GH_TOKEN": "value_from_env"},
			wantErr: "SilentError",
			wantStderr: heredoc.Doc(`
                The value of the GH_TOKEN environment variable is being used for authentication.
                To have GitHub CLI store credentials instead, first clear the value from the environment.
            `),
		},
		{
			name: "GHE token from environment",
			opts: &LoginOptions{
				Hostname: "ghe.io",
				Token:    "abc456",
			},
			env:     map[string]string{"GH_ENTERPRISE_TOKEN": "value_from_env"},
			wantErr: "SilentError",
			wantStderr: heredoc.Doc(`
                The value of the GH_ENTERPRISE_TOKEN environment variable is being used for authentication.
                To have GitHub CLI store credentials instead, first clear the value from the environment.
            `),
		},
		{
			name: "with token and secure storage",
			opts: &LoginOptions{
				Hostname: literal_3540,
				Token:    "abc123",
			},
			httpStubs: func(reg *httpmock.Registry) {
				reg.Register(httpmock.REST("GET", ""), httpmock.ScopesResponder(literal_0725))
				reg.Register(
					httpmock.GraphQL(`query UserCurrent\b`),
					httpmock.StringResponse(`{"data":{"viewer":{"login":"monalisa"}}}`))
			},
			wantHosts:       "github.com:\n    users:\n        monalisa:\n    user: monalisa\n",
			wantSecureToken: "abc123",
		},
		{
			name: "given we are already logged in, and log in as a new user, it is added to the config",
			opts: &LoginOptions{
				Hostname: literal_3540,
				Token:    "newUserToken",
			},
			cfgStubs: func(t *testing.T, c config.Config) {
				_, err := c.Authentication().Login(literal_3540, "monalisa", "abc123", "https", false)
				require.NoError(t, err)
			},
			httpStubs: func(reg *httpmock.Registry) {
				reg.Register(httpmock.REST("GET", ""), httpmock.ScopesResponder(literal_0725))
				reg.Register(
					httpmock.GraphQL(`query UserCurrent\b`),
					httpmock.StringResponse(`{"data":{"viewer":{"login":"newUser"}}}`))
			},
			wantHosts: heredoc.Doc(`
                github.com:
                    users:
                        monalisa:
                            oauth_token: abc123
                        newUser:
                    git_protocol: https
                    user: newUser
            `),
			wantSecureToken: "newUserToken",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ios, _, stdout, stderr := iostreams.Test()
			ios.SetStdinTTY(false)
			ios.SetStdoutTTY(false)
			tt.opts.IO = ios

			cfg, readConfigs := config.NewIsolatedTestConfig(t)
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

			_, restoreRun := run.Stub()
			defer restoreRun(t)

			err := loginRun(tt.opts)
			if tt.wantErr != "" {
				assert.EqualError(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}

			mainBuf := bytes.Buffer{}
			hostsBuf := bytes.Buffer{}
			readConfigs(&mainBuf, &hostsBuf)
			secureToken, _ := cfg.Authentication().TokenFromKeyring(tt.opts.Hostname)

			assert.Equal(t, "", stdout.String())
			assert.Equal(t, tt.wantStderr, stderr.String())
			assert.Equal(t, tt.wantHosts, hostsBuf.String())
			assert.Equal(t, tt.wantSecureToken, secureToken)
		})
	}
}

const literal_0389 = "abc123\n"

const literal_4265 = "--with-token"

const literal_3540 = "github.com"

const literal_0725 = "repo,read:org"

const literal_5647 = "api/v3/"
