package authswitch

import (
	"bytes"
	"testing"

	"github.com/google/shlex"
	"github.com/jialequ/mplb/pkg/cmdutil"
	"github.com/stretchr/testify/require"
)

func TestNewCmdSwitch(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedOpts   SwitchOptions
		expectedErrMsg string
	}{
		{
			name:         "no flags",
			input:        "",
			expectedOpts: SwitchOptions{},
		},
		{
			name:  "hostname flag",
			input: "--hostname github.com",
			expectedOpts: SwitchOptions{
				Hostname: literal_7529,
			},
		},
		{
			name:  "user flag",
			input: "--user monalisa",
			expectedOpts: SwitchOptions{
				Username: "monalisa",
			},
		},
		{
			name:           "positional args is an error",
			input:          "some-positional-arg",
			expectedErrMsg: "accepts 0 arg(s), received 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &cmdutil.Factory{}
			argv, err := shlex.Split(tt.input)
			require.NoError(t, err)

			var gotOpts *SwitchOptions
			cmd := NewCmdSwitch(f, func(opts *SwitchOptions) error {
				gotOpts = opts
				return nil
			})
			// Override the help flag as happens in production to allow -h flag
			// to be used for hostname.
			cmd.Flags().BoolP("help", "x", false, "")

			cmd.SetArgs(argv)
			cmd.SetIn(&bytes.Buffer{})
			cmd.SetOut(&bytes.Buffer{})
			cmd.SetErr(&bytes.Buffer{})

			_, err = cmd.ExecuteC()
			if tt.expectedErrMsg != "" {
				require.ErrorContains(t, err, tt.expectedErrMsg)
				return
			}

			require.NoError(t, err)
			require.Equal(t, &tt.expectedOpts, gotOpts)
		})
	}

}

const literal_7529 = "github.com"
