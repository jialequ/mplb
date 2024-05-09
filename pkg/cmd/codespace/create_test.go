package codespace

import (
	"bytes"
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/google/shlex"
	"github.com/jialequ/mplb/internal/browser"
	"github.com/jialequ/mplb/internal/codespaces/api"
	"github.com/jialequ/mplb/pkg/cmdutil"
	"github.com/jialequ/mplb/pkg/iostreams"
	"github.com/stretchr/testify/assert"
)

func TestCreateCmdFlagError(t *testing.T) {
	tests := []struct {
		name     string
		args     string
		wantsErr error
	}{
		{
			name:     "return error when using web flag with display-name, idle-timeout, or retention-period flags",
			args:     "--web --display-name foo --idle-timeout 30m",
			wantsErr: fmt.Errorf("using --web with --display-name, --idle-timeout, or --retention-period is not supported"),
		},
		{
			name:     "return error when using web flag with one of display-name, idle-timeout or retention-period flags",
			args:     "--web --idle-timeout 30m",
			wantsErr: fmt.Errorf("using --web with --display-name, --idle-timeout, or --retention-period is not supported"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ios, _, _, _ := iostreams.Test()
			a := &App{
				io: ios,
			}
			cmd := newCreateCmd(a)

			args, _ := shlex.Split(tt.args)
			cmd.SetArgs(args)
			cmd.SetIn(&bytes.Buffer{})
			cmd.SetOut(&bytes.Buffer{})
			cmd.SetErr(&bytes.Buffer{})

			_, err := cmd.ExecuteC()

			assert.Error(t, err)
			assert.EqualError(t, err, tt.wantsErr.Error())
		})
	}
}

