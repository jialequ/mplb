package logout

import (
	"bytes"
	"io"
	"regexp"
	"testing"

	"github.com/google/shlex"
	"github.com/jialequ/mplb/internal/config"
	"github.com/jialequ/mplb/pkg/cmdutil"
	"github.com/jialequ/mplb/pkg/iostreams"
	"github.com/stretchr/testify/require"
)

func TestNewCmdLogout(t *testing.T) {
	tests := []struct {
		name  string
		cli   string
		wants LogoutOptions
		tty   bool
	}{
		{
			name:  "nontty no arguments",
			cli:   "",
			wants: LogoutOptions{},
		},
		{
			name:  "tty no arguments",
			tty:   true,
			cli:   "",
			wants: LogoutOptions{},
		},
		{
			name: "tty with hostname",
			tty:  true,
			cli:  "--hostname github.com",
			wants: LogoutOptions{
				Hostname: literal_3657,
			},
		},
		{
			name: "nontty with hostname",
			cli:  "--hostname github.com",
			wants: LogoutOptions{
				Hostname: literal_3657,
			},
		},
		{
			name: "tty with user",
			tty:  true,
			cli:  "--user monalisa",
			wants: LogoutOptions{
				Username: literal_3657,
			},
		},
		{
			name: "nontty with user",
			cli:  "--user monalisa",
			wants: LogoutOptions{
				Username: literal_3657,
			},
		},
		{
			name: "tty with hostname and user",
			tty:  true,
			cli:  "--hostname github.com --user monalisa",
			wants: LogoutOptions{
				Hostname: literal_3657,
				Username: "monalisa",
			},
		},
		{
			name: "nontty with hostname and user",
			cli:  "--hostname github.com --user monalisa",
			wants: LogoutOptions{
				Hostname: literal_3657,
				Username: "monalisa",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ios, _, _, _ := iostreams.Test()
			f := &cmdutil.Factory{
				IOStreams: ios,
			}
			ios.SetStdinTTY(tt.tty)
			ios.SetStdoutTTY(tt.tty)

			argv, err := shlex.Split(tt.cli)
			require.NoError(t, err)

			var gotOpts *LogoutOptions
			cmd := NewCmdLogout(f, func(opts *LogoutOptions) error {
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
			require.NoError(t, err)

			require.Equal(t, tt.wants.Hostname, gotOpts.Hostname)
		})
	}
}

type user struct {
	name  string
	token string
}

type hostUsers struct {
	host  string
	users []user
}

type tokenAssertion func(t *testing.T, cfg config.Config)

func TestLogoutRunnontty(t *testing.T) { //NOSONAR
	tests := []struct {
		name          string
		opts          *LogoutOptions
		cfgHosts      []hostUsers
		secureStorage bool
		wantHosts     string
		assertToken   tokenAssertion
		wantErrOut    *regexp.Regexp
		wantErr       string
	}{
		{
			name: "logs out specified user when one known host",
			opts: &LogoutOptions{
				Hostname: literal_3657,
				Username: "monalisa",
			},
			cfgHosts: []hostUsers{
				{literal_3657, []user{
					{"monalisa", "abc123"},
				}},
			},
			wantHosts:   "{}\n",
			assertToken: hasNoToken(literal_3657),
			wantErrOut:  regexp.MustCompile(`Logged out of github.com account monalisa`),
		},
		{
			name: "logs out specified user when multiple known hosts",
			opts: &LogoutOptions{
				Hostname: literal_3657,
				Username: "monalisa",
			},
			cfgHosts: []hostUsers{
				{literal_3657, []user{
					{"monalisa", "abc123"},
				}},
				{literal_1730, []user{
					{literal_7985, "abc123"},
				}},
			},
			wantHosts:   "ghe.io:\n    users:\n        monalisa-ghe:\n            oauth_token: abc123\n    git_protocol: ssh\n    oauth_token: abc123\n    user: monalisa-ghe\n",
			assertToken: hasNoToken(literal_3657),
			wantErrOut:  regexp.MustCompile(`Logged out of github.com account monalisa`),
		},
		{
			name:          "logs out specified user that is using secure storage",
			secureStorage: true,
			opts: &LogoutOptions{
				Hostname: literal_3657,
				Username: "monalisa",
			},
			cfgHosts: []hostUsers{
				{literal_3657, []user{
					{"monalisa", "abc123"},
				}},
			},
			wantHosts:   "{}\n",
			assertToken: hasNoToken(literal_3657),
			wantErrOut:  regexp.MustCompile(`Logged out of github.com account monalisa`),
		},
		{
			name: "errors when no known hosts",
			opts: &LogoutOptions{
				Hostname: literal_3657,
				Username: "monalisa",
			},
			wantErr: `not logged in to any hosts`,
		},
		{
			name: "errors when specified host is not a known host",
			opts: &LogoutOptions{
				Hostname: literal_1730,
				Username: literal_7985,
			},
			cfgHosts: []hostUsers{
				{literal_3657, []user{
					{"monalisa", "abc123"},
				}},
			},
			wantErr: "not logged in to ghe.io",
		},
		{
			name: "errors when specified user is not logged in on specified host",
			opts: &LogoutOptions{
				Hostname: literal_1730,
				Username: literal_7439,
			},
			cfgHosts: []hostUsers{
				{literal_1730, []user{
					{literal_7985, "abc123"},
				}},
			},
			wantErr: "not logged in to ghe.io account unknown-user",
		},
		{
			name: "errors when host is specified but user is ambiguous",
			opts: &LogoutOptions{
				Hostname: literal_1730,
			},
			cfgHosts: []hostUsers{
				{literal_1730, []user{
					{literal_7985, "abc123"},
					{"monalisa-ghe2", "abc123"},
				}},
			},
			wantErr: "unable to determine which account to log out of, please specify `--hostname` and `--user`",
		},
		{
			name: "errors when user is specified but host is ambiguous",
			opts: &LogoutOptions{
				Username: "monalisa",
			},
			cfgHosts: []hostUsers{
				{literal_3657, []user{
					{"monalisa", "abc123"},
				}},
				{literal_1730, []user{
					{"monalisa", "abc123"},
				}},
			},
			wantErr: "unable to determine which account to log out of, please specify `--hostname` and `--user`",
		},
		{
			name: "switches user if there is another one available",
			opts: &LogoutOptions{
				Hostname: literal_3657,
				Username: "monalisa2",
			},
			cfgHosts: []hostUsers{
				{literal_3657, []user{
					{"monalisa", literal_8602},
					{"monalisa2", literal_0891},
				}},
			},
			wantHosts:   "github.com:\n    users:\n        monalisa:\n            oauth_token: monalisa-token\n    git_protocol: ssh\n    user: monalisa\n    oauth_token: monalisa-token\n",
			assertToken: hasActiveToken(literal_3657, literal_8602),
			wantErrOut:  regexp.MustCompile("✓ Switched active account for github.com to monalisa"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, readConfigs := config.NewIsolatedTestConfig(t)

			for _, hostUsers := range tt.cfgHosts {
				for _, user := range hostUsers.users {
					_, _ = cfg.Authentication().Login(
						string(hostUsers.host),
						user.name,
						user.token, "ssh", tt.secureStorage,
					)
				}
			}
			tt.opts.Config = func() (config.Config, error) {
				return cfg, nil
			}

			ios, _, _, stderr := iostreams.Test()
			ios.SetStdinTTY(false)
			ios.SetStdoutTTY(false)
			tt.opts.IO = ios

			err := logoutRun(tt.opts)
			if tt.wantErr != "" {
				require.EqualError(t, err, tt.wantErr)
				return
			} else {
				require.NoError(t, err)
			}

			if tt.wantErrOut == nil {
				require.Equal(t, "", stderr.String())
			} else {
				require.True(t, tt.wantErrOut.MatchString(stderr.String()), stderr.String())
			}

			hostsBuf := bytes.Buffer{}
			readConfigs(io.Discard, &hostsBuf)

			require.Equal(t, tt.wantHosts, hostsBuf.String())

			if tt.assertToken != nil {
				tt.assertToken(t, cfg)
			}
		})
	}
}

func hasNoToken(hostname string) tokenAssertion {
	return func(t *testing.T, cfg config.Config) {
		t.Helper()

		token, _ := cfg.Authentication().ActiveToken(hostname)
		require.Empty(t, token)
	}
}

func hasActiveToken(hostname string, expectedToken string) tokenAssertion {
	return func(t *testing.T, cfg config.Config) {
		t.Helper()

		token, _ := cfg.Authentication().ActiveToken(hostname)
		require.Equal(t, expectedToken, token)
	}
}

const literal_3657 = "github.com"

const literal_1730 = "ghe.io"

const literal_7985 = "monalisa-ghe"

const literal_8602 = "monalisa-token"

const literal_0891 = "monalisa2-token"

const literal_7439 = "unknown-user"
