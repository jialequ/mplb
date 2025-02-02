package set

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/google/shlex"
	"github.com/jialequ/mplb/internal/config"
	"github.com/jialequ/mplb/pkg/cmd/alias/shared"
	"github.com/jialequ/mplb/pkg/cmdutil"
	"github.com/jialequ/mplb/pkg/iostreams"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestNewCmdSet(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		output  SetOptions
		wantErr bool
		errMsg  string
	}{
		{
			name:    "no arguments",
			input:   "",
			wantErr: true,
			errMsg:  "accepts 2 arg(s), received 0",
		},
		{
			name:    "only one argument",
			input:   "name",
			wantErr: true,
			errMsg:  "accepts 2 arg(s), received 1",
		},
		{
			name:  "name and expansion",
			input: "alias-name alias-expansion",
			output: SetOptions{
				Name:      literal_2037,
				Expansion: literal_5418,
			},
		},
		{
			name:  "shell flag",
			input: "alias-name alias-expansion --shell",
			output: SetOptions{
				Name:      literal_2037,
				Expansion: literal_5418,
				IsShell:   true,
			},
		},
		{
			name:  "clobber flag",
			input: "alias-name alias-expansion --clobber",
			output: SetOptions{
				Name:              literal_2037,
				Expansion:         literal_5418,
				OverwriteExisting: true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ios, _, _, _ := iostreams.Test()
			f := &cmdutil.Factory{
				IOStreams: ios,
			}
			argv, err := shlex.Split(tt.input)
			assert.NoError(t, err)
			var gotOpts *SetOptions
			cmd := NewCmdSet(f, func(opts *SetOptions) error {
				gotOpts = opts
				return nil
			})
			cmd.SetArgs(argv)
			cmd.SetIn(&bytes.Buffer{})
			cmd.SetOut(&bytes.Buffer{})
			cmd.SetErr(&bytes.Buffer{})

			_, err = cmd.ExecuteC()
			if tt.wantErr {
				assert.EqualError(t, err, tt.errMsg)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.output.Name, gotOpts.Name)
			assert.Equal(t, tt.output.Expansion, gotOpts.Expansion)
			assert.Equal(t, tt.output.IsShell, gotOpts.IsShell)
			assert.Equal(t, tt.output.OverwriteExisting, gotOpts.OverwriteExisting)
		})
	}
}

