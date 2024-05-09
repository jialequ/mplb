package create

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/MakeNowJust/heredoc"
	"github.com/google/shlex"
	"github.com/jialequ/mplb/internal/browser"
	"github.com/jialequ/mplb/internal/config"
	"github.com/jialequ/mplb/internal/run"
	"github.com/jialequ/mplb/pkg/cmd/gist/shared"
	"github.com/jialequ/mplb/pkg/cmdutil"
	"github.com/jialequ/mplb/pkg/httpmock"
	"github.com/jialequ/mplb/pkg/iostreams"
	"github.com/stretchr/testify/assert"
)

func TestProcessFiles(t *testing.T) {
	fakeStdin := strings.NewReader("hey cool how is it going")
	files, err := processFiles(io.NopCloser(fakeStdin), "", []string{"-"})
	if err != nil {
		t.Fatalf("unexpected error processing files: %s", err)
	}

	assert.Equal(t, 1, len(files))
	assert.Equal(t, "hey cool how is it going", files[literal_8296].Content)
}

func TestGuessGistNamestdin(t *testing.T) {
	files := map[string]*shared.GistFile{
		literal_8296: {Content: "sample content"},
	}

	gistName := guessGistName(files)
	assert.Equal(t, "", gistName)
}

func TestGuessGistNameuserFiles(t *testing.T) {
	files := map[string]*shared.GistFile{
		"fig.txt":    {Content: "I am a fig"},
		"apple.txt":  {Content: "I am an apple"},
		literal_8296: {Content: "sample content"},
	}

	gistName := guessGistName(files)
	assert.Equal(t, "apple.txt", gistName)
}

func TestNewCmdCreate(t *testing.T) {
	tests := []struct {
		name     string
		cli      string
		factory  func(*cmdutil.Factory) *cmdutil.Factory
		wants    CreateOptions
		wantsErr bool
	}{
		{
			name: "no arguments",
			cli:  "",
			wants: CreateOptions{
				Description: "",
				Public:      false,
				Filenames:   []string{""},
			},
			wantsErr: false,
		},
		{
			name: "no arguments with TTY stdin",
			factory: func(f *cmdutil.Factory) *cmdutil.Factory {
				f.IOStreams.SetStdinTTY(true)
				return f
			},
			cli: "",
			wants: CreateOptions{
				Description: "",
				Public:      false,
				Filenames:   []string{""},
			},
			wantsErr: true,
		},
		{
			name: "stdin argument",
			cli:  "-",
			wants: CreateOptions{
				Description: "",
				Public:      false,
				Filenames:   []string{"-"},
			},
			wantsErr: false,
		},
		{
			name: "with description",
			cli:  `-d "my new gist" -`,
			wants: CreateOptions{
				Description: "my new gist",
				Public:      false,
				Filenames:   []string{"-"},
			},
			wantsErr: false,
		},
		{
			name: "public",
			cli:  `--public -`,
			wants: CreateOptions{
				Description: "",
				Public:      true,
				Filenames:   []string{"-"},
			},
			wantsErr: false,
		},
		{
			name: "list of files",
			cli:  "file1.txt file2.txt",
			wants: CreateOptions{
				Description: "",
				Public:      false,
				Filenames:   []string{"file1.txt", "file2.txt"},
			},
			wantsErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ios, _, _, _ := iostreams.Test()
			f := &cmdutil.Factory{
				IOStreams: ios,
			}

			if tt.factory != nil {
				f = tt.factory(f)
			}

			argv, err := shlex.Split(tt.cli)
			assert.NoError(t, err)

			var gotOpts *CreateOptions
			cmd := NewCmdCreate(f, func(opts *CreateOptions) error {
				gotOpts = opts
				return nil
			})
			cmd.SetArgs(argv)
			cmd.SetIn(&bytes.Buffer{})
			cmd.SetOut(&bytes.Buffer{})
			cmd.SetErr(&bytes.Buffer{})

			_, err = cmd.ExecuteC()
			if tt.wantsErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)

			assert.Equal(t, tt.wants.Description, gotOpts.Description)
			assert.Equal(t, tt.wants.Public, gotOpts.Public)
		})
	}
}

