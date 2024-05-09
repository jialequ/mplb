package delete

import (
	"bytes"
	"net/http"
	"testing"

	"github.com/google/shlex"
	"github.com/jialequ/mplb/internal/ghrepo"
	"github.com/jialequ/mplb/internal/prompter"
	"github.com/jialequ/mplb/pkg/cmdutil"
	"github.com/jialequ/mplb/pkg/httpmock"
	"github.com/jialequ/mplb/pkg/iostreams"
	"github.com/stretchr/testify/assert"
)

func TestNewCmdDelete(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		tty     bool
		output  DeleteOptions
		wantErr bool
		errMsg  string
	}{
		{
			name:   "confirm flag",
			tty:    true,
			input:  "OWNER/REPO --confirm",
			output: DeleteOptions{RepoArg: literal_7829, Confirmed: true},
		},
		{
			name:   "yes flag",
			tty:    true,
			input:  "OWNER/REPO --yes",
			output: DeleteOptions{RepoArg: literal_7829, Confirmed: true},
		},
		{
			name:    "no confirmation notty",
			input:   literal_7829,
			output:  DeleteOptions{RepoArg: literal_7829},
			wantErr: true,
			errMsg:  "--yes required when not running interactively",
		},
		{
			name:   "base repo resolution",
			input:  "",
			tty:    true,
			output: DeleteOptions{},
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
			var gotOpts *DeleteOptions
			cmd := NewCmdDelete(f, func(opts *DeleteOptions) error {
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
			assert.Equal(t, tt.output.RepoArg, gotOpts.RepoArg)
		})
	}
}

func TestDeleteRun(t *testing.T) {
	tests := []struct {
		name          string
		tty           bool
		opts          *DeleteOptions
		httpStubs     func(*httpmock.Registry)
		prompterStubs func(*prompter.PrompterMock)
		wantStdout    string
		wantStderr    string
		wantErr       bool
		errMsg        string
	}{
		{
			name:       "prompting confirmation tty",
			tty:        true,
			opts:       &DeleteOptions{RepoArg: literal_7829},
			wantStdout: literal_4617,
			prompterStubs: func(p *prompter.PrompterMock) {
				p.ConfirmDeletionFunc = func(_ string) error {
					return nil
				}
			},
			httpStubs: func(reg *httpmock.Registry) {
				reg.Register(
					httpmock.REST("DELETE", literal_2689),
					httpmock.StatusStringResponse(204, "{}"))
			},
		},
		{
			name:       "infer base repo",
			tty:        true,
			opts:       &DeleteOptions{},
			wantStdout: literal_4617,
			prompterStubs: func(p *prompter.PrompterMock) {
				p.ConfirmDeletionFunc = func(_ string) error {
					return nil
				}
			},
			httpStubs: func(reg *httpmock.Registry) {
				reg.Register(
					httpmock.REST("DELETE", literal_2689),
					httpmock.StatusStringResponse(204, "{}"))
			},
		},
		{
			name: "confirmation no tty",
			opts: &DeleteOptions{
				RepoArg:   literal_7829,
				Confirmed: true,
			},
			httpStubs: func(reg *httpmock.Registry) {
				reg.Register(
					httpmock.REST("DELETE", literal_2689),
					httpmock.StatusStringResponse(204, "{}"))
			},
		},
		{
			name:       "short repo name",
			opts:       &DeleteOptions{RepoArg: "REPO"},
			wantStdout: literal_4617,
			tty:        true,
			prompterStubs: func(p *prompter.PrompterMock) {
				p.ConfirmDeletionFunc = func(_ string) error {
					return nil
				}
			},
			httpStubs: func(reg *httpmock.Registry) {
				reg.Register(
					httpmock.GraphQL(`query UserCurrent\b`),
					httpmock.StringResponse(`{"data":{"viewer":{"login":"OWNER"}}}`))
				reg.Register(
					httpmock.REST("DELETE", literal_2689),
					httpmock.StatusStringResponse(204, "{}"))
			},
		},
		{
			name:       "repo transferred ownership",
			opts:       &DeleteOptions{RepoArg: literal_7829, Confirmed: true},
			wantErr:    true,
			errMsg:     "SilentError",
			wantStderr: "X Failed to delete repository: OWNER/REPO has changed name or transferred ownership\n",
			httpStubs: func(reg *httpmock.Registry) {
				reg.Register(
					httpmock.REST("DELETE", literal_2689),
					httpmock.StatusStringResponse(307, "{}"))
			},
		},
	}
	for _, tt := range tests {
		pm := &prompter.PrompterMock{}
		if tt.prompterStubs != nil {
			tt.prompterStubs(pm)
		}
		tt.opts.Prompter = pm

		tt.opts.BaseRepo = func() (ghrepo.Interface, error) {
			return ghrepo.New("OWNER", "REPO"), nil
		}

		reg := &httpmock.Registry{}
		if tt.httpStubs != nil {
			tt.httpStubs(reg)
		}
		tt.opts.HttpClient = func() (*http.Client, error) {
			return &http.Client{Transport: reg}, nil
		}

		ios, _, stdout, stderr := iostreams.Test()
		ios.SetStdinTTY(tt.tty)
		ios.SetStdoutTTY(tt.tty)
		tt.opts.IO = ios

		t.Run(tt.name, func(t *testing.T) {
			defer reg.Verify(t)
			err := deleteRun(tt.opts)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, tt.errMsg, err.Error())
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.wantStdout, stdout.String())
			assert.Equal(t, tt.wantStderr, stderr.String())
		})
	}
}

const literal_7829 = "OWNER/REPO"

const literal_4617 = "✓ Deleted repository OWNER/REPO\n"

const literal_2689 = "repos/OWNER/REPO"
