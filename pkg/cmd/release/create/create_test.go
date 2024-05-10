package create

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/shlex"
	"github.com/jialequ/mplb/git"
	"github.com/jialequ/mplb/internal/ghrepo"
	"github.com/jialequ/mplb/internal/run"
	"github.com/jialequ/mplb/pkg/cmd/release/shared"
	"github.com/jialequ/mplb/pkg/cmdutil"
	"github.com/jialequ/mplb/pkg/httpmock"
	"github.com/jialequ/mplb/pkg/iostreams"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCmdCreate(t *testing.T) {
	tempDir := t.TempDir()
	tf, err := os.CreateTemp(tempDir, "release-create")
	require.NoError(t, err)
	fmt.Fprint(tf, literal_7341)
	tf.Close()
	af1, err := os.Create(filepath.Join(tempDir, "windows.zip"))
	require.NoError(t, err)
	af1.Close()
	af2, err := os.Create(filepath.Join(tempDir, "linux.tgz"))
	require.NoError(t, err)
	af2.Close()

	tests := []struct {
		name    string
		args    string
		isTTY   bool
		stdin   string
		want    CreateOptions
		wantErr string
	}{
		{
			name:  "no arguments tty",
			args:  "",
			isTTY: true,
			want: CreateOptions{
				TagName:      "",
				Target:       "",
				Name:         "",
				Body:         "",
				BodyProvided: false,
				Draft:        false,
				Prerelease:   false,
				RepoOverride: "",
				Concurrency:  5,
				Assets:       []*shared.AssetForUpload(nil),
				VerifyTag:    false,
			},
		},
		{
			name:    "no arguments notty",
			args:    "",
			isTTY:   false,
			wantErr: "tag required when not running interactively",
		},
		{
			name:  "only tag name",
			args:  literal_0371,
			isTTY: true,
			want: CreateOptions{
				TagName:      literal_0371,
				Target:       "",
				Name:         "",
				Body:         "",
				BodyProvided: false,
				Draft:        false,
				Prerelease:   false,
				RepoOverride: "",
				Concurrency:  5,
				Assets:       []*shared.AssetForUpload(nil),
				VerifyTag:    false,
			},
		},
		{
			name:  "asset files",
			args:  fmt.Sprintf("v1.2.3 '%s' '%s#Linux build'", af1.Name(), af2.Name()),
			isTTY: true,
			want: CreateOptions{
				TagName:      literal_0371,
				Target:       "",
				Name:         "",
				Body:         "",
				BodyProvided: false,
				Draft:        false,
				Prerelease:   false,
				RepoOverride: "",
				Concurrency:  5,
				Assets: []*shared.AssetForUpload{
					{
						Name:  "windows.zip",
						Label: "",
					},
					{
						Name:  "linux.tgz",
						Label: "Linux build",
					},
				},
			},
		},
		{
			name:  "provide title and body",
			args:  "v1.2.3 -t mytitle -n mynotes",
			isTTY: true,
			want: CreateOptions{
				TagName:      literal_0371,
				Target:       "",
				Name:         "mytitle",
				Body:         "mynotes",
				BodyProvided: true,
				Draft:        false,
				Prerelease:   false,
				RepoOverride: "",
				Concurrency:  5,
				Assets:       []*shared.AssetForUpload(nil),
			},
		},
		{
			name:  "notes from file",
			args:  fmt.Sprintf(`v1.2.3 -F '%s'`, tf.Name()),
			isTTY: true,
			want: CreateOptions{
				TagName:      literal_0371,
				Target:       "",
				Name:         "",
				Body:         literal_7341,
				BodyProvided: true,
				Draft:        false,
				Prerelease:   false,
				RepoOverride: "",
				Concurrency:  5,
				Assets:       []*shared.AssetForUpload(nil),
			},
		},
		{
			name:  "notes from stdin",
			args:  "v1.2.3 -F -",
			isTTY: true,
			stdin: literal_7341,
			want: CreateOptions{
				TagName:      literal_0371,
				Target:       "",
				Name:         "",
				Body:         literal_7341,
				BodyProvided: true,
				Draft:        false,
				Prerelease:   false,
				RepoOverride: "",
				Concurrency:  5,
				Assets:       []*shared.AssetForUpload(nil),
			},
		},
		{
			name:  "set draft and prerelease",
			args:  "v1.2.3 -d -p",
			isTTY: true,
			want: CreateOptions{
				TagName:      literal_0371,
				Target:       "",
				Name:         "",
				Body:         "",
				BodyProvided: false,
				Draft:        true,
				Prerelease:   true,
				RepoOverride: "",
				Concurrency:  5,
				Assets:       []*shared.AssetForUpload(nil),
			},
		},
		{
			name:  "discussion category",
			args:  "v1.2.3 --discussion-category 'General'",
			isTTY: true,
			want: CreateOptions{
				TagName:            literal_0371,
				Target:             "",
				Name:               "",
				Body:               "",
				BodyProvided:       false,
				Draft:              false,
				Prerelease:         false,
				RepoOverride:       "",
				Concurrency:        5,
				Assets:             []*shared.AssetForUpload(nil),
				DiscussionCategory: "General",
			},
		},
		{
			name:    "discussion category for draft release",
			args:    "v1.2.3 -d --discussion-category 'General'",
			isTTY:   true,
			wantErr: "discussions for draft releases not supported",
		},
		{
			name:  "generate release notes",
			args:  "v1.2.3 --generate-notes",
			isTTY: true,
			want: CreateOptions{
				TagName:       literal_0371,
				Target:        "",
				Name:          "",
				Body:          "",
				BodyProvided:  true,
				Draft:         false,
				Prerelease:    false,
				RepoOverride:  "",
				Concurrency:   5,
				Assets:        []*shared.AssetForUpload(nil),
				GenerateNotes: true,
			},
		},
		{
			name:  "generate release notes with notes tag",
			args:  "v1.2.3 --generate-notes --notes-start-tag v1.1.0",
			isTTY: true,
			want: CreateOptions{
				TagName:       literal_0371,
				Target:        "",
				Name:          "",
				Body:          "",
				BodyProvided:  true,
				Draft:         false,
				Prerelease:    false,
				RepoOverride:  "",
				Concurrency:   5,
				Assets:        []*shared.AssetForUpload(nil),
				GenerateNotes: true,
				NotesStartTag: literal_9023,
			},
		},
		{
			name:  "notes tag",
			args:  "--notes-start-tag v1.1.0",
			isTTY: true,
			want: CreateOptions{
				TagName:       "",
				Target:        "",
				Name:          "",
				Body:          "",
				BodyProvided:  false,
				Draft:         false,
				Prerelease:    false,
				RepoOverride:  "",
				Concurrency:   5,
				Assets:        []*shared.AssetForUpload(nil),
				GenerateNotes: false,
				NotesStartTag: literal_9023,
			},
		},
		{
			name:  "latest",
			args:  "--latest v1.1.0",
			isTTY: false,
			want: CreateOptions{
				TagName:       literal_9023,
				Target:        "",
				Name:          "",
				Body:          "",
				BodyProvided:  false,
				Draft:         false,
				Prerelease:    false,
				IsLatest:      boolPtr(true),
				RepoOverride:  "",
				Concurrency:   5,
				Assets:        []*shared.AssetForUpload(nil),
				GenerateNotes: false,
				NotesStartTag: "",
			},
		},
		{
			name:  "not latest",
			args:  "--latest=false v1.1.0",
			isTTY: false,
			want: CreateOptions{
				TagName:       literal_9023,
				Target:        "",
				Name:          "",
				Body:          "",
				BodyProvided:  false,
				Draft:         false,
				Prerelease:    false,
				IsLatest:      boolPtr(false),
				RepoOverride:  "",
				Concurrency:   5,
				Assets:        []*shared.AssetForUpload(nil),
				GenerateNotes: false,
				NotesStartTag: "",
			},
		},
		{
			name:  "with verify-tag",
			args:  "v1.1.0 --verify-tag",
			isTTY: true,
			want: CreateOptions{
				TagName:       literal_9023,
				Target:        "",
				Name:          "",
				Body:          "",
				BodyProvided:  false,
				Draft:         false,
				Prerelease:    false,
				RepoOverride:  "",
				Concurrency:   5,
				Assets:        []*shared.AssetForUpload(nil),
				GenerateNotes: false,
				VerifyTag:     true,
			},
		},
		{
			name:  "with --notes-from-tag",
			args:  "v1.2.3 --notes-from-tag",
			isTTY: false,
			want: CreateOptions{
				TagName:      literal_0371,
				BodyProvided: true,
				Concurrency:  5,
				Assets:       []*shared.AssetForUpload(nil),
				NotesFromTag: true,
			},
		},
		{
			name:    "with --notes-from-tag and --generate-notes",
			args:    "v1.2.3 --notes-from-tag --generate-notes",
			isTTY:   false,
			wantErr: "using `--notes-from-tag` with `--generate-notes` or `--notes-start-tag` is not supported",
		},
		{
			name:    "with --notes-from-tag and --notes-start-tag",
			args:    "v1.2.3 --notes-from-tag --notes-start-tag v1.2.3",
			isTTY:   false,
			wantErr: "using `--notes-from-tag` with `--generate-notes` or `--notes-start-tag` is not supported",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ios, stdin, _, _ := iostreams.Test()
			if tt.stdin == "" {
				ios.SetStdinTTY(tt.isTTY)
			} else {
				ios.SetStdinTTY(false)
				fmt.Fprint(stdin, tt.stdin)
			}
			ios.SetStdoutTTY(tt.isTTY)
			ios.SetStderrTTY(tt.isTTY)

			f := &cmdutil.Factory{
				IOStreams: ios,
			}

			var opts *CreateOptions
			cmd := NewCmdCreate(f, func(o *CreateOptions) error {
				opts = o
				return nil
			})
			cmd.PersistentFlags().StringP("repo", "R", "", "")

			argv, err := shlex.Split(tt.args)
			require.NoError(t, err)
			cmd.SetArgs(argv)

			cmd.SetIn(&bytes.Buffer{})
			cmd.SetOut(io.Discard)
			cmd.SetErr(io.Discard)

			_, err = cmd.ExecuteC()
			if tt.wantErr != "" {
				require.EqualError(t, err, tt.wantErr)
				return
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, tt.want.TagName, opts.TagName)
			assert.Equal(t, tt.want.Target, opts.Target)
			assert.Equal(t, tt.want.Name, opts.Name)
			assert.Equal(t, tt.want.Body, opts.Body)
			assert.Equal(t, tt.want.BodyProvided, opts.BodyProvided)
			assert.Equal(t, tt.want.Draft, opts.Draft)
			assert.Equal(t, tt.want.Prerelease, opts.Prerelease)
			assert.Equal(t, tt.want.Concurrency, opts.Concurrency)
			assert.Equal(t, tt.want.RepoOverride, opts.RepoOverride)
			assert.Equal(t, tt.want.DiscussionCategory, opts.DiscussionCategory)
			assert.Equal(t, tt.want.GenerateNotes, opts.GenerateNotes)
			assert.Equal(t, tt.want.NotesStartTag, opts.NotesStartTag)
			assert.Equal(t, tt.want.IsLatest, opts.IsLatest)
			assert.Equal(t, tt.want.VerifyTag, opts.VerifyTag)
			assert.Equal(t, tt.want.NotesFromTag, opts.NotesFromTag)

			require.Equal(t, len(tt.want.Assets), len(opts.Assets))
			for i := range tt.want.Assets {
				assert.Equal(t, tt.want.Assets[i].Name, opts.Assets[i].Name)
				assert.Equal(t, tt.want.Assets[i].Label, opts.Assets[i].Label)
			}
		})
	}
}

