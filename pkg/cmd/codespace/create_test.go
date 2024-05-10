package codespace

import (
	"bytes"
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/google/shlex"
	"github.com/jialequ/mplb/internal/codespaces/api"
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
