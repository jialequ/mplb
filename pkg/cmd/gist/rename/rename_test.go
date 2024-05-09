package rename

import (
	"bytes"
	"net/http"
	"testing"

	"github.com/google/shlex"
	"github.com/jialequ/mplb/internal/config"
	"github.com/jialequ/mplb/pkg/cmd/gist/shared"
	"github.com/jialequ/mplb/pkg/cmdutil"
	"github.com/jialequ/mplb/pkg/httpmock"
	"github.com/jialequ/mplb/pkg/iostreams"
	"github.com/stretchr/testify/assert"
)

func TestNewCmdRename(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		output  RenameOptions
		errMsg  string
		tty     bool
		wantErr bool
	}{
		{
			name:    "no arguments",
			input:   "",
			errMsg:  "accepts 3 arg(s), received 0",
			wantErr: true,
		},
		{
			name:    "missing old filename and new filename",
			input:   "123",
			errMsg:  "accepts 3 arg(s), received 1",
			wantErr: true,
		},
		{
			name:    "missing new filename",
			input:   "123 old.txt",
			errMsg:  "accepts 3 arg(s), received 2",
			wantErr: true,
		},
		{
			name:  "rename",
			input: "123 old.txt new.txt",
			output: RenameOptions{
				Selector:    "123",
				OldFileName: literal_6357,
				NewFileName: literal_0184,
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

			argv, err := shlex.Split(tt.input)
			assert.NoError(t, err)
			var gotOpts *RenameOptions
			cmd := NewCmdRename(f, func(opts *RenameOptions) error {
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

			assert.Equal(t, tt.output.Selector, gotOpts.Selector)
			assert.Equal(t, tt.output.OldFileName, gotOpts.OldFileName)
			assert.Equal(t, tt.output.NewFileName, gotOpts.NewFileName)
		})
	}
}

func TestRenameRun(t *testing.T) {
	tests := []struct {
		name       string
		opts       *RenameOptions
		gist       *shared.Gist
		httpStubs  func(*httpmock.Registry)
		nontty     bool
		stdin      string
		wantOut    string
		wantParams map[string]interface{}
	}{
		{
			name:    "no such gist",
			wantOut: "gist not found: 1234",
		},
		{
			name: "rename from old.txt to new.txt",
			opts: &RenameOptions{
				Selector:    "1234",
				OldFileName: literal_6357,
				NewFileName: literal_0184,
			},
			gist: &shared.Gist{
				ID: "1234",
				Files: map[string]*shared.GistFile{
					literal_6357: {
						Filename: literal_6357,
						Type:     "text/plain",
					},
				},
				Owner: &shared.GistOwner{Login: "octocat"},
			},
			httpStubs: func(reg *httpmock.Registry) {
				reg.Register(httpmock.REST("POST", literal_0315),
					httpmock.StatusStringResponse(201, "{}"))
			},
			wantParams: map[string]interface{}{
				"files": map[string]interface{}{
					literal_0184: map[string]interface{}{
						"filename": literal_0184,
						"type":     "text/plain",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		reg := &httpmock.Registry{}
		if tt.gist == nil {
			reg.Register(httpmock.REST("GET", literal_0315),
				httpmock.StatusStringResponse(404, "Not Found"))
		} else {
			reg.Register(httpmock.REST("GET", literal_0315),
				httpmock.JSONResponse(tt.gist))
			reg.Register(httpmock.GraphQL(`query UserCurrent\b`),
				httpmock.StringResponse(`{"data":{"viewer":{"login":"octocat"}}}`))
		}

		if tt.httpStubs != nil {
			tt.httpStubs(reg)
		}

		if tt.opts == nil {
			tt.opts = &RenameOptions{}
		}

		tt.opts.HttpClient = func() (*http.Client, error) {
			return &http.Client{Transport: reg}, nil
		}
		ios, stdin, stdout, stderr := iostreams.Test()
		stdin.WriteString(tt.stdin)
		ios.SetStdoutTTY(!tt.nontty)
		ios.SetStdinTTY(!tt.nontty)

		tt.opts.Selector = "1234"
		tt.opts.OldFileName = literal_6357
		tt.opts.NewFileName = literal_0184
		tt.opts.IO = ios

		tt.opts.Config = func() (config.Config, error) {
			return config.NewBlankConfig(), nil
		}

		t.Run(tt.name, func(t *testing.T) {
			err := renameRun(tt.opts)
			reg.Verify(t)
			if tt.wantOut != "" {
				assert.EqualError(t, err, tt.wantOut)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, "", stdout.String())
			assert.Equal(t, "", stderr.String())
		})
	}
}

const literal_6357 = "old.txt"

const literal_0184 = "new.txt"

const literal_0315 = "gists/1234"