func TestApp_Create(t *testing.T) {
	type fields struct {
		apiClient apiClient
	}
	tests := []struct {
		name       string
		fields     fields
		opts       createOptions
		wantErr    error
		wantStdout string
		wantStderr string
		wantURL    string
		isTTY      bool
	}{
		{
			name: "create codespace with default branch and 30m idle timeout",
			fields: fields{
				apiClient: apiCreateDefaults(&apiClientMock{
					CreateCodespaceFunc: func(ctx context.Context, params *api.CreateCodespaceParams) (*api.Codespace, error) {
						if params.Branch != "main" {
							return nil, fmt.Errorf(literal_7213, params.Branch, "main")
						}
						if params.IdleTimeoutMinutes != 30 {
							return nil, fmt.Errorf(literal_7594, params.IdleTimeoutMinutes)
						}
						if *params.RetentionPeriodMinutes != 2880 {
							return nil, fmt.Errorf("retention period minutes expected 2880, was %v", params.RetentionPeriodMinutes)
						}
						if params.DisplayName != "" {
							return nil, fmt.Errorf("display name was %q, expected empty", params.DisplayName)
						}
						return &api.Codespace{
							Name: literal_8194,
						}, nil
					},
				}),
			},
			opts: createOptions{
				repo:            literal_3498,
				branch:          "",
				machine:         "GIGA",
				showStatus:      false,
				idleTimeout:     30 * time.Minute,
				retentionPeriod: NullableDuration{durationPtr(48 * time.Hour)},
			},
			wantStdout: literal_9271,
			wantStderr: literal_6493,
		},
		{
			name: "create with explicit display name",
			fields: fields{
				apiClient: apiCreateDefaults(&apiClientMock{
					CreateCodespaceFunc: func(ctx context.Context, params *api.CreateCodespaceParams) (*api.Codespace, error) {
						if params.DisplayName != literal_6279 {
							return nil, fmt.Errorf("expected display name %q, got %q", literal_6279, params.DisplayName)
						}
						return &api.Codespace{
							Name: literal_8194,
						}, nil
					},
				}),
			},
			opts: createOptions{
				repo:        literal_3498,
				branch:      "main",
				displayName: literal_6279,
			},
			wantStdout: literal_9271,
			wantStderr: literal_6493,
		},
		{
			name: "create codespace with default branch shows idle timeout notice if present",
			fields: fields{
				apiClient: apiCreateDefaults(&apiClientMock{
					CreateCodespaceFunc: func(ctx context.Context, params *api.CreateCodespaceParams) (*api.Codespace, error) {
						if params.Branch != "main" {
							return nil, fmt.Errorf(literal_7213, params.Branch, "main")
						}
						if params.IdleTimeoutMinutes != 30 {
							return nil, fmt.Errorf(literal_7594, params.IdleTimeoutMinutes)
						}
						if params.RetentionPeriodMinutes != nil {
							return nil, fmt.Errorf("retention period minutes expected nil, was %v", params.RetentionPeriodMinutes)
						}
						if params.DevContainerPath != literal_8907 {
							return nil, fmt.Errorf(literal_0195, params.DevContainerPath, literal_8907)
						}
						return &api.Codespace{
							Name: literal_8194,
						}, nil
					},
				}),
			},
			opts: createOptions{
				repo:             literal_3498,
				branch:           "",
				machine:          "GIGA",
				showStatus:       false,
				idleTimeout:      30 * time.Minute,
				devContainerPath: literal_8907,
			},
			wantStdout: literal_9271,
			wantStderr: literal_6493,
		},
		{
			name: "create codespace with nonexistent machine results in error",
			fields: fields{
				apiClient: apiCreateDefaults(&apiClientMock{
					GetCodespacesMachinesFunc: func(ctx context.Context, repoID int, branch, location string, devcontainerPath string) ([]*api.Machine, error) {
						return []*api.Machine{
							{
								Name:        "GIGA",
								DisplayName: literal_9163,
							},
							{
								Name:        "TERA",
								DisplayName: "Terabits of a machine",
							},
						}, nil
					},
				}),
			},
			opts: createOptions{
				repo:    literal_3498,
				machine: "MEGA",
			},
			wantStderr: literal_6493,
			wantErr:    fmt.Errorf("error getting machine type: there is no such machine for the repository: %s\nAvailable machines: %v", "MEGA", []string{"GIGA", "TERA"}),
		},
		{
			name: "create codespace with display name more than 48 characters results in error",
			fields: fields{
				apiClient: apiCreateDefaults(&apiClientMock{
					CreateCodespaceFunc: func(ctx context.Context, params *api.CreateCodespaceParams) (*api.Codespace, error) {
						return &api.Codespace{
							Name: literal_8194,
						}, nil
					},
				}),
			},
			opts: createOptions{
				repo:        literal_3498,
				machine:     "GIGA",
				displayName: "this-is-very-long-display-name-with-49-characters",
			},
			wantStderr: literal_6493,
			wantErr:    fmt.Errorf("error creating codespace: display name should contain a maximum of %d characters", displayNameMaxLength),
		},
		{
			name: "create codespace with devcontainer path results in selecting the correct machine type",
			fields: fields{
				apiClient: apiCreateDefaults(&apiClientMock{
					GetCodespacesMachinesFunc: func(ctx context.Context, repoID int, branch, location string, devcontainerPath string) ([]*api.Machine, error) {
						if devcontainerPath == "" {
							return []*api.Machine{
								{
									Name:        "GIGA",
									DisplayName: literal_9163,
								},
							}, nil
						} else {
							return []*api.Machine{
								{
									Name:        "MEGA",
									DisplayName: "Megabits of a machine",
								},
								{
									Name:        "GIGA",
									DisplayName: literal_9163,
								},
							}, nil
						}
					},
					CreateCodespaceFunc: func(ctx context.Context, params *api.CreateCodespaceParams) (*api.Codespace, error) {
						if params.Branch != "main" {
							return nil, fmt.Errorf(literal_7213, params.Branch, "main")
						}
						if params.IdleTimeoutMinutes != 30 {
							return nil, fmt.Errorf(literal_7594, params.IdleTimeoutMinutes)
						}
						if params.RetentionPeriodMinutes != nil {
							return nil, fmt.Errorf("retention period minutes expected nil, was %v", params.RetentionPeriodMinutes)
						}
						if params.DevContainerPath != literal_8907 {
							return nil, fmt.Errorf(literal_0195, params.DevContainerPath, literal_8907)
						}
						if params.Machine != "MEGA" {
							return nil, fmt.Errorf("want machine %q, got %q", "MEGA", params.Machine)
						}
						return &api.Codespace{
							Name: literal_8194,
							Machine: api.CodespaceMachine{
								Name:        "MEGA",
								DisplayName: "Megabits of a machine",
							},
						}, nil
					},
				}),
			},
			opts: createOptions{
				repo:             literal_3498,
				branch:           "",
				machine:          "MEGA",
				showStatus:       false,
				idleTimeout:      30 * time.Minute,
				devContainerPath: literal_8907,
			},
			wantStdout: literal_9271,
			wantStderr: literal_6493,
		},
		{
			name: "create codespace with default branch with default devcontainer if no path provided and no devcontainer files exist in the repo",
			fields: fields{
				apiClient: apiCreateDefaults(&apiClientMock{
					ListDevContainersFunc: func(ctx context.Context, repoID int, branch string, limit int) ([]api.DevContainerEntry, error) {
						return []api.DevContainerEntry{}, nil
					},
					CreateCodespaceFunc: func(ctx context.Context, params *api.CreateCodespaceParams) (*api.Codespace, error) {
						if params.Branch != "main" {
							return nil, fmt.Errorf(literal_7213, params.Branch, "main")
						}
						if params.IdleTimeoutMinutes != 30 {
							return nil, fmt.Errorf(literal_7594, params.IdleTimeoutMinutes)
						}
						if params.DevContainerPath != "" {
							return nil, fmt.Errorf(literal_0195, params.DevContainerPath, literal_8907)
						}
						return &api.Codespace{
							Name:              literal_8194,
							IdleTimeoutNotice: "Idle timeout for this codespace is set to 10 minutes in compliance with your organization's policy",
						}, nil
					},
				}),
			},
			opts: createOptions{
				repo:        literal_3498,
				branch:      "",
				machine:     "GIGA",
				showStatus:  false,
				idleTimeout: 30 * time.Minute,
			},
			wantStdout: literal_9271,
			wantStderr: "  ✓ Codespaces usage for this repository is paid for by monalisa\nNotice: Idle timeout for this codespace is set to 10 minutes in compliance with your organization's policy\n",
			isTTY:      true,
		},
		{
			name: "returns error when getting devcontainer paths fails",
			fields: fields{
				apiClient: apiCreateDefaults(&apiClientMock{
					ListDevContainersFunc: func(ctx context.Context, repoID int, branch string, limit int) ([]api.DevContainerEntry, error) {
						return nil, fmt.Errorf("some error")
					},
				}),
			},
			opts: createOptions{
				repo:        literal_3498,
				branch:      "",
				machine:     "GIGA",
				showStatus:  false,
				idleTimeout: 30 * time.Minute,
			},
			wantErr:    fmt.Errorf("error getting devcontainer.json paths: some error"),
			wantStderr: literal_6493,
		},
		{
			name: "create codespace with default branch does not show idle timeout notice if not conntected to terminal",
			fields: fields{
				apiClient: apiCreateDefaults(&apiClientMock{
					CreateCodespaceFunc: func(ctx context.Context, params *api.CreateCodespaceParams) (*api.Codespace, error) {
						if params.Branch != "main" {
							return nil, fmt.Errorf(literal_7213, params.Branch, "main")
						}
						if params.IdleTimeoutMinutes != 30 {
							return nil, fmt.Errorf(literal_7594, params.IdleTimeoutMinutes)
						}
						return &api.Codespace{
							Name:              literal_8194,
							IdleTimeoutNotice: "Idle timeout for this codespace is set to 10 minutes in compliance with your organization's policy",
						}, nil
					},
				}),
			},
			opts: createOptions{
				repo:        literal_3498,
				branch:      "",
				machine:     "GIGA",
				showStatus:  false,
				idleTimeout: 30 * time.Minute,
			},
			wantStdout: literal_9271,
			wantStderr: literal_6493,
			isTTY:      false,
		},
		{
			name: "create codespace that requires accepting additional permissions",
			fields: fields{
				apiClient: apiCreateDefaults(&apiClientMock{
					CreateCodespaceFunc: func(ctx context.Context, params *api.CreateCodespaceParams) (*api.Codespace, error) {
						if params.Branch != "main" {
							return nil, fmt.Errorf(literal_7213, params.Branch, "main")
						}
						if params.IdleTimeoutMinutes != 30 {
							return nil, fmt.Errorf(literal_7594, params.IdleTimeoutMinutes)
						}
						return &api.Codespace{}, api.AcceptPermissionsRequiredError{
							AllowPermissionsURL: "https://example.com/permissions",
						}
					},
				}),
			},
			opts: createOptions{
				repo:        literal_3498,
				branch:      "",
				machine:     "GIGA",
				showStatus:  false,
				idleTimeout: 30 * time.Minute,
			},
			wantErr: cmdutil.SilentError,
			wantStderr: `  ✓ Codespaces usage for this repository is paid for by monalisa
You must authorize or deny additional permissions requested by this codespace before continuing.
Open this URL in your browser to review and authorize additional permissions: https://example.com/permissions
Alternatively, you can run "create" with the "--default-permissions" option to continue without authorizing additional permissions.
`,
		},
		{
			name: "create codespace that requires accepting additional permissions for devcontainer path",
			fields: fields{
				apiClient: apiCreateDefaults(&apiClientMock{
					CreateCodespaceFunc: func(ctx context.Context, params *api.CreateCodespaceParams) (*api.Codespace, error) {
						if params.Branch != "feature-branch" {
							return nil, fmt.Errorf(literal_7213, params.Branch, "main")
						}
						if params.IdleTimeoutMinutes != 30 {
							return nil, fmt.Errorf(literal_7594, params.IdleTimeoutMinutes)
						}
						return &api.Codespace{}, api.AcceptPermissionsRequiredError{
							AllowPermissionsURL: "https://example.com/permissions?ref=feature-branch&devcontainer_path=.devcontainer/actions/devcontainer.json",
						}
					},
				}),
			},
			opts: createOptions{
				repo:             literal_3498,
				branch:           "feature-branch",
				devContainerPath: ".devcontainer/actions/devcontainer.json",
				machine:          "GIGA",
				showStatus:       false,
				idleTimeout:      30 * time.Minute,
			},
			wantErr: cmdutil.SilentError,
			wantStderr: `  ✓ Codespaces usage for this repository is paid for by monalisa
You must authorize or deny additional permissions requested by this codespace before continuing.
Open this URL in your browser to review and authorize additional permissions: https://example.com/permissions?ref=feature-branch&devcontainer_path=.devcontainer/actions/devcontainer.json
Alternatively, you can run "create" with the "--default-permissions" option to continue without authorizing additional permissions.
`,
		},
		{
			name: "returns error when user can't create codepaces for a repository",
			fields: fields{
				apiClient: apiCreateDefaults(&apiClientMock{
					GetCodespaceBillableOwnerFunc: func(ctx context.Context, nwo string) (*api.User, error) {
						return nil, fmt.Errorf("some error")
					},
				}),
			},
			opts: createOptions{
				repo:        literal_8159,
				branch:      "",
				machine:     "GIGA",
				showStatus:  false,
				idleTimeout: 30 * time.Minute,
			},
			wantErr: fmt.Errorf("error checking codespace ownership: some error"),
		},
		{
			name: "mentions User as billable owner when org does not cover codepaces for a repository",
			fields: fields{
				apiClient: apiCreateDefaults(&apiClientMock{
					GetCodespaceBillableOwnerFunc: func(ctx context.Context, nwo string) (*api.User, error) {
						return &api.User{
							Type:  "User",
							Login: "monalisa",
						}, nil
					},
					CreateCodespaceFunc: func(ctx context.Context, params *api.CreateCodespaceParams) (*api.Codespace, error) {
						return &api.Codespace{
							Name: literal_8194,
						}, nil
					},
				}),
			},
			opts: createOptions{
				repo:   literal_3498,
				branch: "main",
			},
			wantStderr: literal_6493,
			wantStdout: literal_9271,
		},
		{
			name: "mentions Organization as billable owner when org covers codepaces for a repository",
			fields: fields{
				apiClient: apiCreateDefaults(&apiClientMock{
					GetCodespaceBillableOwnerFunc: func(ctx context.Context, nwo string) (*api.User, error) {
						return &api.User{
							Type:  "Organization",
							Login: "megacorp",
						}, nil
					},
					CreateCodespaceFunc: func(ctx context.Context, params *api.CreateCodespaceParams) (*api.Codespace, error) {
						return &api.Codespace{
							Name: "megacorp-private-abcd1234",
						}, nil
					},
				}),
			},
			opts: createOptions{
				repo:        literal_8159,
				branch:      "",
				machine:     "GIGA",
				showStatus:  false,
				idleTimeout: 30 * time.Minute,
			},
			wantStderr: "  ✓ Codespaces usage for this repository is paid for by megacorp\n",
			wantStdout: "megacorp-private-abcd1234\n",
		},
		{
			name: "does not mention billable owner when not an expected type",
			fields: fields{
				apiClient: apiCreateDefaults(&apiClientMock{
					GetCodespaceBillableOwnerFunc: func(ctx context.Context, nwo string) (*api.User, error) {
						return &api.User{
							Type:  "UnexpectedBillableOwnerType",
							Login: "mega-owner",
						}, nil
					},
					CreateCodespaceFunc: func(ctx context.Context, params *api.CreateCodespaceParams) (*api.Codespace, error) {
						return &api.Codespace{
							Name: "megacorp-private-abcd1234",
						}, nil
					},
				}),
			},
			opts: createOptions{
				repo: literal_8159,
			},
			wantStdout: "megacorp-private-abcd1234\n",
		},
		{
			name: "return default url when using web flag without other flags",
			fields: fields{
				apiClient: apiCreateDefaults(&apiClientMock{
					ServerURLFunc: func() string {
						return literal_3972
					},
				}),
			},
			opts: createOptions{
				useWeb: true,
			},
			wantURL: "https://github.com/codespaces/new",
		},
		{
			name: "return custom server url when using web flag",
			fields: fields{
				apiClient: apiCreateDefaults(&apiClientMock{
					ServerURLFunc: func() string {
						return "https://github.mycompany.com"
					},
				}),
			},
			opts: createOptions{
				useWeb: true,
			},
			wantURL: "https://github.mycompany.com/codespaces/new",
		},
		{
			name: "skip machine check when using web flag and no machine provided",
			fields: fields{
				apiClient: apiCreateDefaults(&apiClientMock{
					GetRepositoryFunc: func(ctx context.Context, nwo string) (*api.Repository, error) {
						return &api.Repository{
							ID:            123,
							DefaultBranch: "main",
						}, nil
					},
					CreateCodespaceFunc: func(ctx context.Context, params *api.CreateCodespaceParams) (*api.Codespace, error) {
						return &api.Codespace{
							Name: literal_8194,
						}, nil
					},
					ServerURLFunc: func() string {
						return literal_3972
					},
				}),
			},
			opts: createOptions{
				repo:     literal_3498,
				useWeb:   true,
				branch:   "custom",
				location: "EastUS",
			},
			wantStderr: literal_6493,
			wantURL:    fmt.Sprintf(literal_7021, 123, "custom", "", "EastUS"),
		},
		{
			name: "return correct url with correct params when using web flag and repo flag",
			fields: fields{
				apiClient: apiCreateDefaults(&apiClientMock{
					GetRepositoryFunc: func(ctx context.Context, nwo string) (*api.Repository, error) {
						return &api.Repository{
							ID:            123,
							DefaultBranch: "main",
						}, nil
					},
					CreateCodespaceFunc: func(ctx context.Context, params *api.CreateCodespaceParams) (*api.Codespace, error) {
						return &api.Codespace{
							Name: literal_8194,
						}, nil
					},
					ServerURLFunc: func() string {
						return literal_3972
					},
				}),
			},
			opts: createOptions{
				repo:   literal_3498,
				useWeb: true,
			},
			wantStderr: literal_6493,
			wantURL:    fmt.Sprintf(literal_7021, 123, "main", "", ""),
		},
		{
			name: "return correct url with correct params when using web flag, repo, branch, location, machine flag",
			fields: fields{
				apiClient: apiCreateDefaults(&apiClientMock{
					GetRepositoryFunc: func(ctx context.Context, nwo string) (*api.Repository, error) {
						return &api.Repository{
							ID:            123,
							DefaultBranch: "main",
						}, nil
					},
					CreateCodespaceFunc: func(ctx context.Context, params *api.CreateCodespaceParams) (*api.Codespace, error) {
						return &api.Codespace{
							Name:    literal_8194,
							Machine: api.CodespaceMachine{Name: "GIGA"},
						}, nil
					},
					ServerURLFunc: func() string {
						return literal_3972
					},
				}),
			},
			opts: createOptions{
				repo:     literal_3498,
				machine:  "GIGA",
				branch:   "custom",
				location: "EastUS",
				useWeb:   true,
			},
			wantStderr: literal_6493,
			wantURL:    fmt.Sprintf(literal_7021, 123, "custom", "GIGA", "EastUS"),
		},
	}
	var a *App
	var b *browser.Stub

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ios, _, stdout, stderr := iostreams.Test()
			ios.SetStdoutTTY(tt.isTTY)
			ios.SetStdinTTY(tt.isTTY)
			ios.SetStderrTTY(tt.isTTY)

			if tt.opts.useWeb {
				b = &browser.Stub{}
				a = &App{
					io:        ios,
					apiClient: tt.fields.apiClient,
					browser:   b,
				}
			} else {
				a = &App{
					io:        ios,
					apiClient: tt.fields.apiClient,
				}
			}

			err := a.Create(context.Background(), tt.opts)
			if err != nil && tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
			}
			if err != nil && tt.wantErr == nil {
				t.Logf(err.Error())
			}
			if got := stdout.String(); got != tt.wantStdout {
				t.Logf(t.Name())
				t.Errorf("  stdout = %v, want %v", got, tt.wantStdout)
			}
			if got := stderr.String(); got != tt.wantStderr {
				t.Logf(t.Name())
				t.Errorf("  stderr = %v, want %v", got, tt.wantStderr)
			}

			if tt.opts.useWeb {
				b.Verify(t, tt.wantURL)
			}
		})
	}
}

