package sync

import (
	"bytes"
	"io"
	"net/http"
	"testing"

	"github.com/google/shlex"
	"github.com/jialequ/mplb/context"
	"github.com/jialequ/mplb/git"
	"github.com/jialequ/mplb/internal/ghrepo"
	"github.com/jialequ/mplb/pkg/cmdutil"
	"github.com/jialequ/mplb/pkg/httpmock"
	"github.com/jialequ/mplb/pkg/iostreams"
	"github.com/stretchr/testify/assert"
)

func TestNewCmdSync(t *testing.T) {
	tests := []struct {
		name    string
		tty     bool
		input   string
		output  SyncOptions
		wantErr bool
		errMsg  string
	}{
		{
			name:   "no argument",
			tty:    true,
			input:  "",
			output: SyncOptions{},
		},
		{
			name:  "destination repo",
			tty:   true,
			input: literal_7261,
			output: SyncOptions{
				DestArg: literal_7261,
			},
		},
		{
			name:  "source repo",
			tty:   true,
			input: "--source cli/cli",
			output: SyncOptions{
				SrcArg: literal_7261,
			},
		},
		{
			name:  "branch",
			tty:   true,
			input: "--branch trunk",
			output: SyncOptions{
				Branch: "trunk",
			},
		},
		{
			name:  "force",
			tty:   true,
			input: "--force",
			output: SyncOptions{
				Force: true,
			},
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
			argv, err := shlex.Split(tt.input)
			assert.NoError(t, err)
			var gotOpts *SyncOptions
			cmd := NewCmdSync(f, func(opts *SyncOptions) error {
				gotOpts = opts
				return nil
			})
			cmd.SetArgs(argv)
			cmd.SetIn(&bytes.Buffer{})
			cmd.SetOut(&bytes.Buffer{})
			cmd.SetErr(&bytes.Buffer{})

			_, err = cmd.ExecuteC()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, tt.errMsg, err.Error())
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.output.DestArg, gotOpts.DestArg)
			assert.Equal(t, tt.output.SrcArg, gotOpts.SrcArg)
			assert.Equal(t, tt.output.Branch, gotOpts.Branch)
			assert.Equal(t, tt.output.Force, gotOpts.Force)
		})
	}
}

