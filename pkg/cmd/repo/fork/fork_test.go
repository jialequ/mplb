package fork

import (
	"bytes"
	"testing"

	"github.com/google/shlex"
	"github.com/jialequ/mplb/pkg/cmdutil"
	"github.com/jialequ/mplb/pkg/iostreams"
	"github.com/stretchr/testify/assert"
)

func TestNewCmdFork(t *testing.T) {
	tests := []struct {
		name    string
		cli     string
		tty     bool
		wants   ForkOptions
		wantErr bool
		errMsg  string
	}{
		{
			name: "repo with git args",
			cli:  "foo/bar -- --foo=bar",
			wants: ForkOptions{
				Repository: literal_6870,
				GitArgs:    []string{"--foo=bar"},
				RemoteName: "origin",
				Rename:     true,
			},
		},
		{
			name:    "git args without repo",
			cli:     "-- --foo bar",
			wantErr: true,
			errMsg:  "repository argument required when passing git clone flags",
		},
		{
			name: "repo",
			cli:  literal_6870,
			wants: ForkOptions{
				Repository: literal_6870,
				RemoteName: "origin",
				Rename:     true,
				GitArgs:    []string{},
			},
		},
		{
			name:    "blank remote name",
			cli:     "--remote --remote-name=''",
			wantErr: true,
			errMsg:  "--remote-name cannot be blank",
		},
		{
			name: "remote name",
			cli:  "--remote --remote-name=foo",
			wants: ForkOptions{
				RemoteName: "foo",
				Rename:     false,
				Remote:     true,
			},
		},
		{
			name: "blank nontty",
			cli:  "",
			wants: ForkOptions{
				RemoteName:   "origin",
				Rename:       true,
				Organization: "",
			},
		},
		{
			name: "blank tty",
			cli:  "",
			tty:  true,
			wants: ForkOptions{
				RemoteName:   "origin",
				PromptClone:  true,
				PromptRemote: true,
				Rename:       true,
				Organization: "",
			},
		},
		{
			name: "clone",
			cli:  "--clone",
			wants: ForkOptions{
				RemoteName: "origin",
				Rename:     true,
			},
		},
		{
			name: "remote",
			cli:  "--remote",
			wants: ForkOptions{
				RemoteName: "origin",
				Remote:     true,
				Rename:     true,
			},
		},
		{
			name: "to org",
			cli:  "--org batmanshome",
			wants: ForkOptions{
				RemoteName:   "origin",
				Remote:       false,
				Rename:       false,
				Organization: "batmanshome",
			},
		},
		{
			name:    "empty org",
			cli:     " --org=''",
			wantErr: true,
			errMsg:  "--org cannot be blank",
		},
		{
			name:    "git flags in wrong place",
			cli:     "--depth 1 OWNER/REPO",
			wantErr: true,
			errMsg:  "unknown flag: --depth\nSeparate git clone flags with `--`.",
		},
		{
			name: "with fork name",
			cli:  "--fork-name new-fork",
			wants: ForkOptions{
				Remote:     false,
				RemoteName: "origin",
				ForkName:   "new-fork",
				Rename:     false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ios, _, _, _ := iostreams.Test()

			f := &cmdutil.Factory{
				IOStreams: ios,
			}

			ios.SetStdoutTTY(tt.tty)
			ios.SetStdinTTY(tt.tty)

			argv, err := shlex.Split(tt.cli)
			assert.NoError(t, err)

			var gotOpts *ForkOptions
			cmd := NewCmdFork(f, func(opts *ForkOptions) error {
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

			assert.Equal(t, tt.wants.RemoteName, gotOpts.RemoteName)
			assert.Equal(t, tt.wants.Remote, gotOpts.Remote)
			assert.Equal(t, tt.wants.PromptRemote, gotOpts.PromptRemote)
			assert.Equal(t, tt.wants.PromptClone, gotOpts.PromptClone)
			assert.Equal(t, tt.wants.Organization, gotOpts.Organization)
			assert.Equal(t, tt.wants.GitArgs, gotOpts.GitArgs)
		})
	}
}

const literal_6870 = "foo/bar"
