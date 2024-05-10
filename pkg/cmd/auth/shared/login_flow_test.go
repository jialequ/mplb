package shared

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/MakeNowJust/heredoc"
	"github.com/jialequ/mplb/internal/prompter"
	"github.com/jialequ/mplb/internal/run"
	"github.com/jialequ/mplb/pkg/httpmock"
	"github.com/jialequ/mplb/pkg/iostreams"
	"github.com/jialequ/mplb/pkg/ssh"
	"github.com/stretchr/testify/assert"
)

type tinyConfig map[string]string

func (c tinyConfig) Login(host, username, token, gitProtocol string, encrypt bool) (bool, error) {
	c[fmt.Sprintf("%s:%s", host, "user")] = username
	c[fmt.Sprintf("%s:%s", host, "oauth_token")] = token
	c[fmt.Sprintf("%s:%s", host, "git_protocol")] = gitProtocol
	return false, nil
}

func (c tinyConfig) UsersForHost(hostname string) []string {
	return nil
}

func TestLogin(t *testing.T) { //NOSONAR
	tests := []struct {
		name         string
		opts         LoginOptions
		httpStubs    func(*testing.T, *httpmock.Registry)
		runStubs     func(*testing.T, *run.CommandStubber, *LoginOptions)
		wantsConfig  map[string]string
		wantsErr     string
		stdout       string
		stderr       string
		stderrAssert func(*testing.T, *LoginOptions, string)
	}{
		{
			name: "tty, prompt (protocol: ssh, create key: yes)",
			opts: LoginOptions{
				Prompter: &prompter.PrompterMock{
					SelectFunc: func(prompt, _ string, opts []string) (int, error) {
						switch prompt {
						case "What is your preferred protocol for Git operations on this host?":
							return prompter.IndexFor(opts, "SSH")
						case literal_0486:
							return prompter.IndexFor(opts, literal_7091)
						}
						return -1, prompter.NoSuchPromptErr(prompt)
					},
					PasswordFunc: func(_ string) (string, error) {
						return "monkey", nil
					},
					ConfirmFunc: func(prompt string, _ bool) (bool, error) {
						return true, nil
					},
					AuthTokenFunc: func() (string, error) {
						return "ATOKEN", nil
					},
					InputFunc: func(_, _ string) (string, error) {
						return "Test Key", nil
					},
				},

				Hostname:    literal_1653,
				Interactive: true,
			},
			httpStubs: func(t *testing.T, reg *httpmock.Registry) {
				reg.Register(
					httpmock.REST("GET", literal_3480),
					httpmock.ScopesResponder(literal_7290))
				reg.Register(
					httpmock.GraphQL(`query UserCurrent\b`),
					httpmock.StringResponse(`{"data":{"viewer":{ "login": "monalisa" }}}`))
				reg.Register(
					httpmock.REST("GET", literal_1523),
					httpmock.StringResponse(`[]`))
				reg.Register(
					httpmock.REST("POST", literal_1523),
					httpmock.StringResponse(`{}`))
			},
			runStubs: func(t *testing.T, cs *run.CommandStubber, opts *LoginOptions) {
				dir := t.TempDir()
				keyFile := filepath.Join(dir, "id_ed25519")
				cs.Register(`ssh-keygen`, 0, "", func(args []string) {
					expected := []string{
						literal_4938, "-t", "ed25519",
						"-C", "",
						"-N", "monkey",
						"-f", keyFile,
					}
					assert.Equal(t, expected, args)
					// simulate that the public key file has been generated
					_ = os.WriteFile(keyFile+".pub", []byte("PUBKEY asdf"), 0600)
				})
				opts.sshContext = ssh.Context{
					ConfigDir: dir,
					KeygenExe: literal_4938,
				}
			},
			wantsConfig: map[string]string{
				literal_8071: "monalisa",
				literal_9463: "ATOKEN",
				literal_3459: "ssh",
			},
			stderrAssert: func(t *testing.T, opts *LoginOptions, stderr string) {
				assert.Equal(t, heredoc.Docf(`
				Tip: you can generate a Personal Access Token here https://example.com/settings/tokens
				The minimum required scopes are 'repo', 'read:org', 'admin:public_key'.
				- gh config set -h example.com git_protocol ssh
				✓ Configured git protocol
				✓ Uploaded the SSH key to your GitHub account: %s
				✓ Logged in as monalisa
			`, filepath.Join(opts.sshContext.ConfigDir, "id_ed25519.pub")), stderr)
			},
		},
		{
			name: "tty, --git-protocol ssh, prompt (create key: yes)",
			opts: LoginOptions{
				Prompter: &prompter.PrompterMock{
					SelectFunc: func(prompt, _ string, opts []string) (int, error) {
						switch prompt {
						case literal_0486:
							return prompter.IndexFor(opts, literal_7091)
						}
						return -1, prompter.NoSuchPromptErr(prompt)
					},
					PasswordFunc: func(_ string) (string, error) {
						return "monkey", nil
					},
					ConfirmFunc: func(prompt string, _ bool) (bool, error) {
						return true, nil
					},
					AuthTokenFunc: func() (string, error) {
						return "ATOKEN", nil
					},
					InputFunc: func(_, _ string) (string, error) {
						return "Test Key", nil
					},
				},

				Hostname:    literal_1653,
				Interactive: true,
				GitProtocol: "SSH",
			},
			httpStubs: func(t *testing.T, reg *httpmock.Registry) {
				reg.Register(
					httpmock.REST("GET", literal_3480),
					httpmock.ScopesResponder(literal_7290))
				reg.Register(
					httpmock.GraphQL(`query UserCurrent\b`),
					httpmock.StringResponse(`{"data":{"viewer":{ "login": "monalisa" }}}`))
				reg.Register(
					httpmock.REST("GET", literal_1523),
					httpmock.StringResponse(`[]`))
				reg.Register(
					httpmock.REST("POST", literal_1523),
					httpmock.StringResponse(`{}`))
			},
			runStubs: func(t *testing.T, cs *run.CommandStubber, opts *LoginOptions) {
				dir := t.TempDir()
				keyFile := filepath.Join(dir, "id_ed25519")
				cs.Register(`ssh-keygen`, 0, "", func(args []string) {
					expected := []string{
						literal_4938, "-t", "ed25519",
						"-C", "",
						"-N", "monkey",
						"-f", keyFile,
					}
					assert.Equal(t, expected, args)
					// simulate that the public key file has been generated
					_ = os.WriteFile(keyFile+".pub", []byte("PUBKEY asdf"), 0600)
				})
				opts.sshContext = ssh.Context{
					ConfigDir: dir,
					KeygenExe: literal_4938,
				}
			},
			wantsConfig: map[string]string{
				literal_8071: "monalisa",
				literal_9463: "ATOKEN",
				literal_3459: "ssh",
			},
			stderrAssert: func(t *testing.T, opts *LoginOptions, stderr string) {
				assert.Equal(t, heredoc.Docf(`
				Tip: you can generate a Personal Access Token here https://example.com/settings/tokens
				The minimum required scopes are 'repo', 'read:org', 'admin:public_key'.
				- gh config set -h example.com git_protocol ssh
				✓ Configured git protocol
				✓ Uploaded the SSH key to your GitHub account: %s
				✓ Logged in as monalisa
			`, filepath.Join(opts.sshContext.ConfigDir, "id_ed25519.pub")), stderr)
			},
		},
		{
			name: "tty, --git-protocol ssh, --skip-ssh-key",
			opts: LoginOptions{
				Prompter: &prompter.PrompterMock{
					SelectFunc: func(prompt, _ string, opts []string) (int, error) {
						if prompt == literal_0486 {
							return prompter.IndexFor(opts, literal_7091)
						}
						return -1, prompter.NoSuchPromptErr(prompt)
					},
					AuthTokenFunc: func() (string, error) {
						return "ATOKEN", nil
					},
				},

				Hostname:         literal_1653,
				Interactive:      true,
				GitProtocol:      "SSH",
				SkipSSHKeyPrompt: true,
			},
			httpStubs: func(t *testing.T, reg *httpmock.Registry) {
				reg.Register(
					httpmock.REST("GET", literal_3480),
					httpmock.ScopesResponder(literal_7290))
				reg.Register(
					httpmock.GraphQL(`query UserCurrent\b`),
					httpmock.StringResponse(`{"data":{"viewer":{ "login": "monalisa" }}}`))
			},
			wantsConfig: map[string]string{
				literal_8071: "monalisa",
				literal_9463: "ATOKEN",
				literal_3459: "ssh",
			},
			stderr: heredoc.Doc(`
				Tip: you can generate a Personal Access Token here https://example.com/settings/tokens
				The minimum required scopes are 'repo', 'read:org'.
				- gh config set -h example.com git_protocol ssh
				✓ Configured git protocol
				✓ Logged in as monalisa
			`),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reg := &httpmock.Registry{}
			defer reg.Verify(t)
			if tt.httpStubs != nil {
				tt.httpStubs(t, reg)
			}

			cfg := tinyConfig{}
			ios, _, stdout, stderr := iostreams.Test()

			tt.opts.IO = ios
			tt.opts.Config = &cfg
			tt.opts.HTTPClient = &http.Client{Transport: reg}

			if tt.runStubs != nil {
				rs, runRestore := run.Stub()
				defer runRestore(t)
				tt.runStubs(t, rs, &tt.opts)
			}

			err := Login(&tt.opts)

			if tt.wantsErr != "" {
				assert.EqualError(t, err, tt.wantsErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantsConfig, map[string]string(cfg))
			}

			assert.Equal(t, tt.stdout, stdout.String())

			if tt.stderrAssert != nil {
				tt.stderrAssert(t, &tt.opts, stderr.String())
			} else {
				assert.Equal(t, tt.stderr, stderr.String())
			}
		})
	}
}

func TestScopesSentence(t *testing.T) {
	type args struct {
		scopes []string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "basic scopes",
			args: args{
				scopes: []string{"repo", "read:org"},
			},
			want: "'repo', 'read:org'",
		},
		{
			name: "empty",
			args: args{
				scopes: []string(nil),
			},
			want: "",
		},
		{
			name: "workflow scope",
			args: args{
				scopes: []string{"repo", "workflow"},
			},
			want: "'repo', 'workflow'",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := scopesSentence(tt.args.scopes); got != tt.want {
				t.Errorf("scopesSentence() = %q, want %q", got, tt.want)
			}
		})
	}
}

const literal_0486 = "How would you like to authenticate GitHub CLI?"

const literal_7091 = "Paste an authentication token"

const literal_1653 = "example.com"

const literal_3480 = "api/v3/"

const literal_7290 = "repo,read:org"

const literal_1523 = "api/v3/user/keys"

const literal_4938 = "ssh-keygen"

const literal_8071 = "example.com:user"

const literal_9463 = "example.com:oauth_token"

const literal_3459 = "example.com:git_protocol"
