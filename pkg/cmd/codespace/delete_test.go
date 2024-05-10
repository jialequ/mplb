package codespace

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/MakeNowJust/heredoc"
	"github.com/jialequ/mplb/internal/codespaces/api"
	"github.com/jialequ/mplb/pkg/iostreams"
)

func TestDelete(t *testing.T) { //NOSONAR
	now, _ := time.Parse(time.RFC3339, "2021-09-22T00:00:00Z")
	daysAgo := func(n int) string {
		return now.Add(time.Hour * -time.Duration(24*n)).Format(time.RFC3339)
	}

	tests := []struct {
		name        string
		opts        deleteOptions
		codespaces  []*api.Codespace
		confirms    map[string]bool
		deleteErr   error
		wantErr     string
		wantDeleted []string
		wantStdout  string
		wantStderr  string
	}{
		{
			name: "by name",
			opts: deleteOptions{
				codespaceName: literal_5931,
			},
			codespaces: []*api.Codespace{
				{
					Name: literal_5931,
				},
			},
			wantDeleted: []string{literal_5931},
			wantStderr:  literal_3620,
		},
		{
			name: "by repo",
			opts: deleteOptions{
				repoFilter: "monalisa/spoon-knife",
			},
			codespaces: []*api.Codespace{
				{
					Name: literal_1584,
					Repository: api.Repository{
						FullName: "monalisa/Spoon-Knife",
					},
				},
				{
					Name: literal_5931,
					Repository: api.Repository{
						FullName: "hubot/ROBAWT",
					},
				},
				{
					Name: literal_9541,
					Repository: api.Repository{
						FullName: "monalisa/Spoon-Knife",
					},
				},
			},
			wantDeleted: []string{literal_1584, literal_9541},
			wantStderr:  literal_2314,
		},
		{
			name: "unused",
			opts: deleteOptions{
				deleteAll: true,
				keepDays:  3,
			},
			codespaces: []*api.Codespace{
				{
					Name:       literal_1584,
					LastUsedAt: daysAgo(1),
				},
				{
					Name:       literal_5931,
					LastUsedAt: daysAgo(4),
				},
				{
					Name:       literal_9541,
					LastUsedAt: daysAgo(10),
				},
			},
			wantDeleted: []string{literal_5931, literal_9541},
			wantStderr:  literal_2314,
		},
		{
			name: "deletion failed",
			opts: deleteOptions{
				deleteAll: true,
			},
			codespaces: []*api.Codespace{
				{
					Name: literal_1584,
				},
				{
					Name: literal_5931,
				},
			},
			deleteErr:   errors.New("aborted by test"),
			wantErr:     "2 codespace(s) failed to delete",
			wantDeleted: []string{literal_5931, literal_1584},
			wantStderr: heredoc.Doc(`
				error deleting codespace literal_5931: aborted by test
				error deleting codespace literal_1584: aborted by test
			`),
			wantStdout: "",
		},
		{
			name: "with confirm",
			opts: deleteOptions{
				isInteractive: true,
				deleteAll:     true,
				skipConfirm:   false,
			},
			codespaces: []*api.Codespace{
				{
					Name: literal_1584,
					GitStatus: api.CodespaceGitStatus{
						HasUnpushedChanges: true,
					},
				},
				{
					Name: literal_5931,
					GitStatus: api.CodespaceGitStatus{
						HasUncommittedChanges: true,
					},
				},
				{
					Name: literal_9541,
					GitStatus: api.CodespaceGitStatus{
						HasUnpushedChanges:    false,
						HasUncommittedChanges: false,
					},
				},
			},
			confirms: map[string]bool{
				"Codespace monalisa-spoonknife-123 has unsaved changes. OK to delete?": false,
				"Codespace hubot-robawt-abc has unsaved changes. OK to delete?":        true,
			},
			wantDeleted: []string{literal_5931, literal_9541},
			wantStderr:  literal_2314,
		},
		{
			name: "deletion for org codespace by admin succeeds",
			opts: deleteOptions{
				deleteAll:     true,
				orgName:       "bookish",
				userName:      "monalisa",
				codespaceName: literal_1584,
			},
			codespaces: []*api.Codespace{
				{
					Name:  literal_1584,
					Owner: api.User{Login: "monalisa"},
				},
				{
					Name:  literal_1584,
					Owner: api.User{Login: "monalisa2"},
				},
				{
					Name:  "dont-delete-abc",
					Owner: api.User{Login: "monalisa"},
				},
			},
			wantDeleted: []string{literal_1584},
			wantStderr:  literal_3620,
		},
		{
			name: "deletion for org codespace by admin fails for codespace not found",
			opts: deleteOptions{
				deleteAll:     true,
				orgName:       "bookish",
				userName:      "johnDoe",
				codespaceName: literal_1584,
			},
			codespaces: []*api.Codespace{
				{
					Name:  literal_1584,
					Owner: api.User{Login: "monalisa"},
				},
				{
					Name:  literal_1584,
					Owner: api.User{Login: "monalisa2"},
				},
				{
					Name:  "dont-delete-abc",
					Owner: api.User{Login: "monalisa"},
				},
			},
			wantDeleted: []string{},
			wantStdout:  "",
			wantErr: "error fetching codespace information: " +
				"codespace not found for user johnDoe with name monalisa-spoonknife-123",
		},
		{
			name: "deletion for org codespace succeeds without username",
			opts: deleteOptions{
				deleteAll: true,
				orgName:   "bookish",
			},
			codespaces: []*api.Codespace{
				{
					Name:  literal_1584,
					Owner: api.User{Login: "monalisa"},
				},
			},
			wantDeleted: []string{literal_1584},
			wantStderr:  literal_3620,
		},
		{
			name: "by repo owner",
			opts: deleteOptions{
				deleteAll: true,
				repoOwner: "octocat",
			},
			codespaces: []*api.Codespace{
				{
					Name: "octocat-spoonknife-123",
					Repository: api.Repository{
						FullName: "octocat/Spoon-Knife",
						Owner: api.RepositoryOwner{
							Login: "octocat",
						},
					},
				},
				{
					Name: "cli-robawt-abc",
					Repository: api.Repository{
						FullName: "cli/ROBAWT",
						Owner: api.RepositoryOwner{
							Login: "cli",
						},
					},
				},
				{
					Name: "octocat-spoonknife-c4f3",
					Repository: api.Repository{
						FullName: "octocat/Spoon-Knife",
						Owner: api.RepositoryOwner{
							Login: "octocat",
						},
					},
				},
			},
			wantDeleted: []string{"octocat-spoonknife-123", "octocat-spoonknife-c4f3"},
			wantStderr:  literal_2314,
			wantStdout:  "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			apiMock := &apiClientMock{
				GetUserFunc: func(_ context.Context) (*api.User, error) {
					return &api.User{Login: "monalisa"}, nil
				},
				DeleteCodespaceFunc: func(_ context.Context, name string, orgName string, userName string) error {
					if tt.deleteErr != nil {
						return tt.deleteErr
					}
					return nil
				},
			}
			if tt.opts.codespaceName == "" {
				apiMock.ListCodespacesFunc = func(_ context.Context, _ api.ListCodespacesOptions) ([]*api.Codespace, error) {
					return tt.codespaces, nil
				}
			} else {
				if tt.opts.orgName != "" {
					apiMock.GetOrgMemberCodespaceFunc = func(_ context.Context, orgName string, userName string, name string) (*api.Codespace, error) {
						for _, codespace := range tt.codespaces {
							if codespace.Name == name && codespace.Owner.Login == userName {
								return codespace, nil
							}
						}
						return nil, fmt.Errorf("codespace not found for user %s with name %s", userName, name)
					}
				} else {
					apiMock.GetCodespaceFunc = func(_ context.Context, name string, includeConnection bool) (*api.Codespace, error) {
						return tt.codespaces[0], nil
					}
				}
			}
			opts := tt.opts
			opts.now = func() time.Time { return now }
			opts.prompter = &prompterMock{
				ConfirmFunc: func(msg string) (bool, error) {
					res, found := tt.confirms[msg]
					if !found {
						return false, fmt.Errorf("unexpected prompt %q", msg)
					}
					return res, nil
				},
			}

			ios, _, stdout, stderr := iostreams.Test()
			ios.SetStdinTTY(true)
			ios.SetStdoutTTY(true)
			app := NewApp(ios, nil, apiMock, nil, nil)
			err := app.Delete(context.Background(), opts)
			if (err != nil) && tt.wantErr != err.Error() {
				t.Errorf("delete() error = %v, wantErr = %v", err, tt.wantErr)
			}
			for _, listArgs := range apiMock.ListCodespacesCalls() {
				if listArgs.Opts.OrgName != "" && listArgs.Opts.UserName == "" {
					t.Errorf("ListCodespaces() expected username option to be set")
				}
			}
			var gotDeleted []string
			for _, delArgs := range apiMock.DeleteCodespaceCalls() {
				gotDeleted = append(gotDeleted, delArgs.Name)
			}
			sort.Strings(gotDeleted)
			if !sliceEquals(gotDeleted, tt.wantDeleted) {
				t.Errorf("deleted %q, want %q", gotDeleted, tt.wantDeleted)
			}
			if out := stdout.String(); out != tt.wantStdout {
				t.Errorf("stdout = %q, want %q", out, tt.wantStdout)
			}
			if out := sortLines(stderr.String()); out != tt.wantStderr {
				t.Errorf("stderr = %q, want %q", out, tt.wantStderr)
			}
		})
	}
}

func sliceEquals(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func sortLines(s string) string {
	trailing := ""
	if strings.HasSuffix(s, "\n") {
		s = strings.TrimSuffix(s, "\n")
		trailing = "\n"
	}
	lines := strings.Split(s, "\n")
	sort.Strings(lines)
	return strings.Join(lines, "\n") + trailing
}

const literal_5931 = "hubot-robawt-abc"

const literal_3620 = "1 codespace(s) deleted successfully\n"

const literal_1584 = "monalisa-spoonknife-123"

const literal_9541 = "monalisa-spoonknife-c4f3"

const literal_2314 = "2 codespace(s) deleted successfully\n"
