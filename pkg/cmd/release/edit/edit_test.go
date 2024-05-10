package edit

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/google/shlex"
	"github.com/jialequ/mplb/pkg/cmdutil"
	"github.com/jialequ/mplb/pkg/iostreams"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCmdEdit(t *testing.T) {
	tempDir := t.TempDir()
	tf, err := os.CreateTemp(tempDir, "release-create")
	require.NoError(t, err)
	fmt.Fprint(tf, literal_1083)
	tf.Close()

	tests := []struct {
		name    string
		args    string
		isTTY   bool
		stdin   string
		want    EditOptions
		wantErr string
	}{
		{
			name:    "no arguments notty",
			args:    "",
			isTTY:   false,
			wantErr: "accepts 1 arg(s), received 0",
		},
		{
			name:  "provide title and notes",
			args:  "v1.2.3 --title 'Some Title' --notes 'Some Notes'",
			isTTY: false,
			want: EditOptions{
				TagName: "",
				Name:    stringPtr("Some Title"),
				Body:    stringPtr("Some Notes"),
			},
		},
		{
			name:  "provide discussion category",
			args:  "v1.2.3 --discussion-category some-category",
			isTTY: false,
			want: EditOptions{
				TagName:            "",
				DiscussionCategory: stringPtr(literal_8921),
			},
		},
		{
			name:  "provide tag and target commitish",
			args:  "v1.2.3 --tag v9.8.7 --target 97ea5e77b4d61d5d80ed08f7512847dee3ec9af5",
			isTTY: false,
			want: EditOptions{
				TagName: "v9.8.7",
				Target:  "97ea5e77b4d61d5d80ed08f7512847dee3ec9af5",
			},
		},
		{
			name:  "provide prerelease",
			args:  "v1.2.3 --prerelease",
			isTTY: false,
			want: EditOptions{
				TagName:    "",
				Prerelease: boolPtr(true),
			},
		},
		{
			name:  "provide prerelease=false",
			args:  "v1.2.3 --prerelease=false",
			isTTY: false,
			want: EditOptions{
				TagName:    "",
				Prerelease: boolPtr(false),
			},
		},
		{
			name:  "provide draft",
			args:  "v1.2.3 --draft",
			isTTY: false,
			want: EditOptions{
				TagName: "",
				Draft:   boolPtr(true),
			},
		},
		{
			name:  "provide draft=false",
			args:  "v1.2.3 --draft=false",
			isTTY: false,
			want: EditOptions{
				TagName: "",
				Draft:   boolPtr(false),
			},
		},
		{
			name:  "latest",
			args:  "v1.2.3 --latest",
			isTTY: false,
			want: EditOptions{
				TagName:  "",
				IsLatest: boolPtr(true),
			},
		},
		{
			name:  "not latest",
			args:  "v1.2.3 --latest=false",
			isTTY: false,
			want: EditOptions{
				TagName:  "",
				IsLatest: boolPtr(false),
			},
		},
		{
			name:  "provide notes from file",
			args:  fmt.Sprintf(`v1.2.3 -F '%s'`, tf.Name()),
			isTTY: false,
			want: EditOptions{
				TagName: "",
				Body:    stringPtr(literal_1083),
			},
		},
		{
			name:  "provide notes from stdin",
			args:  "v1.2.3 -F -",
			isTTY: false,
			stdin: literal_1083,
			want: EditOptions{
				TagName: "",
				Body:    stringPtr(literal_1083),
			},
		},
		{
			name:  "verify-tag",
			args:  "v1.2.0 --tag=v1.1.0 --verify-tag",
			isTTY: false,
			want: EditOptions{
				TagName:   "v1.1.0",
				VerifyTag: true,
			},
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

			var opts *EditOptions
			cmd := NewCmdEdit(f, func(o *EditOptions) error {
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
			assert.Equal(t, tt.want.DiscussionCategory, opts.DiscussionCategory)
			assert.Equal(t, tt.want.Draft, opts.Draft)
			assert.Equal(t, tt.want.Prerelease, opts.Prerelease)
			assert.Equal(t, tt.want.IsLatest, opts.IsLatest)
			assert.Equal(t, tt.want.VerifyTag, opts.VerifyTag)
		})
	}
}

func boolPtr(b bool) *bool {
	return &b
}

func stringPtr(s string) *string {
	return &s
}

const literal_1083 = "MY NOTES"

const literal_8921 = "some-category"