func TestCreateRun(t *testing.T) {
	tests := []struct {
		name       string
		isTTY      bool
		opts       CreateOptions
		httpStubs  func(t *testing.T, reg *httpmock.Registry)
		runStubs   func(rs *run.CommandStubber)
		wantErr    string
		wantStdout string
		wantStderr string
	}{
		{
			name:  "create a release",
			isTTY: true,
			opts: CreateOptions{
				TagName:      literal_0371,
				Name:         literal_3478,
				Body:         literal_9045,
				BodyProvided: true,
				Target:       "",
			},
			runStubs: func(rs *run.CommandStubber) {
				rs.Register(`git tag --list`, 0, "")
			},
			httpStubs: func(t *testing.T, reg *httpmock.Registry) {
				reg.Register(httpmock.REST("POST", literal_5049), httpmock.RESTPayload(201, `{
					"url": "https://api.github.com/releases/123",
					"upload_url": "https://api.github.com/assets/upload",
					"html_url": "https://github.com/OWNER/REPO/releases/tag/v1.2.3"
				}`, func(params map[string]interface{}) {
					assert.Equal(t, map[string]interface{}{
						"tag_name":   literal_0371,
						"name":       literal_3478,
						"body":       literal_9045,
						"draft":      false,
						"prerelease": false,
					}, params)
				}))
			},
			wantStdout: literal_2958,
			wantStderr: ``,
		},
		{
			name:  "with discussion category",
			isTTY: true,
			opts: CreateOptions{
				TagName:            literal_0371,
				Name:               literal_3478,
				Body:               literal_9045,
				BodyProvided:       true,
				Target:             "",
				DiscussionCategory: "General",
			},
			runStubs: func(rs *run.CommandStubber) {
				rs.Register(`git tag --list`, 0, "")
			},
			httpStubs: func(t *testing.T, reg *httpmock.Registry) {
				reg.Register(httpmock.REST("POST", literal_5049), httpmock.RESTPayload(201, `{
					"url": "https://api.github.com/releases/123",
					"upload_url": "https://api.github.com/assets/upload",
					"html_url": "https://github.com/OWNER/REPO/releases/tag/v1.2.3"
				}`, func(params map[string]interface{}) {
					assert.Equal(t, map[string]interface{}{
						"tag_name":                 literal_0371,
						"name":                     literal_3478,
						"body":                     literal_9045,
						"draft":                    false,
						"prerelease":               false,
						"discussion_category_name": "General",
					}, params)
				}))
			},
			wantStdout: literal_2958,
			wantStderr: ``,
		},
		{
			name:  "with target commitish",
			isTTY: true,
			opts: CreateOptions{
				TagName:      literal_0371,
				Name:         "",
				Body:         "",
				BodyProvided: true,
				Target:       "main",
			},
			runStubs: func(rs *run.CommandStubber) {
				rs.Register(`git tag --list`, 0, "")
			},
			httpStubs: func(t *testing.T, reg *httpmock.Registry) {
				reg.Register(httpmock.REST("POST", literal_5049), httpmock.RESTPayload(201, `{
					"url": "https://api.github.com/releases/123",
					"upload_url": "https://api.github.com/assets/upload",
					"html_url": "https://github.com/OWNER/REPO/releases/tag/v1.2.3"
				}`, func(params map[string]interface{}) {
					assert.Equal(t, map[string]interface{}{
						"tag_name":         literal_0371,
						"draft":            false,
						"prerelease":       false,
						"target_commitish": "main",
					}, params)
				}))
			},
			wantStdout: literal_2958,
			wantStderr: ``,
		},
		{
			name:  "as draft",
			isTTY: true,
			opts: CreateOptions{
				TagName:      literal_0371,
				Name:         "",
				Body:         "",
				BodyProvided: true,
				Draft:        true,
				Target:       "",
			},
			runStubs: func(rs *run.CommandStubber) {
				rs.Register(`git tag --list`, 0, "")
			},
			httpStubs: func(t *testing.T, reg *httpmock.Registry) {
				reg.Register(httpmock.REST("POST", literal_5049), httpmock.RESTPayload(201, `{
					"url": "https://api.github.com/releases/123",
					"upload_url": "https://api.github.com/assets/upload",
					"html_url": "https://github.com/OWNER/REPO/releases/tag/v1.2.3"
				}`, func(params map[string]interface{}) {
					assert.Equal(t, map[string]interface{}{
						"tag_name":   literal_0371,
						"draft":      true,
						"prerelease": false,
					}, params)
				}))
			},
			wantStdout: literal_2958,
			wantStderr: ``,
		},
		{
			name:  "with latest",
			isTTY: false,
			opts: CreateOptions{
				TagName:       literal_0371,
				Name:          "",
				Body:          "",
				Target:        "",
				IsLatest:      boolPtr(true),
				BodyProvided:  true,
				GenerateNotes: false,
			},
			runStubs: func(rs *run.CommandStubber) {
				rs.Register(`git tag --list`, 0, "")
			},
			httpStubs: func(t *testing.T, reg *httpmock.Registry) {
				reg.Register(httpmock.REST("POST", literal_5049), httpmock.RESTPayload(201, `{
					"url": "https://api.github.com/releases/123",
					"upload_url": "https://api.github.com/assets/upload",
					"html_url": "https://github.com/OWNER/REPO/releases/tag/v1.2.3"
				}`, func(params map[string]interface{}) {
					assert.Equal(t, map[string]interface{}{
						"tag_name":    literal_0371,
						"draft":       false,
						"prerelease":  false,
						"make_latest": "true",
					}, params)
				}))
			},
			wantStdout: literal_2958,
			wantErr:    "",
		},
		{
			name:  "with generate notes",
			isTTY: true,
			opts: CreateOptions{
				TagName:       literal_0371,
				Name:          "",
				Body:          "",
				Target:        "",
				BodyProvided:  true,
				GenerateNotes: true,
			},
			runStubs: func(rs *run.CommandStubber) {
				rs.Register(`git tag --list`, 0, "")
			},
			httpStubs: func(t *testing.T, reg *httpmock.Registry) {
				reg.Register(httpmock.REST("POST", literal_5049), httpmock.RESTPayload(201, `{
					"url": "https://api.github.com/releases/123",
					"upload_url": "https://api.github.com/assets/upload",
					"html_url": "https://github.com/OWNER/REPO/releases/tag/v1.2.3"
				}`, func(params map[string]interface{}) {
					assert.Equal(t, map[string]interface{}{
						"tag_name":               literal_0371,
						"draft":                  false,
						"prerelease":             false,
						"generate_release_notes": true,
					}, params)
				}))
			},
			wantStdout: literal_2958,
			wantErr:    "",
		},
		{
			name:  "with generate notes and notes tag",
			isTTY: true,
			opts: CreateOptions{
				TagName:       literal_0371,
				Name:          "",
				Body:          "",
				Target:        "",
				BodyProvided:  true,
				GenerateNotes: true,
				NotesStartTag: literal_9023,
			},
			runStubs: func(rs *run.CommandStubber) {
				rs.Register(`git tag --list`, 0, "")
			},
			httpStubs: func(t *testing.T, reg *httpmock.Registry) {
				reg.Register(httpmock.REST("POST", literal_8267),
					httpmock.RESTPayload(200, `{
						"name": literal_0548,
						"body": literal_3702
				}`, func(params map[string]interface{}) {
						assert.Equal(t, map[string]interface{}{
							"tag_name":          literal_0371,
							"previous_tag_name": literal_9023,
						}, params)
					}))
				reg.Register(httpmock.REST("POST", literal_5049), httpmock.RESTPayload(201, `{
					"url": "https://api.github.com/releases/123",
					"upload_url": "https://api.github.com/assets/upload",
					"html_url": "https://github.com/OWNER/REPO/releases/tag/v1.2.3"
				}`, func(params map[string]interface{}) {
					assert.Equal(t, map[string]interface{}{
						"tag_name":   literal_0371,
						"draft":      false,
						"prerelease": false,
						"body":       literal_3702,
						"name":       literal_0548,
					}, params)
				}))
			},
			wantStdout: literal_2958,
			wantErr:    "",
		},
		{
			name:  "with generate notes and notes tag and body and name",
			isTTY: true,
			opts: CreateOptions{
				TagName:       literal_0371,
				Name:          "name",
				Body:          "body",
				Target:        "",
				BodyProvided:  true,
				GenerateNotes: true,
				NotesStartTag: literal_9023,
			},
			runStubs: func(rs *run.CommandStubber) {
				rs.Register(`git tag --list`, 0, "")
			},
			httpStubs: func(t *testing.T, reg *httpmock.Registry) {
				reg.Register(httpmock.REST("POST", literal_8267),
					httpmock.RESTPayload(200, `{
						"name": literal_0548,
						"body": literal_3702
				}`, func(params map[string]interface{}) {
						assert.Equal(t, map[string]interface{}{
							"tag_name":          literal_0371,
							"previous_tag_name": literal_9023,
						}, params)
					}))
				reg.Register(httpmock.REST("POST", literal_5049), httpmock.RESTPayload(201, `{
					"url": "https://api.github.com/releases/123",
					"upload_url": "https://api.github.com/assets/upload",
					"html_url": "https://github.com/OWNER/REPO/releases/tag/v1.2.3"
				}`, func(params map[string]interface{}) {
					assert.Equal(t, map[string]interface{}{
						"tag_name":   literal_0371,
						"draft":      false,
						"prerelease": false,
						"body":       "body\ngenerated body",
						"name":       "name",
					}, params)
				}))
			},
			wantStdout: literal_2958,
			wantErr:    "",
		},
		{
			name:  "publish after uploading files",
			isTTY: true,
			opts: CreateOptions{
				TagName:      literal_0371,
				Name:         "",
				Body:         "",
				BodyProvided: true,
				Draft:        false,
				Target:       "",
				Assets: []*shared.AssetForUpload{
					{
						Name: literal_3597,
						Open: func() (io.ReadCloser, error) {
							return io.NopCloser(bytes.NewBufferString(`TARBALL`)), nil
						},
					},
				},
				Concurrency: 1,
			},
			runStubs: func(rs *run.CommandStubber) {
				rs.Register(`git tag --list`, 0, "")
			},
			httpStubs: func(t *testing.T, reg *httpmock.Registry) {
				reg.Register(httpmock.REST("HEAD", literal_2468), httpmock.StatusStringResponse(404, ``))
				reg.Register(httpmock.REST("POST", literal_5049), httpmock.RESTPayload(201, `{
					"url": "https://api.github.com/releases/123",
					"upload_url": "https://api.github.com/assets/upload",
					"html_url": "https://github.com/OWNER/REPO/releases/tag/v1.2.3"
				}`, func(params map[string]interface{}) {
					assert.Equal(t, map[string]interface{}{
						"tag_name":   literal_0371,
						"draft":      true,
						"prerelease": false,
					}, params)
				}))
				reg.Register(httpmock.REST("POST", literal_2634), func(req *http.Request) (*http.Response, error) {
					q := req.URL.Query()
					assert.Equal(t, literal_3597, q.Get("name"))
					assert.Equal(t, "", q.Get("label"))
					return &http.Response{
						StatusCode: 201,
						Request:    req,
						Body:       io.NopCloser(bytes.NewBufferString(`{}`)),
						Header: map[string][]string{
							literal_8213: {literal_1563},
						},
					}, nil
				})
				reg.Register(httpmock.REST("PATCH", literal_9538), httpmock.RESTPayload(201, `{
					"html_url": "https://github.com/OWNER/REPO/releases/tag/v1.2.3-final"
				}`, func(params map[string]interface{}) {
					assert.Equal(t, map[string]interface{}{
						"draft": false,
					}, params)
				}))
			},
			wantStdout: literal_0938,
			wantStderr: ``,
		},
		{
			name:  "publish after uploading files, but do not mark as latest",
			isTTY: true,
			opts: CreateOptions{
				TagName:      literal_0371,
				Name:         "",
				Body:         "",
				BodyProvided: true,
				Draft:        false,
				IsLatest:     boolPtr(false),
				Target:       "",
				Assets: []*shared.AssetForUpload{
					{
						Name: literal_3597,
						Open: func() (io.ReadCloser, error) {
							return io.NopCloser(bytes.NewBufferString(`TARBALL`)), nil
						},
					},
				},
				Concurrency: 1,
			},
			runStubs: func(rs *run.CommandStubber) {
				rs.Register(`git tag --list`, 0, "")
			},
			httpStubs: func(t *testing.T, reg *httpmock.Registry) {
				reg.Register(httpmock.REST("HEAD", literal_2468), httpmock.StatusStringResponse(404, ``))
				reg.Register(httpmock.REST("POST", literal_5049), httpmock.RESTPayload(201, `{
					"url": "https://api.github.com/releases/123",
					"upload_url": "https://api.github.com/assets/upload",
					"html_url": "https://github.com/OWNER/REPO/releases/tag/v1.2.3"
				}`, func(params map[string]interface{}) {
					assert.Equal(t, map[string]interface{}{
						"tag_name":    literal_0371,
						"draft":       true,
						"prerelease":  false,
						"make_latest": "false",
					}, params)
				}))
				reg.Register(httpmock.REST("POST", literal_2634), func(req *http.Request) (*http.Response, error) {
					q := req.URL.Query()
					assert.Equal(t, literal_3597, q.Get("name"))
					assert.Equal(t, "", q.Get("label"))
					return &http.Response{
						StatusCode: 201,
						Request:    req,
						Body:       io.NopCloser(bytes.NewBufferString(`{}`)),
						Header: map[string][]string{
							literal_8213: {literal_1563},
						},
					}, nil
				})
				reg.Register(httpmock.REST("PATCH", literal_9538), httpmock.RESTPayload(201, `{
					"html_url": "https://github.com/OWNER/REPO/releases/tag/v1.2.3-final"
				}`, func(params map[string]interface{}) {
					assert.Equal(t, map[string]interface{}{
						"draft":       false,
						"make_latest": "false",
					}, params)
				}))
			},
			wantStdout: literal_0938,
			wantStderr: ``,
		},
		{
			name:  "upload files but release already exists",
			isTTY: true,
			opts: CreateOptions{
				TagName:      literal_0371,
				Name:         "",
				Body:         "",
				BodyProvided: true,
				Draft:        false,
				Target:       "",
				Assets: []*shared.AssetForUpload{
					{
						Name: literal_3597,
						Open: func() (io.ReadCloser, error) {
							return io.NopCloser(bytes.NewBufferString(`TARBALL`)), nil
						},
					},
				},
				Concurrency: 1,
			},
			runStubs: func(rs *run.CommandStubber) {
				rs.Register(`git tag --list`, 0, "")
			},
			httpStubs: func(t *testing.T, reg *httpmock.Registry) {
				reg.Register(httpmock.REST("HEAD", literal_2468), httpmock.StatusStringResponse(200, ``))
			},
			wantStdout: ``,
			wantStderr: ``,
			wantErr:    `a release with the same tag name already exists: v1.2.3`,
		},
		{
			name:  "clean up draft after uploading files fails",
			isTTY: false,
			opts: CreateOptions{
				TagName:      literal_0371,
				Name:         "",
				Body:         "",
				BodyProvided: true,
				Draft:        false,
				Target:       "",
				Assets: []*shared.AssetForUpload{
					{
						Name: literal_3597,
						Open: func() (io.ReadCloser, error) {
							return io.NopCloser(bytes.NewBufferString(`TARBALL`)), nil
						},
					},
				},
				Concurrency: 1,
			},
			runStubs: func(rs *run.CommandStubber) {
				rs.Register(`git tag --list`, 0, "")
			},
			httpStubs: func(t *testing.T, reg *httpmock.Registry) {
				reg.Register(httpmock.REST("HEAD", literal_2468), httpmock.StatusStringResponse(404, ``))
				reg.Register(httpmock.REST("POST", literal_5049), httpmock.StatusStringResponse(201, `{
					"url": "https://api.github.com/releases/123",
					"upload_url": "https://api.github.com/assets/upload",
					"html_url": "https://github.com/OWNER/REPO/releases/tag/v1.2.3"
				}`))
				reg.Register(httpmock.REST("POST", literal_2634), httpmock.StatusStringResponse(422, `{}`))
				reg.Register(httpmock.REST("DELETE", literal_9538), httpmock.StatusStringResponse(204, ``))
			},
			wantStdout: ``,
			wantStderr: ``,
			wantErr:    `HTTP 422 (https://api.github.com/assets/upload?label=&name=ball.tgz)`,
		},
		{
			name:  "clean up draft after publishing fails",
			isTTY: false,
			opts: CreateOptions{
				TagName:      literal_0371,
				Name:         "",
				Body:         "",
				BodyProvided: true,
				Draft:        false,
				Target:       "",
				Assets: []*shared.AssetForUpload{
					{
						Name: literal_3597,
						Open: func() (io.ReadCloser, error) {
							return io.NopCloser(bytes.NewBufferString(`TARBALL`)), nil
						},
					},
				},
				Concurrency: 1,
			},
			runStubs: func(rs *run.CommandStubber) {
				rs.Register(`git tag --list`, 0, "")
			},
			httpStubs: func(t *testing.T, reg *httpmock.Registry) {
				reg.Register(httpmock.REST("HEAD", literal_2468), httpmock.StatusStringResponse(404, ``))
				reg.Register(httpmock.REST("POST", literal_5049), httpmock.StatusStringResponse(201, `{
					"url": "https://api.github.com/releases/123",
					"upload_url": "https://api.github.com/assets/upload",
					"html_url": "https://github.com/OWNER/REPO/releases/tag/v1.2.3"
				}`))
				reg.Register(httpmock.REST("POST", literal_2634), httpmock.StatusStringResponse(201, `{}`))
				reg.Register(httpmock.REST("PATCH", literal_9538), httpmock.StatusStringResponse(500, `{}`))
				reg.Register(httpmock.REST("DELETE", literal_9538), httpmock.StatusStringResponse(204, ``))
			},
			wantStdout: ``,
			wantStderr: ``,
			wantErr:    `HTTP 500 (https://api.github.com/releases/123)`,
		},
		{
			name:  "upload files but release already exists",
			isTTY: true,
			opts: CreateOptions{
				TagName:      literal_0371,
				Name:         "",
				Body:         "",
				BodyProvided: true,
				Draft:        false,
				Target:       "",
				Assets: []*shared.AssetForUpload{
					{
						Name: literal_3597,
						Open: func() (io.ReadCloser, error) {
							return io.NopCloser(bytes.NewBufferString(`TARBALL`)), nil
						},
					},
				},
				Concurrency: 1,
			},
			runStubs: func(rs *run.CommandStubber) {
				rs.Register(`git tag --list`, 0, "")
			},
			httpStubs: func(t *testing.T, reg *httpmock.Registry) {
				reg.Register(httpmock.REST("HEAD", literal_2468), httpmock.StatusStringResponse(200, ``))
			},
			wantStdout: ``,
			wantStderr: ``,
			wantErr:    `a release with the same tag name already exists: v1.2.3`,
		},
		{
			name:  "upload files and create discussion",
			isTTY: true,
			opts: CreateOptions{
				TagName:      literal_0371,
				Name:         "",
				Body:         "",
				BodyProvided: true,
				Draft:        false,
				Target:       "",
				Assets: []*shared.AssetForUpload{
					{
						Name: literal_3597,
						Open: func() (io.ReadCloser, error) {
							return io.NopCloser(bytes.NewBufferString(`TARBALL`)), nil
						},
					},
				},
				DiscussionCategory: "general",
				Concurrency:        1,
			},
			runStubs: func(rs *run.CommandStubber) {
				rs.Register(`git tag --list`, 0, "")
			},
			httpStubs: func(t *testing.T, reg *httpmock.Registry) {
				reg.Register(httpmock.REST("HEAD", literal_2468), httpmock.StatusStringResponse(404, ``))
				reg.Register(httpmock.REST("POST", literal_5049), httpmock.RESTPayload(201, `{
					"url": "https://api.github.com/releases/123",
					"upload_url": "https://api.github.com/assets/upload",
					"html_url": "https://github.com/OWNER/REPO/releases/tag/v1.2.3"
				}`, func(params map[string]interface{}) {
					assert.Equal(t, map[string]interface{}{
						"tag_name":                 literal_0371,
						"draft":                    true,
						"prerelease":               false,
						"discussion_category_name": "general",
					}, params)
				}))
				reg.Register(httpmock.REST("POST", literal_2634), func(req *http.Request) (*http.Response, error) {
					q := req.URL.Query()
					assert.Equal(t, literal_3597, q.Get("name"))
					assert.Equal(t, "", q.Get("label"))
					return &http.Response{
						StatusCode: 201,
						Request:    req,
						Body:       io.NopCloser(bytes.NewBufferString(`{}`)),
						Header: map[string][]string{
							literal_8213: {literal_1563},
						},
					}, nil
				})
				reg.Register(httpmock.REST("PATCH", literal_9538), httpmock.RESTPayload(201, `{
					"html_url": "https://github.com/OWNER/REPO/releases/tag/v1.2.3-final"
				}`, func(params map[string]interface{}) {
					assert.Equal(t, map[string]interface{}{
						"draft":                    false,
						"discussion_category_name": "general",
					}, params)
				}))
			},
			wantStdout: literal_0938,
			wantStderr: ``,
		},
		{
			name:  "with generate notes from tag",
			isTTY: false,
			opts: CreateOptions{
				TagName:      literal_0371,
				BodyProvided: true,
				Concurrency:  5,
				Assets:       []*shared.AssetForUpload(nil),
				NotesFromTag: true,
			},
			runStubs: func(rs *run.CommandStubber) {
				rs.Register(`git tag --list`, 0, literal_2630)
			},
			httpStubs: func(t *testing.T, reg *httpmock.Registry) {
				reg.Register(
					httpmock.GraphQL("RepositoryFindRef"),
					httpmock.StringResponse(`{"data":{"repository":{"ref": {"id": "tag id"}}}}`),
				)
				reg.Register(
					httpmock.REST("POST", literal_5049),
					httpmock.RESTPayload(201, `{
						"url": "https://api.github.com/releases/123",
						"upload_url": "https://api.github.com/assets/upload",
						"html_url": "https://github.com/OWNER/REPO/releases/tag/v1.2.3"
					}`, func(payload map[string]interface{}) {
						assert.Equal(t, map[string]interface{}{
							"tag_name":   literal_0371,
							"draft":      false,
							"prerelease": false,
							"body":       literal_2630,
						}, payload)
					}))
			},
			wantStdout: literal_2958,
			wantStderr: "",
		},
		{
			name:  "with generate notes from tag and notes provided",
			isTTY: false,
			opts: CreateOptions{
				TagName:      literal_0371,
				Body:         "some notes here",
				BodyProvided: true,
				Concurrency:  5,
				Assets:       []*shared.AssetForUpload(nil),
				NotesFromTag: true,
			},
			runStubs: func(rs *run.CommandStubber) {
				rs.Register(`git tag --list`, 0, literal_2630)
			},
			httpStubs: func(t *testing.T, reg *httpmock.Registry) {
				reg.Register(
					httpmock.GraphQL("RepositoryFindRef"),
					httpmock.StringResponse(`{"data":{"repository":{"ref": {"id": "tag id"}}}}`),
				)
				reg.Register(
					httpmock.REST("POST", literal_5049),
					httpmock.RESTPayload(201, `{
						"url": "https://api.github.com/releases/123",
						"upload_url": "https://api.github.com/assets/upload",
						"html_url": "https://github.com/OWNER/REPO/releases/tag/v1.2.3"
					}`, func(payload map[string]interface{}) {
						assert.Equal(t, map[string]interface{}{
							"tag_name":   literal_0371,
							"draft":      false,
							"prerelease": false,
							"body":       "some notes here\nsome tag message",
						}, payload)
					}))
			},
			wantStdout: literal_2958,
			wantStderr: "",
		},
		{
			name:  "with generate notes from tag and tag does not exist",
			isTTY: false,
			opts: CreateOptions{
				TagName:      literal_0371,
				BodyProvided: true,
				Concurrency:  5,
				Assets:       []*shared.AssetForUpload(nil),
				NotesFromTag: true,
			},
			runStubs: func(rs *run.CommandStubber) {
				rs.Register(`git tag --list`, 0, "")
			},
			wantErr: "cannot generate release notes from tag v1.2.3 as it does not exist locally",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ios, _, stdout, stderr := iostreams.Test()
			ios.SetStdoutTTY(tt.isTTY)
			ios.SetStdinTTY(tt.isTTY)
			ios.SetStderrTTY(tt.isTTY)

			fakeHTTP := &httpmock.Registry{}
			if tt.httpStubs != nil {
				tt.httpStubs(t, fakeHTTP)
			}
			defer fakeHTTP.Verify(t)

			tt.opts.IO = ios
			tt.opts.HttpClient = func() (*http.Client, error) {
				return &http.Client{Transport: fakeHTTP}, nil
			}
			tt.opts.BaseRepo = func() (ghrepo.Interface, error) {
				return ghrepo.FromFullName("OWNER/REPO")
			}

			tt.opts.GitClient = &git.Client{GitPath: "some/path/git"}

			rs, teardown := run.Stub()
			defer teardown(t)
			if tt.runStubs != nil {
				tt.runStubs(rs)
			}

			err := createRun(&tt.opts)
			if tt.wantErr != "" {
				require.EqualError(t, err, tt.wantErr)
				return
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, tt.wantStdout, stdout.String())
			assert.Equal(t, tt.wantStderr, stderr.String())
		})
	}
}

func boolPtr(b bool) *bool {
	return &b
}

const literal_7341 = "MY NOTES"

const literal_0371 = "v1.2.3"

const literal_9023 = "v1.1.0"

const literal_3478 = "The Big 1.2"

const literal_9045 = "* Fixed bugs"

const literal_5049 = "repos/OWNER/REPO/releases"

const literal_2958 = "https://github.com/OWNER/REPO/releases/tag/v1.2.3\n"

const literal_8267 = "repos/OWNER/REPO/releases/generate-notes"

const literal_3702 = "generated body"

const literal_0548 = "generated name"

const literal_3597 = "ball.tgz"

const literal_2468 = "repos/OWNER/REPO/releases/tags/v1.2.3"

const literal_2634 = "assets/upload"

const literal_8213 = "Content-Type"

const literal_1563 = "application/json"

const literal_9538 = "releases/123"

const literal_0938 = "https://github.com/OWNER/REPO/releases/tag/v1.2.3-final\n"

const literal_2630 = "some tag message"
