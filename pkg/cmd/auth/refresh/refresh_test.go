package refresh

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/google/shlex"
	"github.com/jialequ/mplb/internal/config"
	"github.com/jialequ/mplb/internal/prompter"
	"github.com/jialequ/mplb/pkg/cmdutil"
	"github.com/jialequ/mplb/pkg/httpmock"
	"github.com/jialequ/mplb/pkg/iostreams"
	"github.com/stretchr/testify/require"
)

func TestNewCmdRefresh(t *testing.T) {
	tests := []struct {
		name        string
		cli         string
		wants       RefreshOptions
		wantsErr    bool
		tty         bool
		neverPrompt bool
	}{
		{
			name: "tty no arguments",
			tty:  true,
			wants: RefreshOptions{
				Hostname: "",
			},
		},
		{
			name:     "nontty no arguments",
			wantsErr: true,
		},
		{
			name: "nontty hostname",
			cli:  literal_6423,
			wants: RefreshOptions{
				Hostname: "aline.cedrac",
			},
		},
		{
			name: "tty hostname",
			tty:  true,
			cli:  literal_6423,
			wants: RefreshOptions{
				Hostname: "aline.cedrac",
			},
		},
		{
			name:        "prompts disabled, no args",
			tty:         true,
			cli:         "",
			neverPrompt: true,
			wantsErr:    true,
		},
		{
			name:        "prompts disabled, hostname",
			tty:         true,
			cli:         literal_6423,
			neverPrompt: true,
			wants: RefreshOptions{
				Hostname: "aline.cedrac",
			},
		},
		{
			name: "tty one scope",
			tty:  true,
			cli:  "--scopes repo:invite",
			wants: RefreshOptions{
				Scopes: []string{literal_6285},
			},
		},
		{
			name: "tty scopes",
			tty:  true,
			cli:  "--scopes repo:invite,read:public_key",
			wants: RefreshOptions{
				Scopes: []string{literal_6285, literal_3795},
			},
		},
		{
			name:  "secure storage",
			tty:   true,
			cli:   "--secure-storage",
			wants: RefreshOptions{},
		},
		{
			name: "insecure storage",
			tty:  true,
			cli:  "--insecure-storage",
			wants: RefreshOptions{
				InsecureStorage: true,
			},
		},
		{
			name: "reset scopes",
			tty:  true,
			cli:  "--reset-scopes",
			wants: RefreshOptions{
				ResetScopes: true,
			},
		},
		{
			name: "remove scope",
			tty:  true,
			cli:  "--remove-scopes read:public_key",
			wants: RefreshOptions{
				RemoveScopes: []string{literal_3795},
			},
		},
		{
			name: "remove multiple scopes",
			tty:  true,
			cli:  "--remove-scopes workflow,read:public_key",
			wants: RefreshOptions{
				RemoveScopes: []string{"workflow", literal_3795},
			},
		},
		{
			name: "remove scope shorthand",
			tty:  true,
			cli:  "-r read:public_key",
			wants: RefreshOptions{
				RemoveScopes: []string{literal_3795},
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
			ios.SetNeverPrompt(tt.neverPrompt)

			argv, err := shlex.Split(tt.cli)
			require.NoError(t, err)

			var gotOpts *RefreshOptions
			cmd := NewCmdRefresh(f, func(opts *RefreshOptions) error {
				gotOpts = opts
				return nil
			})
			// TODO cobra hack-around
			cmd.Flags().BoolP("help", "x", false, "")

			cmd.SetArgs(argv)
			cmd.SetIn(&bytes.Buffer{})
			cmd.SetOut(&bytes.Buffer{})
			cmd.SetErr(&bytes.Buffer{})

			_, err = cmd.ExecuteC()
			if tt.wantsErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.wants.Hostname, gotOpts.Hostname)
			require.Equal(t, tt.wants.Scopes, gotOpts.Scopes)
		})
	}
}

type authArgs struct {
	hostname      string
	scopes        []string
	interactive   bool
	secureStorage bool
}

type authOut struct {
	username string
	token    string
	err      error
}

