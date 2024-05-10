package rerun

import (
	"bytes"
	"io"
	"testing"

	"github.com/google/shlex"
	"github.com/jialequ/mplb/pkg/cmdutil"
	"github.com/jialequ/mplb/pkg/iostreams"
	"github.com/stretchr/testify/assert"
)

func TestNewCmdRerun(t *testing.T) {
	tests := []struct {
		name     string
		cli      string
		tty      bool
		wants    RerunOptions
		wantsErr bool
	}{
		{
			name:     "blank nontty",
			wantsErr: true,
		},
		{
			name: "blank tty",
			tty:  true,
			wants: RerunOptions{
				Prompt: true,
			},
		},
		{
			name: "with arg nontty",
			cli:  "1234",
			wants: RerunOptions{
				RunID: "1234",
			},
		},
		{
			name: "with arg tty",
			tty:  true,
			cli:  "1234",
			wants: RerunOptions{
				RunID: "1234",
			},
		},
		{
			name: "failed arg nontty",
			cli:  "4321 --failed",
			wants: RerunOptions{
				RunID:      "4321",
				OnlyFailed: true,
			},
		},
		{
			name: "failed arg",
			tty:  true,
			cli:  "--failed",
			wants: RerunOptions{
				Prompt:     true,
				OnlyFailed: true,
			},
		},
		{
			name: "with arg job",
			tty:  true,
			cli:  "--job 1234",
			wants: RerunOptions{
				JobID: "1234",
			},
		},
		{
			name: "with args jobID and runID uses jobID",
			tty:  true,
			cli:  "1234 --job 5678",
			wants: RerunOptions{
				JobID: "5678",
				RunID: "",
			},
		},
		{
			name:     "with arg job with no ID fails",
			tty:      true,
			cli:      "--job",
			wantsErr: true,
		},
		{
			name:     "with arg job with no ID no tty fails",
			cli:      "--job",
			wantsErr: true,
		},
		{
			name: "debug nontty",
			cli:  "4321 --debug",
			wants: RerunOptions{
				RunID: "4321",
				Debug: true,
			},
		},
		{
			name: "debug tty",
			tty:  true,
			cli:  "--debug",
			wants: RerunOptions{
				Prompt: true,
				Debug:  true,
			},
		},
		{
			name: "debug off",
			cli:  "4321 --debug=false",
			wants: RerunOptions{
				RunID: "4321",
				Debug: false,
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

			argv, err := shlex.Split(tt.cli)
			assert.NoError(t, err)

			var gotOpts *RerunOptions
			cmd := NewCmdRerun(f, func(opts *RerunOptions) error {
				gotOpts = opts
				return nil
			})
			cmd.SetArgs(argv)
			cmd.SetIn(&bytes.Buffer{})
			cmd.SetOut(io.Discard)
			cmd.SetErr(io.Discard)

			_, err = cmd.ExecuteC()
			if tt.wantsErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)

			assert.Equal(t, tt.wants.RunID, gotOpts.RunID)
			assert.Equal(t, tt.wants.Prompt, gotOpts.Prompt)
		})
	}

}
