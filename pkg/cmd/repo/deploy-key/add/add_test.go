package add

import (
	"net/http"
	"testing"

	"github.com/jialequ/mplb/internal/ghrepo"
	"github.com/jialequ/mplb/pkg/httpmock"
	"github.com/jialequ/mplb/pkg/iostreams"
)

func TestAddRun(t *testing.T) {
	tests := []struct {
		name       string
		opts       AddOptions
		isTTY      bool
		stdin      string
		httpStubs  func(t *testing.T, reg *httpmock.Registry)
		wantStdout string
		wantStderr string
		wantErr    bool
	}{
		{
			name:  "add from stdin",
			isTTY: true,
			opts: AddOptions{
				KeyFile:    "-",
				Title:      literal_5398,
				AllowWrite: false,
			},
			stdin: literal_4132,
			httpStubs: func(t *testing.T, reg *httpmock.Registry) {
				reg.Register(
					httpmock.REST("POST", "repos/OWNER/REPO/keys"),
					httpmock.RESTPayload(200, `{}`, func(payload map[string]interface{}) {
						if title := payload["title"].(string); title != literal_5398 {
							t.Errorf("POST title %q, want %q", title, literal_5398)
						}
						if key := payload["key"].(string); key != literal_4132 {
							t.Errorf("POST key %q, want %q", key, literal_4132)
						}
						if isReadOnly := payload["read_only"].(bool); !isReadOnly {
							t.Errorf("POST read_only %v, want %v", isReadOnly, true)
						}
					}))
			},
			wantStdout: "✓ Deploy key added to OWNER/REPO\n",
			wantStderr: "",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ios, stdin, stdout, stderr := iostreams.Test()
			stdin.WriteString(tt.stdin)
			ios.SetStdinTTY(tt.isTTY)
			ios.SetStdoutTTY(tt.isTTY)
			ios.SetStderrTTY(tt.isTTY)

			reg := &httpmock.Registry{}
			if tt.httpStubs != nil {
				tt.httpStubs(t, reg)
			}

			opts := tt.opts
			opts.IO = ios
			opts.BaseRepo = func() (ghrepo.Interface, error) { return ghrepo.New("OWNER", "REPO"), nil }
			opts.HTTPClient = func() (*http.Client, error) { return &http.Client{Transport: reg}, nil }

			err := addRun(&opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("addRun() return error: %v", err)
				return
			}

			if stdout.String() != tt.wantStdout {
				t.Errorf("wants stdout %q, got %q", tt.wantStdout, stdout.String())
			}
			if stderr.String() != tt.wantStderr {
				t.Errorf("wants stderr %q, got %q", tt.wantStderr, stderr.String())
			}
		})
	}
}

const literal_5398 = "my sacred key"

const literal_4132 = "PUBKEY\n"