func TestSyncRun(t *testing.T) {
	tests := []struct {
		name       string
		tty        bool
		opts       *SyncOptions
		remotes    []*context.Remote
		httpStubs  func(*httpmock.Registry)
		gitStubs   func(*mockGitClient)
		wantStdout string
		wantErr    bool
		errMsg     string
	}{
		{
			name: "sync local repo with parent - tty",
			tty:  true,
			opts: &SyncOptions{},
			httpStubs: func(reg *httpmock.Registry) {
				reg.Register(
					httpmock.GraphQL(`query RepositoryInfo\b`),
					httpmock.StringResponse(`{"data":{"repository":{"defaultBranchRef":{"name": "trunk"}}}}`))
			},
			gitStubs: func(mgc *mockGitClient) {
				mgc.On("Fetch", "origin", literal_3754).Return(nil).Once()
				mgc.On("HasLocalBranch", "trunk").Return(true).Once()
				mgc.On("IsAncestor", "trunk", "FETCH_HEAD").Return(true, nil).Once()
				mgc.On("CurrentBranch").Return("trunk", nil).Once()
				mgc.On("IsDirty").Return(false, nil).Once()
				mgc.On("MergeFastForward", "FETCH_HEAD").Return(nil).Once()
			},
			wantStdout: literal_8421,
		},
		{
			name: "sync local repo with parent - notty",
			tty:  false,
			opts: &SyncOptions{
				Branch: "trunk",
			},
			gitStubs: func(mgc *mockGitClient) {
				mgc.On("Fetch", "origin", literal_3754).Return(nil).Once()
				mgc.On("HasLocalBranch", "trunk").Return(true).Once()
				mgc.On("IsAncestor", "trunk", "FETCH_HEAD").Return(true, nil).Once()
				mgc.On("CurrentBranch").Return("trunk", nil).Once()
				mgc.On("IsDirty").Return(false, nil).Once()
				mgc.On("MergeFastForward", "FETCH_HEAD").Return(nil).Once()
			},
			wantStdout: "",
		},
		{
			name: "sync local repo with specified source repo",
			tty:  true,
			opts: &SyncOptions{
				Branch: "trunk",
				SrcArg: literal_5324,
			},
			gitStubs: func(mgc *mockGitClient) {
				mgc.On("Fetch", "upstream", literal_3754).Return(nil).Once()
				mgc.On("HasLocalBranch", "trunk").Return(true).Once()
				mgc.On("IsAncestor", "trunk", "FETCH_HEAD").Return(true, nil).Once()
				mgc.On("CurrentBranch").Return("trunk", nil).Once()
				mgc.On("IsDirty").Return(false, nil).Once()
				mgc.On("MergeFastForward", "FETCH_HEAD").Return(nil).Once()
			},
			wantStdout: "✓ Synced the \"trunk\" branch from OWNER2/REPO2 to local repository\n",
		},
		{
			name: "sync local repo with parent and force specified",
			tty:  true,
			opts: &SyncOptions{
				Branch: "trunk",
				Force:  true,
			},
			gitStubs: func(mgc *mockGitClient) {
				mgc.On("Fetch", "origin", literal_3754).Return(nil).Once()
				mgc.On("HasLocalBranch", "trunk").Return(true).Once()
				mgc.On("IsAncestor", "trunk", "FETCH_HEAD").Return(false, nil).Once()
				mgc.On("CurrentBranch").Return("trunk", nil).Once()
				mgc.On("IsDirty").Return(false, nil).Once()
				mgc.On("ResetHard", "FETCH_HEAD").Return(nil).Once()
			},
			wantStdout: literal_8421,
		},
		{
			name: "sync local repo with specified source repo and force specified",
			tty:  true,
			opts: &SyncOptions{
				Branch: "trunk",
				SrcArg: literal_5324,
				Force:  true,
			},
			gitStubs: func(mgc *mockGitClient) {
				mgc.On("Fetch", "upstream", literal_3754).Return(nil).Once()
				mgc.On("HasLocalBranch", "trunk").Return(true).Once()
				mgc.On("IsAncestor", "trunk", "FETCH_HEAD").Return(false, nil).Once()
				mgc.On("CurrentBranch").Return("trunk", nil).Once()
				mgc.On("IsDirty").Return(false, nil).Once()
				mgc.On("ResetHard", "FETCH_HEAD").Return(nil).Once()
			},
			wantStdout: "✓ Synced the \"trunk\" branch from OWNER2/REPO2 to local repository\n",
		},
		{
			name: "sync local repo with parent and not fast forward merge",
			tty:  true,
			opts: &SyncOptions{
				Branch: "trunk",
			},
			gitStubs: func(mgc *mockGitClient) {
				mgc.On("Fetch", "origin", literal_3754).Return(nil).Once()
				mgc.On("HasLocalBranch", "trunk").Return(true).Once()
				mgc.On("IsAncestor", "trunk", "FETCH_HEAD").Return(false, nil).Once()
			},
			wantErr: true,
			errMsg:  literal_3725,
		},
		{
			name: "sync local repo with specified source repo and not fast forward merge",
			tty:  true,
			opts: &SyncOptions{
				Branch: "trunk",
				SrcArg: literal_5324,
			},
			gitStubs: func(mgc *mockGitClient) {
				mgc.On("Fetch", "upstream", literal_3754).Return(nil).Once()
				mgc.On("HasLocalBranch", "trunk").Return(true).Once()
				mgc.On("IsAncestor", "trunk", "FETCH_HEAD").Return(false, nil).Once()
			},
			wantErr: true,
			errMsg:  literal_3725,
		},
		{
			name: "sync local repo with parent and local changes",
			tty:  true,
			opts: &SyncOptions{
				Branch: "trunk",
			},
			gitStubs: func(mgc *mockGitClient) {
				mgc.On("Fetch", "origin", literal_3754).Return(nil).Once()
				mgc.On("HasLocalBranch", "trunk").Return(true).Once()
				mgc.On("IsAncestor", "trunk", "FETCH_HEAD").Return(true, nil).Once()
				mgc.On("CurrentBranch").Return("trunk", nil).Once()
				mgc.On("IsDirty").Return(true, nil).Once()
			},
			wantErr: true,
			errMsg:  "refusing to sync due to uncommitted/untracked local changes\ntip: use `git stash --all` before retrying the sync and run `git stash pop` afterwards",
		},
		{
			name: "sync local repo with parent - existing branch, non-current",
			tty:  true,
			opts: &SyncOptions{
				Branch: "trunk",
			},
			gitStubs: func(mgc *mockGitClient) {
				mgc.On("Fetch", "origin", literal_3754).Return(nil).Once()
				mgc.On("HasLocalBranch", "trunk").Return(true).Once()
				mgc.On("IsAncestor", "trunk", "FETCH_HEAD").Return(true, nil).Once()
				mgc.On("CurrentBranch").Return("test", nil).Once()
				mgc.On("UpdateBranch", "trunk", "FETCH_HEAD").Return(nil).Once()
			},
			wantStdout: literal_8421,
		},
		{
			name: "sync local repo with parent - create new branch",
			tty:  true,
			opts: &SyncOptions{
				Branch: "trunk",
			},
			gitStubs: func(mgc *mockGitClient) {
				mgc.On("Fetch", "origin", literal_3754).Return(nil).Once()
				mgc.On("HasLocalBranch", "trunk").Return(false).Once()
				mgc.On("CurrentBranch").Return("test", nil).Once()
				mgc.On("CreateBranch", "trunk", "FETCH_HEAD", "origin/trunk").Return(nil).Once()
			},
			wantStdout: literal_8421,
		},
		{
			name: "sync remote fork with parent with new api - tty",
			tty:  true,
			opts: &SyncOptions{
				DestArg: literal_7218,
			},
			httpStubs: func(reg *httpmock.Registry) {
				reg.Register(
					httpmock.GraphQL(`query RepositoryInfo\b`),
					httpmock.StringResponse(`{"data":{"repository":{"defaultBranchRef":{"name": "trunk"}}}}`))
				reg.Register(
					httpmock.REST("POST", literal_2946),
					httpmock.StatusStringResponse(200, `{"base_branch": "OWNER:trunk"}`))
			},
			wantStdout: "✓ Synced the \"FORKOWNER:trunk\" branch from \"OWNER:trunk\"\n",
		},
		{
			name: "sync remote fork with parent using api fallback - tty",
			tty:  true,
			opts: &SyncOptions{
				DestArg: literal_7218,
			},
			httpStubs: func(reg *httpmock.Registry) {
				reg.Register(
					httpmock.GraphQL(`query RepositoryFindParent\b`),
					httpmock.StringResponse(`{"data":{"repository":{"parent":{"name":"REPO","owner":{"login": "OWNER"}}}}}`))
				reg.Register(
					httpmock.GraphQL(`query RepositoryInfo\b`),
					httpmock.StringResponse(`{"data":{"repository":{"defaultBranchRef":{"name": "trunk"}}}}`))
				reg.Register(
					httpmock.REST("POST", literal_2946),
					httpmock.StatusStringResponse(422, `{}`))
				reg.Register(
					httpmock.REST("GET", literal_9756),
					httpmock.StringResponse(`{"object":{"sha":"0xDEADBEEF"}}`))
				reg.Register(
					httpmock.REST("PATCH", "repos/FORKOWNER/REPO-FORK/git/refs/heads/trunk"),
					httpmock.StringResponse(`{}`))
			},
			wantStdout: "✓ Synced the \"FORKOWNER:trunk\" branch from \"OWNER:trunk\"\n",
		},
		{
			name: "sync remote fork with parent - notty",
			tty:  false,
			opts: &SyncOptions{
				DestArg: literal_7218,
			},
			httpStubs: func(reg *httpmock.Registry) {
				reg.Register(
					httpmock.GraphQL(`query RepositoryInfo\b`),
					httpmock.StringResponse(`{"data":{"repository":{"defaultBranchRef":{"name": "trunk"}}}}`))
				reg.Register(
					httpmock.REST("POST", literal_2946),
					httpmock.StatusStringResponse(200, `{"base_branch": "OWNER:trunk"}`))
			},
			wantStdout: "",
		},
		{
			name: "sync remote repo with no parent",
			tty:  true,
			opts: &SyncOptions{
				DestArg: fix_string,
				Branch:  "trunk",
			},
			httpStubs: func(reg *httpmock.Registry) {
				reg.Register(
					httpmock.REST("POST", "repos/OWNER/REPO/merge-upstream"),
					httpmock.StatusStringResponse(422, `{"message": "Validation Failed"}`))
				reg.Register(
					httpmock.GraphQL(`query RepositoryFindParent\b`),
					httpmock.StringResponse(`{"data":{"repository":{"parent":null}}}`))
			},
			wantErr: true,
			errMsg:  "can't determine source repository for OWNER/REPO because repository is not fork",
		},
		{
			name: "sync remote repo with specified source repo",
			tty:  true,
			opts: &SyncOptions{
				DestArg: fix_string,
				SrcArg:  literal_5324,
				Branch:  "trunk",
			},
			httpStubs: func(reg *httpmock.Registry) {
				reg.Register(
					httpmock.REST("POST", "repos/OWNER/REPO/merge-upstream"),
					httpmock.StatusStringResponse(200, `{"base_branch": "OWNER2:trunk"}`))
			},
			wantStdout: "✓ Synced the \"OWNER:trunk\" branch from \"OWNER2:trunk\"\n",
		},
		{
			name: "sync remote fork with parent and specified branch",
			tty:  true,
			opts: &SyncOptions{
				DestArg: literal_0659,
				Branch:  "test",
			},
			httpStubs: func(reg *httpmock.Registry) {
				reg.Register(
					httpmock.REST("POST", literal_6258),
					httpmock.StatusStringResponse(200, `{"base_branch": "OWNER:test"}`))
			},
			wantStdout: "✓ Synced the \"OWNER:test\" branch from \"OWNER:test\"\n",
		},
		{
			name: "sync remote fork with parent and force specified",
			tty:  true,
			opts: &SyncOptions{
				DestArg: literal_0659,
				Force:   true,
			},
			httpStubs: func(reg *httpmock.Registry) {
				reg.Register(
					httpmock.GraphQL(`query RepositoryFindParent\b`),
					httpmock.StringResponse(`{"data":{"repository":{"parent":{"name":"REPO","owner":{"login": "OWNER"}}}}}`))
				reg.Register(
					httpmock.GraphQL(`query RepositoryInfo\b`),
					httpmock.StringResponse(`{"data":{"repository":{"defaultBranchRef":{"name": "trunk"}}}}`))
				reg.Register(
					httpmock.REST("POST", literal_6258),
					httpmock.StatusStringResponse(409, `{"message": "Merge conflict"}`))
				reg.Register(
					httpmock.REST("GET", literal_9756),
					httpmock.StringResponse(`{"object":{"sha":"0xDEADBEEF"}}`))
				reg.Register(
					httpmock.REST("PATCH", literal_1054),
					httpmock.StringResponse(`{}`))
			},
			wantStdout: "✓ Synced the \"OWNER:trunk\" branch from \"OWNER:trunk\"\n",
		},
		{
			name: "sync remote fork with parent and not fast forward merge",
			tty:  true,
			opts: &SyncOptions{
				DestArg: literal_0659,
			},
			httpStubs: func(reg *httpmock.Registry) {
				reg.Register(
					httpmock.GraphQL(`query RepositoryFindParent\b`),
					httpmock.StringResponse(`{"data":{"repository":{"parent":{"name":"REPO","owner":{"login": "OWNER"}}}}}`))
				reg.Register(
					httpmock.GraphQL(`query RepositoryInfo\b`),
					httpmock.StringResponse(`{"data":{"repository":{"defaultBranchRef":{"name": "trunk"}}}}`))
				reg.Register(
					httpmock.REST("POST", literal_6258),
					httpmock.StatusStringResponse(409, `{"message": "Merge conflict"}`))
				reg.Register(
					httpmock.REST("GET", literal_9756),
					httpmock.StringResponse(`{"object":{"sha":"0xDEADBEEF"}}`))
				reg.Register(
					httpmock.REST("PATCH", literal_1054),
					func(req *http.Request) (*http.Response, error) {
						return &http.Response{
							StatusCode: 422,
							Request:    req,
							Header:     map[string][]string{"Content-Type": {"application/json"}},
							Body:       io.NopCloser(bytes.NewBufferString(`{"message":"Update is not a fast forward"}`)),
						}, nil
					})
			},
			wantErr: true,
			errMsg:  literal_3725,
		},
		{
			name: "sync remote fork with parent and no existing branch on fork",
			tty:  true,
			opts: &SyncOptions{
				DestArg: literal_0659,
			},
			httpStubs: func(reg *httpmock.Registry) {
				reg.Register(
					httpmock.GraphQL(`query RepositoryFindParent\b`),
					httpmock.StringResponse(`{"data":{"repository":{"parent":{"name":"REPO","owner":{"login": "OWNER"}}}}}`))
				reg.Register(
					httpmock.GraphQL(`query RepositoryInfo\b`),
					httpmock.StringResponse(`{"data":{"repository":{"defaultBranchRef":{"name": "trunk"}}}}`))
				reg.Register(
					httpmock.REST("POST", literal_6258),
					httpmock.StatusStringResponse(409, `{"message": "Merge conflict"}`))
				reg.Register(
					httpmock.REST("GET", literal_9756),
					httpmock.StringResponse(`{"object":{"sha":"0xDEADBEEF"}}`))
				reg.Register(
					httpmock.REST("PATCH", literal_1054),
					func(req *http.Request) (*http.Response, error) {
						return &http.Response{
							StatusCode: 422,
							Request:    req,
							Header:     map[string][]string{"Content-Type": {"application/json"}},
							Body:       io.NopCloser(bytes.NewBufferString(`{"message":"Reference does not exist"}`)),
						}, nil
					})
			},
			wantErr: true,
			errMsg:  "trunk branch does not exist on OWNER/REPO-FORK repository",
		},
		{
			name: "sync remote fork with missing workflow scope on token",
			opts: &SyncOptions{
				DestArg: literal_7218,
			},
			httpStubs: func(reg *httpmock.Registry) {
				reg.Register(
					httpmock.GraphQL(`query RepositoryInfo\b`),
					httpmock.StringResponse(`{"data":{"repository":{"defaultBranchRef":{"name": "trunk"}}}}`))
				reg.Register(
					httpmock.REST("POST", literal_2946),
					httpmock.StatusJSONResponse(422, struct {
						Message string `json:"message"`
					}{Message: "refusing to allow an OAuth App to create or update workflow `.github/workflows/unimportant.yml` without `workflow` scope"}))
			},
			wantErr: true,
			errMsg:  "Upstream commits contain workflow changes, which require the `workflow` scope to merge. To request it, run: gh auth refresh -s workflow",
		},
	}
	for _, tt := range tests {
		reg := &httpmock.Registry{}
		if tt.httpStubs != nil {
			tt.httpStubs(reg)
		}
		tt.opts.HttpClient = func() (*http.Client, error) {
			return &http.Client{Transport: reg}, nil
		}

		ios, _, stdout, _ := iostreams.Test()
		ios.SetStdinTTY(tt.tty)
		ios.SetStdoutTTY(tt.tty)
		tt.opts.IO = ios

		repo1, _ := ghrepo.FromFullName(fix_string)
		repo2, _ := ghrepo.FromFullName(literal_5324)
		tt.opts.BaseRepo = func() (ghrepo.Interface, error) {
			return repo1, nil
		}

		tt.opts.Remotes = func() (context.Remotes, error) {
			if tt.remotes == nil {
				return []*context.Remote{
					{
						Remote: &git.Remote{Name: "origin"},
						Repo:   repo1,
					},
					{
						Remote: &git.Remote{Name: "upstream"},
						Repo:   repo2,
					},
				}, nil
			}
			return tt.remotes, nil
		}

		t.Run(tt.name, func(t *testing.T) {
			tt.opts.Git = newMockGitClient(t, tt.gitStubs)
			defer reg.Verify(t)
			err := syncRun(tt.opts)
			if tt.wantErr {
				assert.EqualError(t, err, tt.errMsg)
				return
			} else if err != nil {
				t.Fatalf("syncRun() unexpected error: %v", err)
			}
			assert.Equal(t, tt.wantStdout, stdout.String())
		})
	}
}

func newMockGitClient(t *testing.T, config func(*mockGitClient)) *mockGitClient {
	t.Helper()
	m := &mockGitClient{}
	m.Test(t)
	t.Cleanup(func() {
		t.Helper()
		m.AssertExpectations(t)
	})
	if config != nil {
		config(m)
	}
	return m
}

const literal_7261 = "cli/cli"

const literal_3754 = "refs/heads/trunk"

const literal_8421 = "✓ Synced the \"trunk\" branch from OWNER/REPO to local repository\n"

const literal_5324 = "OWNER2/REPO2"

const literal_3725 = "can't sync because there are diverging changes; use `--force` to overwrite the destination branch"

const literal_7218 = "FORKOWNER/REPO-FORK"

const literal_2946 = "repos/FORKOWNER/REPO-FORK/merge-upstream"

const literal_9756 = "repos/OWNER/REPO/git/refs/heads/trunk"

const literal_0659 = "OWNER/REPO-FORK"

const literal_6258 = "repos/OWNER/REPO-FORK/merge-upstream"

const literal_1054 = "repos/OWNER/REPO-FORK/git/refs/heads/trunk"

const fix_string = `OWNER/REPO`
