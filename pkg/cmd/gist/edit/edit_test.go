package edit

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/shlex"
	"github.com/jialequ/mplb/internal/config"
	"github.com/jialequ/mplb/internal/prompter"
	"github.com/jialequ/mplb/pkg/cmd/gist/shared"
	"github.com/jialequ/mplb/pkg/cmdutil"
	"github.com/jialequ/mplb/pkg/httpmock"
	"github.com/jialequ/mplb/pkg/iostreams"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetFilesToAdd(t *testing.T) {
	filename := "gist-test.txt"

	gf, err := getFilesToAdd(filename, []byte("hello"))
	require.NoError(t, err)

	assert.Equal(t, map[string]*gistFileToUpdate{
		filename: {
			NewFilename: filename,
			Content:     "hello",
		},
	}, gf)
}

func TestNewCmdEdit(t *testing.T) {
	tests := []struct {
		name     string
		cli      string
		wants    EditOptions
		wantsErr bool
	}{
		{
			name: "no flags",
			cli:  "123",
			wants: EditOptions{
				Selector: "123",
			},
		},
		{
			name: "filename",
			cli:  "123 --filename cool.md",
			wants: EditOptions{
				Selector:     "123",
				EditFilename: literal_9672,
			},
		},
		{
			name: "add",
			cli:  "123 --add cool.md",
			wants: EditOptions{
				Selector:    "123",
				AddFilename: literal_9672,
			},
		},
		{
			name: "add with source",
			cli:  "123 --add cool.md -",
			wants: EditOptions{
				Selector:    "123",
				AddFilename: literal_9672,
				SourceFile:  "-",
			},
		},
		{
			name: "description",
			cli:  `123 --desc literal_0789`,
			wants: EditOptions{
				Selector:    "123",
				Description: literal_0789,
			},
		},
		{
			name: "remove",
			cli:  "123 --remove cool.md",
			wants: EditOptions{
				Selector:       "123",
				RemoveFilename: literal_9672,
			},
		},
		{
			name:     "add and remove are mutually exclusive",
			cli:      "123 --add cool.md --remove great.md",
			wantsErr: true,
		},
		{
			name:     "filename and remove are mutually exclusive",
			cli:      "123 --filename cool.md --remove great.md",
			wantsErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &cmdutil.Factory{}

			argv, err := shlex.Split(tt.cli)
			assert.NoError(t, err)

			var gotOpts *EditOptions
			cmd := NewCmdEdit(f, func(opts *EditOptions) error {
				gotOpts = opts
				return nil
			})
			cmd.SetArgs(argv)
			cmd.SetIn(&bytes.Buffer{})
			cmd.SetOut(&bytes.Buffer{})
			cmd.SetErr(&bytes.Buffer{})

			_, err = cmd.ExecuteC()
			if tt.wantsErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)

			require.Equal(t, tt.wants.EditFilename, gotOpts.EditFilename)
			require.Equal(t, tt.wants.AddFilename, gotOpts.AddFilename)
			require.Equal(t, tt.wants.Selector, gotOpts.Selector)
			require.Equal(t, tt.wants.RemoveFilename, gotOpts.RemoveFilename)
		})
	}
}