func TestBuildDisplayName(t *testing.T) {
	tests := []struct {
		name                 string
		prebuildAvailability string
		expectedDisplayName  string
	}{
		{
			name:                 "prebuild availability is none",
			prebuildAvailability: "none",
			expectedDisplayName:  literal_0361,
		},
		{
			name:                 "prebuild availability is empty",
			prebuildAvailability: "",
			expectedDisplayName:  literal_0361,
		},
		{
			name:                 "prebuild availability is ready",
			prebuildAvailability: "ready",
			expectedDisplayName:  "4 cores, 8 GB RAM, 32 GB storage (Prebuild ready)",
		},
		{
			name:                 "prebuild availability is in_progress",
			prebuildAvailability: "in_progress",
			expectedDisplayName:  "4 cores, 8 GB RAM, 32 GB storage (Prebuild in progress)",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			displayName := buildDisplayName(literal_0361, tt.prebuildAvailability)

			if displayName != tt.expectedDisplayName {
				t.Errorf("displayName = %q, expectedDisplayName %q", displayName, tt.expectedDisplayName)
			}
		})
	}
}

type MockSurveyPrompter struct {
	AskFunc func(qs []*survey.Question, response interface{}) error
}

func (m *MockSurveyPrompter) Ask(qs []*survey.Question, response interface{}) error {
	return m.AskFunc(qs, response)
}

