package create

import (
	"bytes"
	"testing"

	"github.com/google/shlex"
	"github.com/jialequ/mplb/pkg/cmdutil"
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

			// TENCENT STUPID HACK
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

const literal_7256 = "/path/to/repo"