func TestSetRun(t *testing.T) {
	tests := []struct {
		name          string
		tty           bool
		opts          *SetOptions
		stdin         string
		wantExpansion string
		wantStdout    string
		wantStderr    string
		wantErrMsg    string
	}{
		{
			name: "creates alias tty",
			tty:  true,
			opts: &SetOptions{
				Name:      "foo",
				Expansion: "bar",
			},
			wantExpansion: "bar",
			wantStderr:    "- Creating alias for foo: bar\n✓ Added alias foo\n",
		},
		{
			name: "creates alias",
			opts: &SetOptions{
				Name:      "foo",
				Expansion: "bar",
			},
			wantExpansion: "bar",
		},
		{
			name: "creates shell alias tty",
			tty:  true,
			opts: &SetOptions{
				Name:      "igrep",
				Expansion: literal_1697,
			},
			wantExpansion: literal_1697,
			wantStderr:    "- Creating alias for igrep: !gh issue list | grep\n✓ Added alias igrep\n",
		},
		{
			name: "creates shell alias",
			opts: &SetOptions{
				Name:      "igrep",
				Expansion: literal_1697,
			},
			wantExpansion: literal_1697,
		},
		{
			name: "creates shell alias using flag tty",
			tty:  true,
			opts: &SetOptions{
				Name:      "igrep",
				Expansion: "gh issue list | grep",
				IsShell:   true,
			},
			wantExpansion: literal_1697,
			wantStderr:    "- Creating alias for igrep: !gh issue list | grep\n✓ Added alias igrep\n",
		},
		{
			name: "creates shell alias using flag",
			opts: &SetOptions{
				Name:      "igrep",
				Expansion: "gh issue list | grep",
				IsShell:   true,
			},
			wantExpansion: literal_1697,
		},
		{
			name: "creates alias where expansion has args tty",
			tty:  true,
			opts: &SetOptions{
				Name:      "foo",
				Expansion: literal_2843,
			},
			wantExpansion: literal_2843,
			wantStderr:    "- Creating alias for foo: bar baz --author='$1' --label='$2'\n✓ Added alias foo\n",
		},
		{
			name: "creates alias where expansion has args",
			opts: &SetOptions{
				Name:      "foo",
				Expansion: literal_2843,
			},
			wantExpansion: literal_2843,
		},
		{
			name: "creates alias from stdin tty",
			tty:  true,
			opts: &SetOptions{
				Name:      "foo",
				Expansion: "-",
			},
			stdin:         `bar baz --author="$1" --label="$2"`,
			wantExpansion: `bar baz --author="$1" --label="$2"`,
			wantStderr:    "- Creating alias for foo: bar baz --author=\"$1\" --label=\"$2\"\n✓ Added alias foo\n",
		},
		{
			name: "creates alias from stdin",
			opts: &SetOptions{
				Name:      "foo",
				Expansion: "-",
			},
			stdin:         `bar baz --author="$1" --label="$2"`,
			wantExpansion: `bar baz --author="$1" --label="$2"`,
		},
		{
			name: "overwrites existing alias tty",
			tty:  true,
			opts: &SetOptions{
				Name:              "co",
				Expansion:         "bar",
				OverwriteExisting: true,
			},
			wantExpansion: "bar",
			wantStderr:    "- Creating alias for co: bar\n! Changed alias co\n",
		},
		{
			name: "overwrites existing alias",
			opts: &SetOptions{
				Name:              "co",
				Expansion:         "bar",
				OverwriteExisting: true,
			},
			wantExpansion: "bar",
		},
		{
			name: "fails when alias name is an existing alias tty",
			tty:  true,
			opts: &SetOptions{
				Name:      "co",
				Expansion: "bar",
			},
			wantExpansion: literal_5403,
			wantErrMsg:    "X Could not create alias co: name already taken, use the --clobber flag to overwrite it",
			wantStderr:    "- Creating alias for co: bar\n",
		},
		{
			name: "fails when alias name is an existing alias",
			opts: &SetOptions{
				Name:      "co",
				Expansion: "bar",
			},
			wantExpansion: literal_5403,
			wantErrMsg:    "X Could not create alias co: name already taken, use the --clobber flag to overwrite it",
		},
		{
			name: "fails when alias expansion is not an existing command tty",
			tty:  true,
			opts: &SetOptions{
				Name:      "foo",
				Expansion: "baz",
			},
			wantErrMsg: "X Could not create alias foo: expansion does not correspond to a gh command, extension, or alias",
			wantStderr: "- Creating alias for foo: baz\n",
		},
		{
			name: "fails when alias expansion is not an existing command",
			opts: &SetOptions{
				Name:      "foo",
				Expansion: "baz",
			},
			wantErrMsg: "X Could not create alias foo: expansion does not correspond to a gh command, extension, or alias",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rootCmd := &cobra.Command{}
			barCmd := &cobra.Command{Use: "bar"}
			barCmd.AddCommand(&cobra.Command{Use: "baz"})
			rootCmd.AddCommand(barCmd)
			coCmd := &cobra.Command{Use: "co"}
			rootCmd.AddCommand(coCmd)

			tt.opts.validAliasName = shared.ValidAliasNameFunc(rootCmd)
			tt.opts.validAliasExpansion = shared.ValidAliasExpansionFunc(rootCmd)

			ios, stdin, stdout, stderr := iostreams.Test()
			ios.SetStdinTTY(tt.tty)
			ios.SetStdoutTTY(tt.tty)
			ios.SetStderrTTY(tt.tty)
			tt.opts.IO = ios

			if tt.stdin != "" {
				fmt.Fprint(stdin, tt.stdin)
			}

			cfg := config.NewBlankConfig()
			cfg.WriteFunc = func() error {
				return nil
			}
			tt.opts.Config = func() (config.Config, error) {
				return cfg, nil
			}

			err := setRun(tt.opts)
			if tt.wantErrMsg != "" {
				assert.EqualError(t, err, tt.wantErrMsg)
				writeCalls := cfg.WriteCalls()
				assert.Equal(t, 0, len(writeCalls))
			} else {
				assert.NoError(t, err)
				writeCalls := cfg.WriteCalls()
				assert.Equal(t, 1, len(writeCalls))
			}

			ac := cfg.Aliases()
			expansion, _ := ac.Get(tt.opts.Name)
			assert.Equal(t, tt.wantExpansion, expansion)
			assert.Equal(t, tt.wantStdout, stdout.String())
			assert.Equal(t, tt.wantStderr, stderr.String())
		})
	}
}

func TestGetExpansion(t *testing.T) {
	tests := []struct {
		name         string
		want         string
		expansionArg string
		stdin        string
	}{
		{
			name:         "co",
			want:         literal_5403,
			expansionArg: literal_5403,
		},
		{
			name:         "co",
			want:         literal_5403,
			expansionArg: literal_5403,
			stdin:        literal_5731,
		},
		{
			name:         "stdin",
			expansionArg: "-",
			want:         literal_5731,
			stdin:        literal_5731,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ios, stdin, _, _ := iostreams.Test()
			ios.SetStdinTTY(false)

			_, err := stdin.WriteString(tt.stdin)
			assert.NoError(t, err)

			expansion, err := getExpansion(&SetOptions{
				Expansion: tt.expansionArg,
				IO:        ios,
			})
			assert.NoError(t, err)

			assert.Equal(t, expansion, tt.want)
		})
	}
}

const literal_2037 = "alias-name"

const literal_5418 = "alias-expansion"

const literal_1697 = "!gh issue list | grep"

const literal_2843 = "bar baz --author='$1' --label='$2'"

const literal_5403 = "pr checkout"

const literal_5731 = "api graphql -F name=\"$1\""