type MockBrowser struct {
	Err error
}

func (b *MockBrowser) Browse(url string) error {
	if b.Err != nil {
		return b.Err
	}

	return nil
}

func TestHandleAdditionalPermissions(t *testing.T) {
	tests := []struct {
		name                  string
		isInteractive         bool
		accept                string
		permissionsOptOut     bool
		browserErr            error
		pollForPermissionsErr error
		createCodespaceErr    error
		wantErr               bool
	}{
		{
			name:              "non-interactive",
			isInteractive:     false,
			permissionsOptOut: false,
			wantErr:           true,
		},
		{
			name:              "interactive, continue in browser, browser error",
			isInteractive:     true,
			accept:            literal_9407,
			permissionsOptOut: false,
			browserErr:        fmt.Errorf("browser error"),
			wantErr:           true,
		},
		{
			name:                  "interactive, continue in browser, poll for permissions error",
			isInteractive:         true,
			accept:                literal_9407,
			permissionsOptOut:     false,
			pollForPermissionsErr: fmt.Errorf("poll for permissions error"),
			wantErr:               true,
		},
		{
			name:               "interactive, continue in browser, create codespace error",
			isInteractive:      true,
			accept:             literal_9407,
			permissionsOptOut:  false,
			createCodespaceErr: fmt.Errorf("create codespace error"),
			wantErr:            true,
		},
		{
			name:               "interactive, continue without authorizing",
			isInteractive:      true,
			accept:             "Continue without authorizing additional permissions",
			permissionsOptOut:  true,
			createCodespaceErr: fmt.Errorf("create codespace error"),
			wantErr:            true,
		},
		{
			name:              "interactive, continue without authorizing, create codespace success",
			isInteractive:     true,
			accept:            "Continue without authorizing additional permissions",
			permissionsOptOut: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ios, _, _, _ := iostreams.Test()
			a := &App{
				io: ios,
				browser: &MockBrowser{
					Err: tt.browserErr,
				},
				apiClient: &apiClientMock{
					CreateCodespaceFunc: func(ctx context.Context, params *api.CreateCodespaceParams) (*api.Codespace, error) {
						return nil, tt.createCodespaceErr
					},
					GetCodespacesPermissionsCheckFunc: func(ctx context.Context, repoID int, branch string, devcontainerPath string) (bool, error) {
						if tt.pollForPermissionsErr != nil {
							return false, tt.pollForPermissionsErr
						}
						return true, nil
					},
				},
			}

			if tt.isInteractive {
				a.io.SetStdinTTY(true)
				a.io.SetStdoutTTY(true)
				a.io.SetStderrTTY(true)
			}

			params := &api.CreateCodespaceParams{}
			_, err := a.handleAdditionalPermissions(context.Background(), &MockSurveyPrompter{
				AskFunc: func(qs []*survey.Question, response interface{}) error {
					*response.(*struct{ Accept string }) = struct{ Accept string }{Accept: tt.accept}
					return nil
				},
			}, params, "http://example.com")
			if (err != nil) != tt.wantErr {
				t.Errorf("handleAdditionalPermissions() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.permissionsOptOut != params.PermissionsOptOut {
				t.Errorf("handleAdditionalPermissions() permissionsOptOut = %v, want %v", params.PermissionsOptOut, tt.permissionsOptOut)
			}
		})
	}
}

