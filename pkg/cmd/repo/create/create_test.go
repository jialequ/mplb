package create

import (
	"bytes"
	"fmt"
	"net/http"
	"testing"

	"github.com/cenkalti/backoff/v4"
	"github.com/google/shlex"
	"github.com/jialequ/mplb/git"
	"github.com/jialequ/mplb/internal/config"
	"github.com/jialequ/mplb/internal/prompter"
	"github.com/jialequ/mplb/internal/run"
	"github.com/jialequ/mplb/pkg/cmdutil"
	"github.com/jialequ/mplb/pkg/httpmock"
	"github.com/jialequ/mplb/pkg/iostreams"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCmdCreate(t *testing.T) {
	tests := []struct {
		name      string
		tty       bool
		cli       string
		wantsErr  bool
		errMsg    string
		wantsOpts CreateOptions
	}{
		{
			name:      "no args tty",
			tty:       true,
			cli:       "",
			wantsOpts: CreateOptions{Interactive: true},
		},
		{
			name:     "no args no-tty",
			tty:      false,
			cli:      "",
			wantsErr: true,
			errMsg:   "at least one argument required in non-interactive mode",
		},
		{
			name: "new repo from remote",
			cli:  "NEWREPO --public --clone",
			wantsOpts: CreateOptions{
				Name:   "NEWREPO",
				Public: true,
				Clone:  true},
		},
		{
			name:     "no visibility",
			tty:      true,
			cli:      "NEWREPO",
			wantsErr: true,
			errMsg:   "`--public`, `--private`, or `--internal` required when not running interactively",
		},
		{
			name:     "multiple visibility",
			tty:      true,
			cli:      "NEWREPO --public --private",
			wantsErr: true,
			errMsg:   "expected exactly one of `--public`, `--private`, or `--internal`",
		},
		{
			name: "new remote from local",
			cli:  "--source=/path/to/repo --private",
			wantsOpts: CreateOptions{
				Private: true,
				Source:  literal_7256},
		},
		{
			name: "new remote from local with remote",
			cli:  "--source=/path/to/repo --public --remote upstream",
			wantsOpts: CreateOptions{
				Public: true,
				Source: literal_7256,
				Remote: "upstream",
			},
		},
		{
			name: "new remote from local with push",
			cli:  "--source=/path/to/repo --push --public",
			wantsOpts: CreateOptions{
				Public: true,
				Source: literal_7256,
				Push:   true,
			},
		},
		{
			name: "new remote from local without visibility",
			cli:  "--source=/path/to/repo --push",
			wantsOpts: CreateOptions{
				Source: literal_7256,
				Push:   true,
			},
			wantsErr: true,
			errMsg:   "`--public`, `--private`, or `--internal` required when not running interactively",
		},
		{
			name:     "source with template",
			cli:      "--source=/path/to/repo --private --template mytemplate",
			wantsErr: true,
			errMsg:   "the `--source` option is not supported with `--clone`, `--template`, `--license`, or `--gitignore`",
		},
		{
			name:     "include all branches without template",
			cli:      "--source=/path/to/repo --private --include-all-branches",
			wantsErr: true,
			errMsg:   "the `--include-all-branches` option is only supported when using `--template`",
		},
		{
			name: "new remote from template with include all branches",
			cli:  "template-repo --template https://github.com/OWNER/REPO --public --include-all-branches",
			wantsOpts: CreateOptions{
				Name:               "template-repo",
				Public:             true,
				Template:           "https://github.com/OWNER/REPO",
				IncludeAllBranches: true,
			},
		},
		{
			name:     "template with .gitignore",
			cli:      "template-repo --template mytemplate --gitignore ../.gitignore --public",
			wantsErr: true,
			errMsg:   ".gitignore and license templates are not added when template is provided",
		},
		{
			name:     "template with license",
			cli:      "template-repo --template mytemplate --license ../.license --public",
			wantsErr: true,
			errMsg:   ".gitignore and license templates are not added when template is provided",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ios, _, _, _ := iostreams.Test()
			ios.SetStdinTTY(tt.tty)
			ios.SetStdoutTTY(tt.tty)

			f := &cmdutil.Factory{
				IOStreams: ios,
			}

			var opts *CreateOptions
			cmd := NewCmdCreate(f, func(o *CreateOptions) error {
				opts = o
				return nil
			})

			// TODO STUPID HACK
			// cobra aggressively adds help to all commands. since we're not running through the root command
			// (which manages help when running for real) and since create has a '-h' flag (for homepage),
			// cobra blows up when it tried to add a help flag and -h is already in use. This hack adds a
			// dummy help flag with a random shorthand to get around this.
			cmd.Flags().BoolP("help", "x", false, "")

			args, err := shlex.Split(tt.cli)
			require.NoError(t, err)
			cmd.SetArgs(args)
			cmd.SetIn(&bytes.Buffer{})
			cmd.SetOut(&bytes.Buffer{})
			cmd.SetErr(&bytes.Buffer{})

			_, err = cmd.ExecuteC()
			if tt.wantsErr {
				assert.Error(t, err)
				assert.Equal(t, tt.errMsg, err.Error())
				return
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, tt.wantsOpts.Interactive, opts.Interactive)
			assert.Equal(t, tt.wantsOpts.Source, opts.Source)
			assert.Equal(t, tt.wantsOpts.Name, opts.Name)
			assert.Equal(t, tt.wantsOpts.Public, opts.Public)
			assert.Equal(t, tt.wantsOpts.Internal, opts.Internal)
			assert.Equal(t, tt.wantsOpts.Private, opts.Private)
			assert.Equal(t, tt.wantsOpts.Clone, opts.Clone)
		})
	}
}

