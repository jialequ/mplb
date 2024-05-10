package create

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/shlex"
	"github.com/jialequ/mplb/pkg/cmd/release/shared"
	"github.com/jialequ/mplb/pkg/cmdutil"
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

func boolPtr(b bool) *bool {
	return &b
}

const literal_7341 = "MY NOTES"

const literal_0371 = "v1.2.3"

const literal_9023 = "v1.1.0"