func apiCreateDefaults(c *apiClientMock) *apiClientMock {
	if c.GetRepositoryFunc == nil {
		c.GetRepositoryFunc = func(ctx context.Context, nwo string) (*api.Repository, error) {
			return &api.Repository{
				ID:            1234,
				FullName:      nwo,
				DefaultBranch: "main",
			}, nil
		}
	}
	if c.GetCodespaceBillableOwnerFunc == nil {
		c.GetCodespaceBillableOwnerFunc = func(ctx context.Context, nwo string) (*api.User, error) {
			return &api.User{
				Login: "monalisa",
				Type:  "User",
			}, nil
		}
	}
	if c.ListDevContainersFunc == nil {
		c.ListDevContainersFunc = func(ctx context.Context, repoID int, branch string, limit int) ([]api.DevContainerEntry, error) {
			return []api.DevContainerEntry{{Path: ".devcontainer/devcontainer.json"}}, nil
		}
	}
	if c.GetCodespacesMachinesFunc == nil {
		c.GetCodespacesMachinesFunc = func(ctx context.Context, repoID int, branch, location string, devcontainerPath string) ([]*api.Machine, error) {
			return []*api.Machine{
				{
					Name:        "GIGA",
					DisplayName: literal_9163,
				},
			}, nil
		}
	}
	return c
}

func durationPtr(d time.Duration) *time.Duration {
	return &d
}

const literal_7213 = "got branch %q, want %q"

const literal_7594 = "idle timeout minutes was %v"

const literal_8194 = "monalisa-dotfiles-abcd1234"

const literal_3498 = "monalisa/dotfiles"

const literal_9271 = "monalisa-dotfiles-abcd1234\n"

const literal_6493 = "  ✓ Codespaces usage for this repository is paid for by monalisa\n"

const literal_6279 = "funky flute"

const literal_8907 = ".devcontainer/foobar/devcontainer.json"

const literal_0195 = "got dev container path %q, want %q"

const literal_9163 = "Gigabits of a machine"

const literal_8159 = "megacorp/private"

const literal_3972 = "https://github.com"

const literal_7021 = "https://github.com/codespaces/new?repo=%d&ref=%s&machine=%s&location=%s"

const literal_0361 = "4 cores, 8 GB RAM, 32 GB storage"

const literal_9407 = "Continue in browser to review and authorize additional permissions (Recommended)"