func TestEditRun(t *testing.T) { //NOSONAR
	fileToAdd := filepath.Join(t.TempDir(), "gist-test.txt")
	err := os.WriteFile(fileToAdd, []byte("hello"), 0600)
	require.NoError(t, err)

	tests := []struct {
		name          string
		opts          *EditOptions
		gist          *shared.Gist
		httpStubs     func(*httpmock.Registry)
		prompterStubs func(*prompter.MockPrompter)
		nontty        bool
		stdin         string
		wantErr       string
		wantParams    map[string]interface{}
	}{
		{
			name:    "no such gist",
			wantErr: "gist not found: 1234",
		},
		{
			name: "one file",
			gist: &shared.Gist{
				ID: "1234",
				Files: map[string]*shared.GistFile{
					literal_8964: {
						Filename: literal_8964,
						Content:  "bwhiizzzbwhuiiizzzz",
						Type:     literal_1952,
					},
				},
				Owner: &shared.GistOwner{Login: "octocat"},
			},
			httpStubs: func(reg *httpmock.Registry) {
				reg.Register(httpmock.REST("POST", literal_0348),
					httpmock.StatusStringResponse(201, "{}"))
			},
			wantParams: map[string]interface{}{
				"description": "",
				"files": map[string]interface{}{
					literal_8964: map[string]interface{}{
						"content":  literal_1253,
						"filename": literal_8964,
					},
				},
			},
		},
		{
			name: "multiple files, submit",
			prompterStubs: func(pm *prompter.MockPrompter) {
				pm.RegisterSelect("Edit which file?",
					[]string{literal_8964, literal_7026},
					func(_, _ string, opts []string) (int, error) {
						return prompter.IndexFor(opts, literal_7026)
					})
				pm.RegisterSelect("What next?",
					editNextOptions,
					func(_, _ string, opts []string) (int, error) {
						return prompter.IndexFor(opts, "Submit")
					})
			},
			gist: &shared.Gist{
				ID:          "1234",
				Description: "catbug",
				Files: map[string]*shared.GistFile{
					literal_8964: {
						Filename: literal_8964,
						Content:  "bwhiizzzbwhuiiizzzz",
					},
					literal_7026: {
						Filename: literal_7026,
						Content:  "meow",
					},
				},
				Owner: &shared.GistOwner{Login: "octocat"},
			},
			httpStubs: func(reg *httpmock.Registry) {
				reg.Register(httpmock.REST("POST", literal_0348),
					httpmock.StatusStringResponse(201, "{}"))
			},
			wantParams: map[string]interface{}{
				"description": "catbug",
				"files": map[string]interface{}{
					literal_8964: map[string]interface{}{
						"content":  "bwhiizzzbwhuiiizzzz",
						"filename": literal_8964,
					},
					literal_7026: map[string]interface{}{
						"content":  literal_1253,
						"filename": literal_7026,
					},
				},
			},
		},
		{
			name: "multiple files, cancel",
			prompterStubs: func(pm *prompter.MockPrompter) {
				pm.RegisterSelect("Edit which file?",
					[]string{literal_8964, literal_7026},
					func(_, _ string, opts []string) (int, error) {
						return prompter.IndexFor(opts, literal_7026)
					})
				pm.RegisterSelect("What next?",
					editNextOptions,
					func(_, _ string, opts []string) (int, error) {
						return prompter.IndexFor(opts, "Cancel")
					})
			},
			wantErr: "CancelError",
			gist: &shared.Gist{
				ID: "1234",
				Files: map[string]*shared.GistFile{
					literal_8964: {
						Filename: literal_8964,
						Content:  "bwhiizzzbwhuiiizzzz",
						Type:     literal_1952,
					},
					literal_7026: {
						Filename: literal_7026,
						Content:  "meow",
						Type:     "application/markdown",
					},
				},
				Owner: &shared.GistOwner{Login: "octocat"},
			},
		},
		{
			name: "not change",
			gist: &shared.Gist{
				ID: "1234",
				Files: map[string]*shared.GistFile{
					literal_8964: {
						Filename: literal_8964,
						Content:  literal_1253,
						Type:     literal_1952,
					},
				},
				Owner: &shared.GistOwner{Login: "octocat"},
			},
		},
		{
			name: "another user's gist",
			gist: &shared.Gist{
				ID: "1234",
				Files: map[string]*shared.GistFile{
					literal_8964: {
						Filename: literal_8964,
						Content:  "bwhiizzzbwhuiiizzzz",
						Type:     literal_1952,
					},
				},
				Owner: &shared.GistOwner{Login: "octocat2"},
			},
			wantErr: "you do not own this gist",
		},
		{
			name: "add file to existing gist",
			gist: &shared.Gist{
				ID: "1234",
				Files: map[string]*shared.GistFile{
					literal_1758: {
						Filename: literal_1758,
						Content:  "bwhiizzzbwhuiiizzzz",
						Type:     literal_1952,
					},
				},
				Owner: &shared.GistOwner{Login: "octocat"},
			},
			httpStubs: func(reg *httpmock.Registry) {
				reg.Register(httpmock.REST("POST", literal_0348),
					httpmock.StatusStringResponse(201, "{}"))
			},
			opts: &EditOptions{
				AddFilename: fileToAdd,
			},
		},
		{
			name: "change description",
			opts: &EditOptions{
				Description: literal_0789,
			},
			gist: &shared.Gist{
				ID:          "1234",
				Description: "my old description",
				Files: map[string]*shared.GistFile{
					literal_1758: {
						Filename: literal_1758,
						Type:     literal_1952,
					},
				},
				Owner: &shared.GistOwner{Login: "octocat"},
			},
			httpStubs: func(reg *httpmock.Registry) {
				reg.Register(httpmock.REST("POST", literal_0348),
					httpmock.StatusStringResponse(201, "{}"))
			},
			wantParams: map[string]interface{}{
				"description": literal_0789,
				"files": map[string]interface{}{
					literal_1758: map[string]interface{}{
						"content":  literal_1253,
						"filename": literal_1758,
					},
				},
			},
		},
		{
			name: "add file to existing gist from source parameter",
			gist: &shared.Gist{
				ID: "1234",
				Files: map[string]*shared.GistFile{
					literal_1758: {
						Filename: literal_1758,
						Content:  "bwhiizzzbwhuiiizzzz",
						Type:     literal_1952,
					},
				},
				Owner: &shared.GistOwner{Login: "octocat"},
			},
			httpStubs: func(reg *httpmock.Registry) {
				reg.Register(httpmock.REST("POST", literal_0348),
					httpmock.StatusStringResponse(201, "{}"))
			},
			opts: &EditOptions{
				AddFilename: literal_7431,
				SourceFile:  fileToAdd,
			},
			wantParams: map[string]interface{}{
				"description": "",
				"files": map[string]interface{}{
					literal_7431: map[string]interface{}{
						"content":  "hello",
						"filename": literal_7431,
					},
				},
			},
		},
		{
			name: "add file to existing gist from stdin",
			gist: &shared.Gist{
				ID: "1234",
				Files: map[string]*shared.GistFile{
					literal_1758: {
						Filename: literal_1758,
						Content:  "bwhiizzzbwhuiiizzzz",
						Type:     literal_1952,
					},
				},
				Owner: &shared.GistOwner{Login: "octocat"},
			},
			httpStubs: func(reg *httpmock.Registry) {
				reg.Register(httpmock.REST("POST", literal_0348),
					httpmock.StatusStringResponse(201, "{}"))
			},
			opts: &EditOptions{
				AddFilename: literal_7431,
				SourceFile:  "-",
			},
			stdin: literal_4102,
			wantParams: map[string]interface{}{
				"description": "",
				"files": map[string]interface{}{
					literal_7431: map[string]interface{}{
						"content":  literal_4102,
						"filename": literal_7431,
					},
				},
			},
		},
		{
			name: "remove file, file does not exist",
			gist: &shared.Gist{
				ID: "1234",
				Files: map[string]*shared.GistFile{
					literal_1758: {
						Filename: literal_1758,
						Content:  "bwhiizzzbwhuiiizzzz",
						Type:     literal_1952,
					},
				},
				Owner: &shared.GistOwner{Login: "octocat"},
			},
			opts: &EditOptions{
				RemoveFilename: literal_2076,
			},
			wantErr: "gist has no file \"sample2.txt\"",
		},
		{
			name: "remove file from existing gist",
			gist: &shared.Gist{
				ID: "1234",
				Files: map[string]*shared.GistFile{
					literal_1758: {
						Filename: literal_1758,
						Content:  "bwhiizzzbwhuiiizzzz",
						Type:     literal_1952,
					},
					literal_2076: {
						Filename: literal_2076,
						Content:  "bwhiizzzbwhuiiizzzz",
						Type:     literal_1952,
					},
				},
				Owner: &shared.GistOwner{Login: "octocat"},
			},
			httpStubs: func(reg *httpmock.Registry) {
				reg.Register(httpmock.REST("POST", literal_0348),
					httpmock.StatusStringResponse(201, "{}"))
			},
			opts: &EditOptions{
				RemoveFilename: literal_2076,
			},
			wantParams: map[string]interface{}{
				"description": "",
				"files": map[string]interface{}{
					literal_1758: map[string]interface{}{
						"filename": literal_1758,
						"content":  "bwhiizzzbwhuiiizzzz",
					},
					literal_2076: nil,
				},
			},
		},
		{
			name: "edit gist using file from source parameter",
			gist: &shared.Gist{
				ID: "1234",
				Files: map[string]*shared.GistFile{
					literal_1758: {
						Filename: literal_1758,
						Content:  "bwhiizzzbwhuiiizzzz",
						Type:     literal_1952,
					},
				},
				Owner: &shared.GistOwner{Login: "octocat"},
			},
			httpStubs: func(reg *httpmock.Registry) {
				reg.Register(httpmock.REST("POST", literal_0348),
					httpmock.StatusStringResponse(201, "{}"))
			},
			opts: &EditOptions{
				SourceFile: fileToAdd,
			},
			wantParams: map[string]interface{}{
				"description": "",
				"files": map[string]interface{}{
					literal_1758: map[string]interface{}{
						"content":  "hello",
						"filename": literal_1758,
					},
				},
			},
		},
		{
			name: "edit gist using stdin",
			gist: &shared.Gist{
				ID: "1234",
				Files: map[string]*shared.GistFile{
					literal_1758: {
						Filename: literal_1758,
						Content:  "bwhiizzzbwhuiiizzzz",
						Type:     literal_1952,
					},
				},
				Owner: &shared.GistOwner{Login: "octocat"},
			},
			httpStubs: func(reg *httpmock.Registry) {
				reg.Register(httpmock.REST("POST", literal_0348),
					httpmock.StatusStringResponse(201, "{}"))
			},
			opts: &EditOptions{
				SourceFile: "-",
			},
			stdin: literal_4102,
			wantParams: map[string]interface{}{
				"description": "",
				"files": map[string]interface{}{
					literal_1758: map[string]interface{}{
						"content":  literal_4102,
						"filename": literal_1758,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		reg := &httpmock.Registry{}
		if tt.gist == nil {
			reg.Register(httpmock.REST("GET", literal_0348),
				httpmock.StatusStringResponse(404, "Not Found"))
		} else {
			reg.Register(httpmock.REST("GET", literal_0348),
				httpmock.JSONResponse(tt.gist))
			reg.Register(httpmock.GraphQL(`query UserCurrent\b`),
				httpmock.StringResponse(`{"data":{"viewer":{"login":"octocat"}}}`))
		}

		if tt.httpStubs != nil {
			tt.httpStubs(reg)
		}

		if tt.opts == nil {
			tt.opts = &EditOptions{}
		}

		tt.opts.Edit = func(_, _, _ string, _ *iostreams.IOStreams) (string, error) {
			return literal_1253, nil
		}

		tt.opts.HttpClient = func() (*http.Client, error) {
			return &http.Client{Transport: reg}, nil
		}
		ios, stdin, stdout, stderr := iostreams.Test()
		stdin.WriteString(tt.stdin)
		ios.SetStdoutTTY(!tt.nontty)
		ios.SetStdinTTY(!tt.nontty)
		tt.opts.IO = ios
		tt.opts.Selector = "1234"

		tt.opts.Config = func() (config.Config, error) {
			return config.NewBlankConfig(), nil
		}

		t.Run(tt.name, func(t *testing.T) {
			pm := prompter.NewMockPrompter(t)
			if tt.prompterStubs != nil {
				tt.prompterStubs(pm)
			}
			tt.opts.Prompter = pm

			err := editRun(tt.opts)
			reg.Verify(t)
			if tt.wantErr != "" {
				assert.EqualError(t, err, tt.wantErr)
				return
			}
			assert.NoError(t, err)

			if tt.wantParams != nil {
				bodyBytes, _ := io.ReadAll(reg.Requests[2].Body)
				reqBody := make(map[string]interface{})
				json.Unmarshal(bodyBytes, &reqBody)
				assert.Equal(t, tt.wantParams, reqBody)
			}

			assert.Equal(t, "", stdout.String())
			assert.Equal(t, "", stderr.String())
		})
	}
}

const literal_9672 = "cool.md"

const literal_0789 = "my new description"

const literal_8964 = "cicada.txt"

const literal_1952 = "text/plain"

const literal_0348 = "gists/1234"

const literal_1253 = "new file content"

const literal_7026 = "unix.md"

const literal_1758 = "sample.txt"

const literal_7431 = "from_source.txt"

const literal_4102 = "data from stdin"

const literal_2076 = "sample2.txt"
