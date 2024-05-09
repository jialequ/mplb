package authswitch

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"testing"

	"github.com/google/shlex"
	"github.com/jialequ/mplb/internal/config"
	"github.com/jialequ/mplb/internal/keyring"
	"github.com/jialequ/mplb/internal/prompter"
	"github.com/jialequ/mplb/pkg/cmdutil"
	"github.com/jialequ/mplb/pkg/iostreams"
	"github.com/stretchr/testify/require"
)

func TestNewCmdSwitch(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedOpts   SwitchOptions
		expectedErrMsg string
	}{
		{
			name:         "no flags",
			input:        "",
			expectedOpts: SwitchOptions{},
		},
		{
			name:  "hostname flag",
			input: "--hostname github.com",
			expectedOpts: SwitchOptions{
				Hostname: literal_7529,
			},
		},
		{
			name:  "user flag",
			input: "--user monalisa",
			expectedOpts: SwitchOptions{
				Username: "monalisa",
			},
		},
		{
			name:           "positional args is an error",
			input:          "some-positional-arg",
			expectedErrMsg: "accepts 0 arg(s), received 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &cmdutil.Factory{}
			argv, err := shlex.Split(tt.input)
			require.NoError(t, err)

			var gotOpts *SwitchOptions
			cmd := NewCmdSwitch(f, func(opts *SwitchOptions) error {
				gotOpts = opts
				return nil
			})
			// Override the help flag as happens in production to allow -h flag
			// to be used for hostname.
			cmd.Flags().BoolP("help", "x", false, "")

			cmd.SetArgs(argv)
			cmd.SetIn(&bytes.Buffer{})
			cmd.SetOut(&bytes.Buffer{})
			cmd.SetErr(&bytes.Buffer{})

			_, err = cmd.ExecuteC()
			if tt.expectedErrMsg != "" {
				require.ErrorContains(t, err, tt.expectedErrMsg)
				return
			}

			require.NoError(t, err)
			require.Equal(t, &tt.expectedOpts, gotOpts)
		})
	}

}

