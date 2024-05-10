package view

import (
	"bytes"
	"io"
	"testing"

	"github.com/google/shlex"
	"github.com/jialequ/mplb/pkg/cmdutil"
	"github.com/jialequ/mplb/pkg/iostreams"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCmdView(t *testing.T) {
	tests := []struct {
		name    string
		args    string
		isTTY   bool
		want    ViewOptions
		wantErr string
	}{
		{
			name:  "no arguments",
			args:  "",
			isTTY: true,
			want: ViewOptions{
				ID:              "",
				WebMode:         false,
				IncludeParents:  true,
				InteractiveMode: true,
				Organization:    "",
			},
		},
		{
			name:  "only ID",
			args:  "3",
			isTTY: true,
			want: ViewOptions{
				ID:              "3",
				WebMode:         false,
				IncludeParents:  true,
				InteractiveMode: false,
				Organization:    "",
			},
		},
		{
			name:  "org",
			args:  "--org \"my-org\"",
			isTTY: true,
			want: ViewOptions{
				ID:              "",
				WebMode:         false,
				IncludeParents:  true,
				InteractiveMode: true,
				Organization:    "my-org",
			},
		},
		{
			name:  "web mode",
			args:  "--web",
			isTTY: true,
			want: ViewOptions{
				ID:              "",
				WebMode:         true,
				IncludeParents:  true,
				InteractiveMode: true,
				Organization:    "",
			},
		},
		{
			name:  "parents",
			args:  "--parents=false",
			isTTY: true,
			want: ViewOptions{
				ID:              "",
				WebMode:         false,
				IncludeParents:  false,
				InteractiveMode: true,
				Organization:    "",
			},
		},
		{
			name:    "repo and org specified",
			args:    "--org \"my-org\" -R \"owner/repo\"",
			isTTY:   true,
			wantErr: "only one of --repo and --org may be specified",
		},
		{
			name:    "invalid ID",
			args:    "1.5",
			isTTY:   true,
			wantErr: "invalid value for ruleset ID: 1.5 is not an integer",
		},
		{
			name:    "ID not provided and not TTY",
			args:    "",
			isTTY:   false,
			wantErr: "a ruleset ID must be provided when not running interactively",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ios, _, _, _ := iostreams.Test()
			ios.SetStdoutTTY(tt.isTTY)
			ios.SetStdinTTY(tt.isTTY)
			ios.SetStderrTTY(tt.isTTY)

			f := &cmdutil.Factory{
				IOStreams: ios,
			}

			var opts *ViewOptions
			cmd := NewCmdView(f, func(o *ViewOptions) error {
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

			assert.Equal(t, tt.want.ID, opts.ID)
			assert.Equal(t, tt.want.WebMode, opts.WebMode)
			assert.Equal(t, tt.want.IncludeParents, opts.IncludeParents)
			assert.Equal(t, tt.want.InteractiveMode, opts.InteractiveMode)
			assert.Equal(t, tt.want.Organization, opts.Organization)
		})
	}
}