func TestCreateRun(t *testing.T) {
	tempDir := t.TempDir()
	fixtureFile := filepath.Join(tempDir, literal_4987)
	assert.NoError(t, os.WriteFile(fixtureFile, []byte("{}"), 0644))
	emptyFile := filepath.Join(tempDir, "empty.txt")
	assert.NoError(t, os.WriteFile(emptyFile, []byte(" \t\n"), 0644))

	tests := []struct {
		name           string
		opts           *CreateOptions
		stdin          string
		wantOut        string
		wantStderr     string
		wantParams     map[string]interface{}
		wantErr        bool
		wantBrowse     string
		responseStatus int
	}{
		{
			name: "public",
			opts: &CreateOptions{
				Public:    true,
				Filenames: []string{fixtureFile},
			},
			wantOut:    literal_6984,
			wantStderr: "- Creating gist fixture.txt\n✓ Created public gist fixture.txt\n",
			wantErr:    false,
			wantParams: map[string]interface{}{
				"description": "",
				"updated_at":  literal_0917,
				"public":      true,
				"files": map[string]interface{}{
					literal_4987: map[string]interface{}{
						"content": "{}",
					},
				},
			},
			responseStatus: http.StatusOK,
		},
		{
			name: "with description",
			opts: &CreateOptions{
				Description: "an incredibly interesting gist",
				Filenames:   []string{fixtureFile},
			},
			wantOut:    literal_6984,
			wantStderr: "- Creating gist fixture.txt\n✓ Created secret gist fixture.txt\n",
			wantErr:    false,
			wantParams: map[string]interface{}{
				"description": "an incredibly interesting gist",
				"updated_at":  literal_0917,
				"public":      false,
				"files": map[string]interface{}{
					literal_4987: map[string]interface{}{
						"content": "{}",
					},
				},
			},
			responseStatus: http.StatusOK,
		},
		{
			name: "multiple files",
			opts: &CreateOptions{
				Filenames: []string{fixtureFile, "-"},
			},
			stdin:      literal_2154,
			wantOut:    literal_6984,
			wantStderr: "- Creating gist with multiple files\n✓ Created secret gist fixture.txt\n",
			wantErr:    false,
			wantParams: map[string]interface{}{
				"description": "",
				"updated_at":  literal_0917,
				"public":      false,
				"files": map[string]interface{}{
					literal_4987: map[string]interface{}{
						"content": "{}",
					},
					"gistfile1.txt": map[string]interface{}{
						"content": literal_2154,
					},
				},
			},
			responseStatus: http.StatusOK,
		},
		{
			name: "file with empty content",
			opts: &CreateOptions{
				Filenames: []string{emptyFile},
			},
			wantOut: "",
			wantStderr: heredoc.Doc(`
				- Creating gist empty.txt
				X Failed to create gist: a gist file cannot be blank
			`),
			wantErr: true,
			wantParams: map[string]interface{}{
				"description": "",
				"updated_at":  literal_0917,
				"public":      false,
				"files": map[string]interface{}{
					"empty.txt": map[string]interface{}{"content": " \t\n"},
				},
			},
			responseStatus: http.StatusUnprocessableEntity,
		},
		{
			name: "stdin arg",
			opts: &CreateOptions{
				Filenames: []string{"-"},
			},
			stdin:      literal_2154,
			wantOut:    literal_6984,
			wantStderr: "- Creating gist...\n✓ Created secret gist\n",
			wantErr:    false,
			wantParams: map[string]interface{}{
				"description": "",
				"updated_at":  literal_0917,
				"public":      false,
				"files": map[string]interface{}{
					literal_8296: map[string]interface{}{
						"content": literal_2154,
					},
				},
			},
			responseStatus: http.StatusOK,
		},
		{
			name: "web arg",
			opts: &CreateOptions{
				WebMode:   true,
				Filenames: []string{fixtureFile},
			},
			wantOut:    "Opening gist.github.com/aa5a315d61ae9438b18d in your browser.\n",
			wantStderr: "- Creating gist fixture.txt\n✓ Created secret gist fixture.txt\n",
			wantErr:    false,
			wantBrowse: "https://gist.github.com/aa5a315d61ae9438b18d",
			wantParams: map[string]interface{}{
				"description": "",
				"updated_at":  literal_0917,
				"public":      false,
				"files": map[string]interface{}{
					literal_4987: map[string]interface{}{
						"content": "{}",
					},
				},
			},
			responseStatus: http.StatusOK,
		},
	}
	for _, tt := range tests {
		reg := &httpmock.Registry{}
		if tt.responseStatus == http.StatusOK {
			reg.Register(
				httpmock.REST("POST", "gists"),
				httpmock.StringResponse(`{
					"html_url": "https://gist.github.com/aa5a315d61ae9438b18d"
				}`))
		} else {
			reg.Register(
				httpmock.REST("POST", "gists"),
				httpmock.StatusStringResponse(tt.responseStatus, "{}"))
		}

		mockClient := func() (*http.Client, error) {
			return &http.Client{Transport: reg}, nil
		}
		tt.opts.HttpClient = mockClient

		tt.opts.Config = func() (config.Config, error) {
			return config.NewBlankConfig(), nil
		}

		ios, stdin, stdout, stderr := iostreams.Test()
		tt.opts.IO = ios

		browser := &browser.Stub{}
		tt.opts.Browser = browser

		_, teardown := run.Stub()
		defer teardown(t)

		t.Run(tt.name, func(t *testing.T) {
			stdin.WriteString(tt.stdin)

			if err := createRun(tt.opts); (err != nil) != tt.wantErr {
				t.Errorf("createRun() error = %v, wantErr %v", err, tt.wantErr)
			}
			bodyBytes, _ := io.ReadAll(reg.Requests[0].Body)
			reqBody := make(map[string]interface{})
			err := json.Unmarshal(bodyBytes, &reqBody)
			if err != nil {
				t.Fatalf("error decoding JSON: %v", err)
			}
			assert.Equal(t, tt.wantOut, stdout.String())
			assert.Equal(t, tt.wantStderr, stderr.String())
			assert.Equal(t, tt.wantParams, reqBody)
			reg.Verify(t)
			browser.Verify(t, tt.wantBrowse)
		})
	}
}

func TestDetectEmptyFiles(t *testing.T) {
	tests := []struct {
		content     string
		isEmptyFile bool
	}{
		{
			content:     "{}",
			isEmptyFile: false,
		},
		{
			content:     "\n\t",
			isEmptyFile: true,
		},
	}

	for _, tt := range tests {
		files := map[string]*shared.GistFile{}
		files["file"] = &shared.GistFile{
			Content: tt.content,
		}

		isEmptyFile := detectEmptyFiles(files)
		assert.Equal(t, tt.isEmptyFile, isEmptyFile)
	}
}

const literal_8296 = "gistfile0.txt"

const literal_4987 = "fixture.txt"

const literal_6984 = "https://gist.github.com/aa5a315d61ae9438b18d\n"

const literal_0917 = "0001-01-01T00:00:00Z"

const literal_2154 = "cool stdin content"