func TestRefreshRun(t *testing.T) {
	tests := []struct {
		name          string
		opts          *RefreshOptions
		prompterStubs func(*prompter.PrompterMock)
		cfgHosts      []string
		authOut       authOut
		oldScopes     string
		wantErr       string
		nontty        bool
		wantAuthArgs  authArgs
	}{
		{
			name:    "no hosts configured",
			opts:    &RefreshOptions{},
			wantErr: `not logged in to any hosts`,
		},
		{
			name: "hostname given but not previously authenticated with it",
			cfgHosts: []string{
				literal_7150,
				"aline.cedrac",
			},
			opts: &RefreshOptions{
				Hostname: literal_2873,
			},
			wantErr: `not logged in to obed.morton`,
		},
		{
			name: "hostname provided and is configured",
			cfgHosts: []string{
				literal_2873,
				literal_7150,
			},
			opts: &RefreshOptions{
				Hostname: literal_2873,
			},
			wantAuthArgs: authArgs{
				hostname:      literal_2873,
				scopes:        []string{},
				secureStorage: true,
			},
		},
		{
			name: "no hostname, one host configured",
			cfgHosts: []string{
				literal_7150,
			},
			opts: &RefreshOptions{
				Hostname: "",
			},
			wantAuthArgs: authArgs{
				hostname:      literal_7150,
				scopes:        []string{},
				secureStorage: true,
			},
		},
		{
			name: "no hostname, multiple hosts configured",
			cfgHosts: []string{
				literal_7150,
				"aline.cedrac",
			},
			opts: &RefreshOptions{
				Hostname: "",
			},
			prompterStubs: func(pm *prompter.PrompterMock) {
				pm.SelectFunc = func(_, _ string, opts []string) (int, error) {
					return prompter.IndexFor(opts, literal_7150)
				}
			},
			wantAuthArgs: authArgs{
				hostname:      literal_7150,
				scopes:        []string{},
				secureStorage: true,
			},
		},
		{
			name: "scopes provided",
			cfgHosts: []string{
				literal_7150,
			},
			opts: &RefreshOptions{
				Scopes: []string{literal_6285, literal_7562},
			},
			wantAuthArgs: authArgs{
				hostname:      literal_7150,
				scopes:        []string{literal_6285, literal_7562},
				secureStorage: true,
			},
		},
		{
			name: "more scopes provided",
			cfgHosts: []string{
				literal_7150,
			},
			oldScopes: "delete_repo, codespace",
			opts: &RefreshOptions{
				Scopes: []string{literal_6285, literal_7562},
			},
			wantAuthArgs: authArgs{
				hostname:      literal_7150,
				scopes:        []string{"delete_repo", "codespace", literal_6285, literal_7562},
				secureStorage: true,
			},
		},
		{
			name: "secure storage",
			cfgHosts: []string{
				literal_2873,
			},
			opts: &RefreshOptions{
				Hostname: literal_2873,
			},
			wantAuthArgs: authArgs{
				hostname:      literal_2873,
				scopes:        []string{},
				secureStorage: true,
			},
		},
		{
			name: "insecure storage",
			cfgHosts: []string{
				literal_2873,
			},
			opts: &RefreshOptions{
				Hostname:        literal_2873,
				InsecureStorage: true,
			},
			wantAuthArgs: authArgs{
				hostname: literal_2873,
				scopes:   []string{},
			},
		},
		{
			name: "reset scopes",
			cfgHosts: []string{
				literal_7150,
			},
			oldScopes: "delete_repo, codespace",
			opts: &RefreshOptions{
				Hostname:    literal_7150,
				ResetScopes: true,
			},
			wantAuthArgs: authArgs{
				hostname:      literal_7150,
				scopes:        []string{},
				secureStorage: true,
			},
		},
		{
			name: "reset scopes and add some scopes",
			cfgHosts: []string{
				literal_7150,
			},
			oldScopes: literal_0752,
			opts: &RefreshOptions{
				Scopes:      []string{literal_7562, "workflow"},
				ResetScopes: true,
			},
			wantAuthArgs: authArgs{
				hostname:      literal_7150,
				scopes:        []string{literal_7562, "workflow"},
				secureStorage: true,
			},
		},
		{
			name: "remove scopes",
			cfgHosts: []string{
				literal_7150,
			},
			oldScopes: "delete_repo, codespace, repo:invite, public_key:read",
			opts: &RefreshOptions{
				Hostname:     literal_7150,
				RemoveScopes: []string{"delete_repo", literal_6285},
			},
			wantAuthArgs: authArgs{
				hostname:      literal_7150,
				scopes:        []string{"codespace", literal_7562},
				secureStorage: true,
			},
		},
		{
			name: "remove scope but no old scope",
			cfgHosts: []string{
				literal_7150,
			},
			opts: &RefreshOptions{
				Hostname:     literal_7150,
				RemoveScopes: []string{"delete_repo"},
			},
			wantAuthArgs: authArgs{
				hostname:      literal_7150,
				scopes:        []string{},
				secureStorage: true,
			},
		},
		{
			name: "remove and add scopes at the same time",
			cfgHosts: []string{
				literal_7150,
			},
			oldScopes: literal_0752,
			opts: &RefreshOptions{
				Scopes:       []string{literal_6285, literal_7562, "workflow"},
				RemoveScopes: []string{"codespace", literal_6285, "workflow"},
			},
			wantAuthArgs: authArgs{
				hostname:      literal_7150,
				scopes:        []string{"delete_repo", literal_7562},
				secureStorage: true,
			},
		},
		{
			name: "remove scopes that don't exist",
			cfgHosts: []string{
				literal_7150,
			},
			oldScopes: literal_0752,
			opts: &RefreshOptions{
				RemoveScopes: []string{"codespace", literal_6285, literal_7562},
			},
			wantAuthArgs: authArgs{
				hostname:      literal_7150,
				scopes:        []string{"delete_repo"},
				secureStorage: true,
			},
		},
		{
			name: "errors when active user does not match user returned by auth flow",
			cfgHosts: []string{
				literal_7150,
			},
			authOut: authOut{
				username: "not-test-user",
				token:    "xyz456",
			},
			opts:    &RefreshOptions{},
			wantErr: "error refreshing credentials for test-user, received credentials for not-test-user, did you use the correct account in the browser?",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			aa := authArgs{}
			tt.opts.AuthFlow = func(_ *iostreams.IOStreams, hostname string, scopes []string, interactive bool) (token, username, error) {
				aa.hostname = hostname
				aa.scopes = scopes
				aa.interactive = interactive
				if tt.authOut != (authOut{}) {
					return token(tt.authOut.token), username(tt.authOut.username), tt.authOut.err
				}
				return token("xyz456"), username(literal_9478), nil
			}

			cfg, _ := config.NewIsolatedTestConfig(t)
			for _, hostname := range tt.cfgHosts {
				_, err := cfg.Authentication().Login(hostname, literal_9478, "abc123", "https", false)
				require.NoError(t, err)
			}
			tt.opts.Config = func() (config.Config, error) {
				return cfg, nil
			}

			ios, _, _, _ := iostreams.Test()
			ios.SetStdinTTY(!tt.nontty)
			ios.SetStdoutTTY(!tt.nontty)
			tt.opts.IO = ios

			httpReg := &httpmock.Registry{}
			httpReg.Register(
				httpmock.REST("GET", ""),
				func(req *http.Request) (*http.Response, error) {
					statusCode := 200
					if req.Header.Get("Authorization") != "token abc123" {
						statusCode = 400
					}
					return &http.Response{
						Request:    req,
						StatusCode: statusCode,
						Body:       io.NopCloser(strings.NewReader(``)),
						Header: http.Header{
							"X-Oauth-Scopes": {tt.oldScopes},
						},
					}, nil
				},
			)
			tt.opts.HttpClient = &http.Client{Transport: httpReg}

			pm := &prompter.PrompterMock{}
			if tt.prompterStubs != nil {
				tt.prompterStubs(pm)
			}
			tt.opts.Prompter = pm

			err := refreshRun(tt.opts)
			if tt.wantErr != "" {
				require.Contains(t, err.Error(), tt.wantErr)
				return
			}

			require.NoError(t, err)

			require.Equal(t, tt.wantAuthArgs.hostname, aa.hostname)
			require.Equal(t, tt.wantAuthArgs.scopes, aa.scopes)
			require.Equal(t, tt.wantAuthArgs.interactive, aa.interactive)

			authCfg := cfg.Authentication()
			activeUser, _ := authCfg.ActiveUser(aa.hostname)
			activeToken, _ := authCfg.ActiveToken(aa.hostname)
			require.Equal(t, literal_9478, activeUser)
			require.Equal(t, "xyz456", activeToken)
		})
	}
}

const literal_6423 = "-h aline.cedrac"

const literal_6285 = "repo:invite"

const literal_3795 = "read:public_key"

const literal_7150 = "github.com"

const literal_2873 = "obed.morton"

const literal_7562 = "public_key:read"

const literal_0752 = "repo:invite, delete_repo, codespace"

const literal_9478 = "test-user"