func TestCreateRun(t *testing.T) {
	tests := []struct {
		name        string
		tty         bool
		opts        *CreateOptions
		httpStubs   func(*httpmock.Registry)
		promptStubs func(*prompter.PrompterMock)
		execStubs   func(*run.CommandStubber)
		wantStdout  string
		wantErr     bool
		errMsg      string
	}{
		{
			name:       "interactive create from scratch with gitignore and license",
			opts:       &CreateOptions{Interactive: true},
			tty:        true,
			wantStdout: literal_1294,
			promptStubs: func(p *prompter.PrompterMock) {
				p.ConfirmFunc = func(message string, defaultValue bool) (bool, error) {
					switch message {
					case literal_1654:
						return false, nil
					case literal_6345:
						return true, nil
					case literal_6572:
						return true, nil
					case `This will create "REPO" as a private repository on GitHub. Continue?`:
						return defaultValue, nil
					case literal_9725:
						return defaultValue, nil
					default:
						return false, fmt.Errorf(literal_6789, message)
					}
				}
				p.InputFunc = func(message, defaultValue string) (string, error) {
					switch message {
					case literal_8176:
						return "REPO", nil
					case "Description":
						return literal_8734, nil
					default:
						return "", fmt.Errorf(literal_7928, message)
					}
				}
				p.SelectFunc = func(message, defaultValue string, options []string) (int, error) {
					switch message {
					case literal_7231:
						return prompter.IndexFor(options, literal_7935)
					case "Visibility":
						return prompter.IndexFor(options, "Private")
					case "Choose a license":
						return prompter.IndexFor(options, "GNU Lesser General Public License v3.0")
					case "Choose a .gitignore template":
						return prompter.IndexFor(options, "Go")
					default:
						return 0, fmt.Errorf(literal_7910, message)
					}
				}
			},
			httpStubs: func(reg *httpmock.Registry) {
				reg.Register(
					httpmock.GraphQL(`query UserCurrent\b`),
					httpmock.StringResponse(`{"data":{"viewer":{"login":"someuser","organizations":{"nodes": []}}}}`))
				reg.Register(
					httpmock.REST("GET", "gitignore/templates"),
					httpmock.StringResponse(`["Actionscript","Android","AppceleratorTitanium","Autotools","Bancha","C","C++","Go"]`))
				reg.Register(
					httpmock.REST("GET", "licenses"),
					httpmock.StringResponse(`[{"key": "mit","name": "MIT License"},{"key": "lgpl-3.0","name": "GNU Lesser General Public License v3.0"}]`))
				reg.Register(
					httpmock.REST("POST", "user/repos"),
					httpmock.StringResponse(`{"name":"REPO", "owner":{"login": "OWNER"}, "html_url":"https://github.com/OWNER/REPO"}`))

			},
			execStubs: func(cs *run.CommandStubber) {
				cs.Register(`git clone https://github.com/OWNER/REPO.git`, 0, "")
			},
		},
		{
			name:       "interactive create from scratch but with prompted owner",
			opts:       &CreateOptions{Interactive: true},
			tty:        true,
			wantStdout: "✓ Created repository org1/REPO on GitHub\n  https://github.com/org1/REPO\n",
			promptStubs: func(p *prompter.PrompterMock) {
				p.ConfirmFunc = func(message string, defaultValue bool) (bool, error) {
					switch message {
					case literal_1654:
						return false, nil
					case literal_6345:
						return false, nil
					case literal_6572:
						return false, nil
					case `This will create "org1/REPO" as a private repository on GitHub. Continue?`:
						return true, nil
					case literal_9725:
						return false, nil
					default:
						return false, fmt.Errorf(literal_6789, message)
					}
				}
				p.InputFunc = func(message, defaultValue string) (string, error) {
					switch message {
					case literal_8176:
						return "REPO", nil
					case "Description":
						return literal_8734, nil
					default:
						return "", fmt.Errorf(literal_7928, message)
					}
				}
				p.SelectFunc = func(message, defaultValue string, options []string) (int, error) {
					switch message {
					case "Repository owner":
						return prompter.IndexFor(options, "org1")
					case literal_7231:
						return prompter.IndexFor(options, literal_7935)
					case "Visibility":
						return prompter.IndexFor(options, "Private")
					default:
						return 0, fmt.Errorf(literal_7910, message)
					}
				}
			},
			httpStubs: func(reg *httpmock.Registry) {
				reg.Register(
					httpmock.GraphQL(`query UserCurrent\b`),
					httpmock.StringResponse(`{"data":{"viewer":{"login":"someuser","organizations":{"nodes": [{"login": "org1"}, {"login": "org2"}]}}}}`))
				reg.Register(
					httpmock.REST("GET", "users/org1"),
					httpmock.StringResponse(`{"login":"org1","type":"Organization"}`))
				reg.Register(
					httpmock.GraphQL(`mutation RepositoryCreate\b`),
					httpmock.StringResponse(`
					{
						"data": {
							"createRepository": {
								"repository": {
									"id": "REPOID",
									"name": "REPO",
									"owner": {"login":"org1"},
									"url": "https://github.com/org1/REPO"
								}
							}
						}
					}`))
			},
		},
		{
			name: "interactive create from scratch but cancel before submit",
			opts: &CreateOptions{Interactive: true},
			tty:  true,
			promptStubs: func(p *prompter.PrompterMock) {
				p.ConfirmFunc = func(message string, defaultValue bool) (bool, error) {
					switch message {
					case literal_1654:
						return false, nil
					case literal_6345:
						return false, nil
					case literal_6572:
						return false, nil
					case `This will create "REPO" as a private repository on GitHub. Continue?`:
						return false, nil
					default:
						return false, fmt.Errorf(literal_6789, message)
					}
				}
				p.InputFunc = func(message, defaultValue string) (string, error) {
					switch message {
					case literal_8176:
						return "REPO", nil
					case "Description":
						return literal_8734, nil
					default:
						return "", fmt.Errorf(literal_7928, message)
					}
				}
				p.SelectFunc = func(message, defaultValue string, options []string) (int, error) {
					switch message {
					case literal_7231:
						return prompter.IndexFor(options, literal_7935)
					case "Visibility":
						return prompter.IndexFor(options, "Private")
					default:
						return 0, fmt.Errorf(literal_7910, message)
					}
				}
			},
			httpStubs: func(reg *httpmock.Registry) {
				reg.Register(
					httpmock.GraphQL(`query UserCurrent\b`),
					httpmock.StringResponse(`{"data":{"viewer":{"login":"someuser","organizations":{"nodes": []}}}}`))
			},
			wantStdout: "",
			wantErr:    true,
			errMsg:     "CancelError",
		},
		{
			name: "interactive with existing repository public",
			opts: &CreateOptions{Interactive: true},
			tty:  true,
			promptStubs: func(p *prompter.PrompterMock) {
				p.ConfirmFunc = func(message string, defaultValue bool) (bool, error) {
					switch message {
					case "Add a remote?":
						return false, nil
					default:
						return false, fmt.Errorf(literal_6789, message)
					}
				}
				p.InputFunc = func(message, defaultValue string) (string, error) {
					switch message {
					case "Path to local repository":
						return defaultValue, nil
					case literal_8176:
						return "REPO", nil
					case "Description":
						return literal_8734, nil
					default:
						return "", fmt.Errorf(literal_7928, message)
					}
				}
				p.SelectFunc = func(message, defaultValue string, options []string) (int, error) {
					switch message {
					case literal_7231:
						return prompter.IndexFor(options, "Push an existing local repository to GitHub")
					case "Visibility":
						return prompter.IndexFor(options, "Private")
					default:
						return 0, fmt.Errorf(literal_7910, message)
					}
				}
			},
			httpStubs: func(reg *httpmock.Registry) {
				reg.Register(
					httpmock.GraphQL(`query UserCurrent\b`),
					httpmock.StringResponse(`{"data":{"viewer":{"login":"someuser","organizations":{"nodes": []}}}}`))
				reg.Register(
					httpmock.GraphQL(`mutation RepositoryCreate\b`),
					httpmock.StringResponse(`
					{
						"data": {
							"createRepository": {
								"repository": {
									"id": "REPOID",
									"name": "REPO",
									"owner": {"login":"OWNER"},
									"url": "https://github.com/OWNER/REPO"
								}
							}
						}
					}`))
			},
			execStubs: func(cs *run.CommandStubber) {
				cs.Register(`git -C . rev-parse --git-dir`, 0, ".git")
				cs.Register(`git -C . rev-parse HEAD`, 0, "commithash")
			},
			wantStdout: literal_1294,
		},
		{
			name: "interactive with existing repository public add remote and push",
			opts: &CreateOptions{Interactive: true},
			tty:  true,
			promptStubs: func(p *prompter.PrompterMock) {
				p.ConfirmFunc = func(message string, defaultValue bool) (bool, error) {
					switch message {
					case "Add a remote?":
						return true, nil
					case `Would you like to push commits from the current branch to "origin"?`:
						return true, nil
					default:
						return false, fmt.Errorf(literal_6789, message)
					}
				}
				p.InputFunc = func(message, defaultValue string) (string, error) {
					switch message {
					case "Path to local repository":
						return defaultValue, nil
					case literal_8176:
						return "REPO", nil
					case "Description":
						return literal_8734, nil
					case "What should the new remote be called?":
						return defaultValue, nil
					default:
						return "", fmt.Errorf(literal_7928, message)
					}
				}
				p.SelectFunc = func(message, defaultValue string, options []string) (int, error) {
					switch message {
					case literal_7231:
						return prompter.IndexFor(options, "Push an existing local repository to GitHub")
					case "Visibility":
						return prompter.IndexFor(options, "Private")
					default:
						return 0, fmt.Errorf(literal_7910, message)
					}
				}
			},
			httpStubs: func(reg *httpmock.Registry) {
				reg.Register(
					httpmock.GraphQL(`query UserCurrent\b`),
					httpmock.StringResponse(`{"data":{"viewer":{"login":"someuser","organizations":{"nodes": []}}}}`))
				reg.Register(
					httpmock.GraphQL(`mutation RepositoryCreate\b`),
					httpmock.StringResponse(`
					{
						"data": {
							"createRepository": {
								"repository": {
									"id": "REPOID",
									"name": "REPO",
									"owner": {"login":"OWNER"},
									"url": "https://github.com/OWNER/REPO"
								}
							}
						}
					}`))
			},
			execStubs: func(cs *run.CommandStubber) {
				cs.Register(`git -C . rev-parse --git-dir`, 0, ".git")
				cs.Register(`git -C . rev-parse HEAD`, 0, "commithash")
				cs.Register(`git -C . remote add origin https://github.com/OWNER/REPO`, 0, "")
				cs.Register(`git -C . push --set-upstream origin HEAD`, 0, "")
			},
			wantStdout: "✓ Created repository OWNER/REPO on GitHub\n  https://github.com/OWNER/REPO\n✓ Added remote https://github.com/OWNER/REPO.git\n✓ Pushed commits to https://github.com/OWNER/REPO.git\n",
		},
		{
			name: "interactive create from a template repository",
			opts: &CreateOptions{Interactive: true},
			tty:  true,
			promptStubs: func(p *prompter.PrompterMock) {
				p.ConfirmFunc = func(message string, defaultValue bool) (bool, error) {
					switch message {
					case `This will create "OWNER/REPO" as a private repository on GitHub. Continue?`:
						return defaultValue, nil
					case literal_9725:
						return defaultValue, nil
					default:
						return false, fmt.Errorf(literal_6789, message)
					}
				}
				p.InputFunc = func(message, defaultValue string) (string, error) {
					switch message {
					case literal_8176:
						return "REPO", nil
					case "Description":
						return literal_8734, nil
					default:
						return "", fmt.Errorf(literal_7928, message)
					}
				}
				p.SelectFunc = func(message, defaultValue string, options []string) (int, error) {
					switch message {
					case "Repository owner":
						return prompter.IndexFor(options, "OWNER")
					case "Choose a template repository":
						return prompter.IndexFor(options, "REPO")
					case literal_7231:
						return prompter.IndexFor(options, "Create a new repository on GitHub from a template repository")
					case "Visibility":
						return prompter.IndexFor(options, "Private")
					default:
						return 0, fmt.Errorf(literal_7910, message)
					}
				}
			},
			httpStubs: func(reg *httpmock.Registry) {
				reg.Register(
					httpmock.GraphQL(`query UserCurrent\b`),
					httpmock.StringResponse(`{"data":{"viewer":{"login":"OWNER","organizations":{"nodes": []}}}}`))
				reg.Register(
					httpmock.GraphQL(`query UserCurrent\b`),
					httpmock.StringResponse(`{"data":{"viewer":{"login":"OWNER","organizations":{"nodes": []}}}}`))
				reg.Register(
					httpmock.GraphQL(`query RepositoryList\b`),
					httpmock.FileResponse("./fixtures/repoTempList.json"))
				reg.Register(
					httpmock.REST("GET", "users/OWNER"),
					httpmock.StringResponse(`{"login":"OWNER","type":"User"}`))
				reg.Register(
					httpmock.GraphQL(`query UserCurrent\b`),
					httpmock.StringResponse(`{"data":{"viewer":{"id":"OWNER"}}}`))
				reg.Register(
					httpmock.GraphQL(`mutation CloneTemplateRepository\b`),
					httpmock.StringResponse(`
						{
							"data": {
								"cloneTemplateRepository": {
									"repository": {
										"id": "REPOID",
										"name": "REPO",
										"owner": {"login":"OWNER"},
										"url": "https://github.com/OWNER/REPO"
									}
								}
							}
						}`))
			},
			execStubs: func(cs *run.CommandStubber) {
				cs.Register(`git clone --branch main https://github.com/OWNER/REPO`, 0, "")
			},
			wantStdout: literal_1294,
		},
		{
			name: "interactive create from template repo but there are no template repos",
			opts: &CreateOptions{Interactive: true},
			tty:  true,
			promptStubs: func(p *prompter.PrompterMock) {
				p.ConfirmFunc = func(message string, defaultValue bool) (bool, error) {
					switch message {
					default:
						return false, fmt.Errorf(literal_6789, message)
					}
				}
				p.InputFunc = func(message, defaultValue string) (string, error) {
					switch message {
					case literal_8176:
						return "REPO", nil
					case "Description":
						return literal_8734, nil
					default:
						return "", fmt.Errorf(literal_7928, message)
					}
				}
				p.SelectFunc = func(message, defaultValue string, options []string) (int, error) {
					switch message {
					case literal_7231:
						return prompter.IndexFor(options, "Create a new repository on GitHub from a template repository")
					case "Visibility":
						return prompter.IndexFor(options, "Private")
					default:
						return 0, fmt.Errorf(literal_7910, message)
					}
				}
			},
			httpStubs: func(reg *httpmock.Registry) {
				reg.Register(
					httpmock.GraphQL(`query UserCurrent\b`),
					httpmock.StringResponse(`{"data":{"viewer":{"login":"OWNER","organizations":{"nodes": []}}}}`))
				reg.Register(
					httpmock.GraphQL(`query UserCurrent\b`),
					httpmock.StringResponse(`{"data":{"viewer":{"login":"OWNER","organizations":{"nodes": []}}}}`))
				reg.Register(
					httpmock.GraphQL(`query RepositoryList\b`),
					httpmock.StringResponse(`{"data":{"repositoryOwner":{"login":"OWNER","repositories":{"nodes":[]},"totalCount":0,"pageInfo":{"hasNextPage":false,"endCursor":""}}}}`))
			},
			execStubs:  func(cs *run.CommandStubber) {},
			wantStdout: "",
			wantErr:    true,
			errMsg:     "OWNER has no template repositories",
		},
		{
			name: "noninteractive create from scratch",
			opts: &CreateOptions{
				Interactive: false,
				Name:        "REPO",
				Visibility:  "PRIVATE",
			},
			tty: false,
			httpStubs: func(reg *httpmock.Registry) {
				reg.Register(
					httpmock.GraphQL(`mutation RepositoryCreate\b`),
					httpmock.StringResponse(`
					{
						"data": {
							"createRepository": {
								"repository": {
									"id": "REPOID",
									"name": "REPO",
									"owner": {"login":"OWNER"},
									"url": "https://github.com/OWNER/REPO"
								}
							}
						}
					}`))
			},
			wantStdout: literal_9370,
		},
		{
			name: "noninteractive create from source",
			opts: &CreateOptions{
				Interactive: false,
				Source:      ".",
				Name:        "REPO",
				Visibility:  "PRIVATE",
			},
			tty: false,
			httpStubs: func(reg *httpmock.Registry) {
				reg.Register(
					httpmock.GraphQL(`mutation RepositoryCreate\b`),
					httpmock.StringResponse(`
					{
						"data": {
							"createRepository": {
								"repository": {
									"id": "REPOID",
									"name": "REPO",
									"owner": {"login":"OWNER"},
									"url": "https://github.com/OWNER/REPO"
								}
							}
						}
					}`))
			},
			execStubs: func(cs *run.CommandStubber) {
				cs.Register(`git -C . rev-parse --git-dir`, 0, ".git")
				cs.Register(`git -C . rev-parse HEAD`, 0, "commithash")
				cs.Register(`git -C . remote add origin https://github.com/OWNER/REPO`, 0, "")
			},
			wantStdout: literal_9370,
		},
		{
			name: "noninteractive clone from scratch",
			opts: &CreateOptions{
				Interactive: false,
				Name:        "REPO",
				Visibility:  "PRIVATE",
				Clone:       true,
			},
			tty: false,
			httpStubs: func(reg *httpmock.Registry) {
				reg.Register(
					httpmock.GraphQL(`mutation RepositoryCreate\b`),
					httpmock.StringResponse(`
					{
						"data": {
							"createRepository": {
								"repository": {
									"id": "REPOID",
									"name": "REPO",
									"owner": {"login":"OWNER"},
									"url": "https://github.com/OWNER/REPO"
								}
							}
						}
					}`))
			},
			execStubs: func(cs *run.CommandStubber) {
				cs.Register(`git init REPO`, 0, "")
				cs.Register(`git -C REPO remote add origin https://github.com/OWNER/REPO`, 0, "")
			},
			wantStdout: literal_9370,
		},
		{
			name: "noninteractive clone with readme",
			opts: &CreateOptions{
				Interactive: false,
				Name:        "ElliotAlderson",
				Visibility:  "PRIVATE",
				Clone:       true,
				AddReadme:   true,
			},
			tty: false,
			httpStubs: func(reg *httpmock.Registry) {
				reg.Register(
					httpmock.REST("POST", "user/repos"),
					httpmock.RESTPayload(200, "{\"name\":\"ElliotAlderson\", \"owner\":{\"login\": \"OWNER\"}, \"html_url\":\"https://github.com/OWNER/ElliotAlderson\"}",
						func(payload map[string]interface{}) {
							payload["name"] = "ElliotAlderson"
							payload["owner"] = map[string]interface{}{"login": "OWNER"}
							payload["auto_init"] = true
							payload["private"] = true
						},
					),
				)
			},
			execStubs: func(cs *run.CommandStubber) {
				cs.Register(`git clone https://github.com/OWNER/ElliotAlderson`, 128, "")
				cs.Register(`git clone https://github.com/OWNER/ElliotAlderson`, 0, "")
			},
			wantStdout: "https://github.com/OWNER/ElliotAlderson\n",
		},
		{
			name: "noninteractive create from template with retry",
			opts: &CreateOptions{
				Interactive: false,
				Name:        "REPO",
				Visibility:  "PRIVATE",
				Clone:       true,
				Template:    "mytemplate",
				BackOff:     &backoff.ZeroBackOff{},
			},
			tty: false,
			httpStubs: func(reg *httpmock.Registry) {
				// Test resolving repo owner from repo name only.
				reg.Register(
					httpmock.GraphQL(`query UserCurrent\b`),
					httpmock.StringResponse(`{"data":{"viewer":{"login":"OWNER"}}}`))
				reg.Register(
					httpmock.GraphQL(`query RepositoryInfo\b`),
					httpmock.GraphQLQuery(`{
						"data": {
							"repository": {
								"id": "REPOID",
								"defaultBranchRef": {
									"name": "main"
								}
							}
						}
					}`, func(s string, m map[string]interface{}) {
						assert.Equal(t, "OWNER", m["owner"])
						assert.Equal(t, "mytemplate", m["name"])
					}),
				)
				reg.Register(
					httpmock.GraphQL(`query UserCurrent\b`),
					httpmock.StringResponse(`{"data":{"viewer":{"id":"OWNERID"}}}`))
				reg.Register(
					httpmock.GraphQL(`mutation CloneTemplateRepository\b`),
					httpmock.GraphQLMutation(`
					{
						"data": {
							"cloneTemplateRepository": {
								"repository": {
									"id": "REPOID",
									"name": "REPO",
									"owner": {"login":"OWNER"},
									"url": "https://github.com/OWNER/REPO"
								}
							}
						}
					}`, func(m map[string]interface{}) {
						assert.Equal(t, "REPOID", m["repositoryId"])
					}))
			},
			execStubs: func(cs *run.CommandStubber) {
				// fatal: Remote branch main not found in upstream origin
				cs.Register(`git clone --branch main https://github.com/OWNER/REPO`, 128, "")
				cs.Register(`git clone --branch main https://github.com/OWNER/REPO`, 0, "")
			},
			wantStdout: literal_9370,
		},
	}
	for _, tt := range tests {
		prompterMock := &prompter.PrompterMock{}
		tt.opts.Prompter = prompterMock
		if tt.promptStubs != nil {
			tt.promptStubs(prompterMock)
		}

		reg := &httpmock.Registry{}
		if tt.httpStubs != nil {
			tt.httpStubs(reg)
		}
		tt.opts.HttpClient = func() (*http.Client, error) {
			return &http.Client{Transport: reg}, nil
		}
		tt.opts.Config = func() (config.Config, error) {
			return config.NewBlankConfig(), nil
		}

		tt.opts.GitClient = &git.Client{
			GhPath:  "some/path/gh",
			GitPath: "some/path/git",
		}

		ios, _, stdout, stderr := iostreams.Test()
		ios.SetStdinTTY(tt.tty)
		ios.SetStdoutTTY(tt.tty)
		tt.opts.IO = ios

		t.Run(tt.name, func(t *testing.T) {
			cs, restoreRun := run.Stub()
			defer restoreRun(t)
			if tt.execStubs != nil {
				tt.execStubs(cs)
			}

			defer reg.Verify(t)
			err := createRun(tt.opts)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, tt.errMsg, err.Error())
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.wantStdout, stdout.String())
			assert.Equal(t, "", stderr.String())
		})
	}
}

const literal_7256 = "/path/to/repo"

const literal_1294 = "✓ Created repository OWNER/REPO on GitHub\n  https://github.com/OWNER/REPO\n"

const literal_1654 = "Would you like to add a README file?"

const literal_6345 = "Would you like to add a .gitignore?"

const literal_6572 = "Would you like to add a license?"

const literal_9725 = "Clone the new repository locally?"

const literal_6789 = "unexpected confirm prompt: %s"

const literal_8176 = "Repository name"

const literal_8734 = "my new repo"

const literal_7928 = "unexpected input prompt: %s"

const literal_7231 = "What would you like to do?"

const literal_7935 = "Create a new repository on GitHub from scratch"

const literal_7910 = "unexpected select prompt: %s"

const literal_9370 = "https://github.com/OWNER/REPO\n"