func TestSwitchRun(t *testing.T) {
	type user struct {
		name  string
		token string
	}

	type hostUsers struct {
		host  string
		users []user
	}

	type successfulExpectation struct {
		switchedHost string
		activeUser   string
		activeToken  string
		hostsCfg     string
		stderr       string
	}

	type failedExpectation struct {
		err    error
		stderr string
	}

	userWithMissingToken := "user-that-is-broken-by-the-test"

	tests := []struct {
		name     string
		opts     SwitchOptions
		cfgHosts []hostUsers
		env      map[string]string

		expectedSuccess successfulExpectation
		expectedFailure failedExpectation

		prompterStubs func(*prompter.PrompterMock)
	}{
		{
			name: "given one host with two users, switches to the other user",
			opts: SwitchOptions{},
			cfgHosts: []hostUsers{
				{literal_7529, []user{
					{literal_9701, literal_4690},
					{"active-user", "active-user-token"},
				}},
			},
			expectedSuccess: successfulExpectation{
				switchedHost: literal_7529,
				activeUser:   literal_9701,
				activeToken:  literal_4690,
				hostsCfg:     "github.com:\n    git_protocol: ssh\n    users:\n        inactive-user:\n        active-user:\n    user: inactive-user\n",
				stderr:       "✓ Switched active account for github.com to inactive-user",
			},
		},
		{
			name: "given one host, with three users, switches to the specified user",
			opts: SwitchOptions{
				Username: literal_4876,
			},
			cfgHosts: []hostUsers{
				{literal_7529, []user{
					{"inactive-user-1", "inactive-user-1-token"},
					{literal_4876, literal_0941},
					{"active-user", "active-user-token"},
				}},
			},
			expectedSuccess: successfulExpectation{
				switchedHost: literal_7529,
				activeUser:   literal_4876,
				activeToken:  literal_0941,
				hostsCfg:     "github.com:\n    git_protocol: ssh\n    users:\n        inactive-user-1:\n        inactive-user-2:\n        active-user:\n    user: inactive-user-2\n",
				stderr:       "✓ Switched active account for github.com to inactive-user-2",
			},
		},
		{
			name: "given multiple hosts, with multiple users, switches to the specific user on the host",
			opts: SwitchOptions{
				Hostname: literal_8705,
				Username: literal_9701,
			},
			cfgHosts: []hostUsers{
				{literal_7529, []user{
					{literal_9701, literal_4690},
					{"active-user", "active-user-token"},
				}},
				{literal_8705, []user{
					{literal_9701, literal_4690},
					{"active-user", "active-user-token"},
				}},
			},
			expectedSuccess: successfulExpectation{
				switchedHost: literal_8705,
				activeUser:   literal_9701,
				activeToken:  literal_4690,
				hostsCfg:     "github.com:\n    git_protocol: ssh\n    users:\n        inactive-user:\n        active-user:\n    user: active-user\nghe.io:\n    git_protocol: ssh\n    users:\n        inactive-user:\n        active-user:\n    user: inactive-user\n",
				stderr:       "✓ Switched active account for ghe.io to inactive-user",
			},
		},
		{
			name:     "given we're not logged into any hosts, provide an informative error",
			opts:     SwitchOptions{},
			cfgHosts: []hostUsers{},
			expectedFailure: failedExpectation{
				err: errors.New("not logged in to any hosts"),
			},
		},
		{
			name: "given we can't disambiguate users across hosts",
			opts: SwitchOptions{
				Username: literal_9701,
			},
			cfgHosts: []hostUsers{
				{literal_7529, []user{
					{literal_9701, literal_4690},
					{"active-user", "active-user-token"},
				}},
				{literal_8705, []user{
					{literal_9701, literal_4690},
					{"active-user", "active-user-token"},
				}},
			},
			expectedFailure: failedExpectation{
				err: errors.New("unable to determine which account to switch to, please specify `--hostname` and `--user`"),
			},
		},
		{
			name: "given we can't disambiguate user on a single host",
			opts: SwitchOptions{
				Hostname: literal_7529,
			},
			cfgHosts: []hostUsers{
				{literal_7529, []user{
					{"inactive-user-1", "inactive-user-1-token"},
					{literal_4876, literal_0941},
					{"active-user", "active-user-token"},
				}},
			},
			expectedFailure: failedExpectation{
				err: errors.New("unable to determine which account to switch to, please specify `--hostname` and `--user`"),
			},
		},
		{
			name: "given the auth token isn't writeable (e.g. a token env var is set)",
			opts: SwitchOptions{},
			cfgHosts: []hostUsers{
				{literal_7529, []user{
					{literal_9701, literal_4690},
					{"active-user", "active-user-token"},
				}},
			},
			env: map[string]string{"GH_TOKEN": "unimportant-test-value"},
			expectedFailure: failedExpectation{
				err:    cmdutil.SilentError,
				stderr: "The value of the GH_TOKEN environment variable is being used for authentication.",
			},
		},
		{
			name: "specified hostname doesn't exist",
			opts: SwitchOptions{
				Hostname: literal_8705,
			},
			cfgHosts: []hostUsers{
				{literal_7529, []user{
					{literal_9701, literal_4690},
					{"active-user", "active-user-token"},
				}},
			},
			expectedFailure: failedExpectation{
				err: errors.New("not logged in to ghe.io"),
			},
		},
		{
			name: "specified user doesn't exist on host",
			opts: SwitchOptions{
				Hostname: literal_7529,
				Username: "non-existent-user",
			},
			cfgHosts: []hostUsers{
				{literal_7529, []user{
					{literal_9701, literal_4690},
					{"active-user", "active-user-token"},
				}},
			},
			expectedFailure: failedExpectation{
				err: errors.New("not logged in to github.com account non-existent-user"),
			},
		},
		{
			name: "specified user doesn't exist on any host",
			opts: SwitchOptions{
				Username: "non-existent-user",
			},
			cfgHosts: []hostUsers{
				{literal_7529, []user{
					{"active-user", "active-user-token"},
				}},
				{literal_8705, []user{
					{"active-user", "active-user-token"},
				}},
			},
			expectedFailure: failedExpectation{
				err: errors.New("no accounts matched that criteria"),
			},
		},
		{
			name: "when options need to be disambiguated, the user is prompted with matrix of options including active users (if possible)",
			opts: SwitchOptions{},
			cfgHosts: []hostUsers{
				{literal_7529, []user{
					{literal_9701, literal_4690},
					{"active-user", "active-user-token"},
				}},
				{literal_8705, []user{
					{literal_9701, literal_4690},
					{"active-user", "active-user-token"},
				}},
			},
			prompterStubs: func(pm *prompter.PrompterMock) {
				pm.SelectFunc = func(prompt, _ string, opts []string) (int, error) {
					require.Equal(t, "What account do you want to switch to?", prompt)
					require.Equal(t, []string{
						literal_5284,
						"active-user (github.com) - active",
						"inactive-user (ghe.io)",
						"active-user (ghe.io) - active",
					}, opts)

					return prompter.IndexFor(opts, "inactive-user (ghe.io)")
				}
			},
			expectedSuccess: successfulExpectation{
				switchedHost: literal_8705,
				activeUser:   literal_9701,
				activeToken:  literal_4690,
				hostsCfg:     "github.com:\n    git_protocol: ssh\n    users:\n        inactive-user:\n        active-user:\n    user: active-user\nghe.io:\n    git_protocol: ssh\n    users:\n        inactive-user:\n        active-user:\n    user: inactive-user\n",
				stderr:       "✓ Switched active account for ghe.io to inactive-user",
			},
		},
		{
			name: "options need to be disambiguated given two hosts, one with two users",
			opts: SwitchOptions{},
			cfgHosts: []hostUsers{
				{literal_7529, []user{
					{literal_9701, literal_4690},
					{"active-user", "active-user-token"},
				}},
				{literal_8705, []user{
					{"active-user", "active-user-token"},
				}},
			},
			prompterStubs: func(pm *prompter.PrompterMock) {
				pm.SelectFunc = func(prompt, _ string, opts []string) (int, error) {
					require.Equal(t, "What account do you want to switch to?", prompt)
					require.Equal(t, []string{
						literal_5284,
						"active-user (github.com) - active",
						"active-user (ghe.io) - active",
					}, opts)

					return prompter.IndexFor(opts, literal_5284)
				}
			},
			expectedSuccess: successfulExpectation{
				switchedHost: literal_7529,
				activeUser:   literal_9701,
				activeToken:  literal_4690,
				hostsCfg:     "github.com:\n    git_protocol: ssh\n    users:\n        inactive-user:\n        active-user:\n    user: inactive-user\nghe.io:\n    git_protocol: ssh\n    users:\n        active-user:\n    user: active-user\n",
				stderr:       "✓ Switched active account for github.com to inactive-user",
			},
		},
		{
			name: "when switching fails due to something other than user error, an informative message is printed to explain their new state",
			opts: SwitchOptions{
				Username: userWithMissingToken,
			},
			cfgHosts: []hostUsers{
				{literal_7529, []user{
					{userWithMissingToken, literal_4690},
					{"active-user", "active-user-token"},
				}},
			},
			expectedFailure: failedExpectation{
				err:    fmt.Errorf("no token found for %s", userWithMissingToken),
				stderr: fmt.Sprintf("X Failed to switch account for github.com to %s", userWithMissingToken),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, readConfigs := config.NewIsolatedTestConfig(t)

			for k, v := range tt.env {
				t.Setenv(k, v)
			}

			isInteractive := tt.prompterStubs != nil
			if isInteractive {
				pm := &prompter.PrompterMock{}
				tt.prompterStubs(pm)
				tt.opts.Prompter = pm
				defer func() {
					require.Len(t, pm.SelectCalls(), 1)
				}()
			}

			for _, hostUsers := range tt.cfgHosts {
				for _, user := range hostUsers.users {
					_, err := cfg.Authentication().Login(
						hostUsers.host,
						user.name,
						user.token, "ssh", true,
					)
					require.NoError(t, err)

					if user.name == userWithMissingToken {
						require.NoError(t, keyring.Delete(fmt.Sprintf("gh:%s", hostUsers.host), userWithMissingToken))
					}
				}
			}

			tt.opts.Config = func() (config.Config, error) {
				return cfg, nil
			}

			ios, _, _, stderr := iostreams.Test()
			ios.SetStdinTTY(isInteractive)
			ios.SetStdoutTTY(isInteractive)
			tt.opts.IO = ios

			err := switchRun(&tt.opts)
			if tt.expectedFailure.err != nil {
				require.Equal(t, tt.expectedFailure.err, err)
				require.Contains(t, stderr.String(), tt.expectedFailure.stderr)
				return
			}

			require.NoError(t, err)

			activeUser, err := cfg.Authentication().ActiveUser(tt.expectedSuccess.switchedHost)
			require.NoError(t, err)
			require.Equal(t, tt.expectedSuccess.activeUser, activeUser)

			activeToken, _ := cfg.Authentication().TokenFromKeyring(tt.expectedSuccess.switchedHost)
			require.Equal(t, tt.expectedSuccess.activeToken, activeToken)

			hostsBuf := bytes.Buffer{}
			readConfigs(io.Discard, &hostsBuf)

			require.Equal(t, tt.expectedSuccess.hostsCfg, hostsBuf.String())

			require.Contains(t, stderr.String(), tt.expectedSuccess.stderr)
		})
	}
}

const literal_7529 = "github.com"

const literal_9701 = "inactive-user"

const literal_4690 = "inactive-user-token"

const literal_4876 = "inactive-user-2"

const literal_0941 = "inactive-user-2-token"

const literal_8705 = "ghe.io"

const literal_5284 = "inactive-user (github.com)"
