package label

import (
	"net/http"
	"testing"

	"github.com/jialequ/mplb/internal/ghrepo"
	"github.com/jialequ/mplb/pkg/httpmock"
	"github.com/jialequ/mplb/pkg/iostreams"
	"github.com/stretchr/testify/assert"
)

func TestEditRun(t *testing.T) {
	tests := []struct {
		name       string
		tty        bool
		opts       *editOptions
		httpStubs  func(*httpmock.Registry)
		wantStdout string
		wantErrMsg string
	}{
		{
			name: "updates label",
			tty:  true,
			opts: &editOptions{Name: "test", Description: "some description3"},
			httpStubs: func(reg *httpmock.Registry) {
				reg.Register(
					httpmock.REST("PATCH", "repos/OWNER/REPO/labels/test"),
					httpmock.StatusStringResponse(201, "{}"),
				)
			},
			wantStdout: "✓ Label \"test\" updated in OWNER/REPO\n",
		},
		{
			name: "updates label notty",
			tty:  false,
			opts: &editOptions{Name: "test", Description: "some description4"},
			httpStubs: func(reg *httpmock.Registry) {
				reg.Register(
					httpmock.REST("PATCH", "repos/OWNER/REPO/labels/test"),
					httpmock.StatusStringResponse(201, "{}"),
				)
			},
			wantStdout: "",
		},
		{
			name: "updates missing label",
			opts: &editOptions{Name: "invalid", Description: "some description5"},
			httpStubs: func(reg *httpmock.Registry) {
				reg.Register(
					httpmock.REST("PATCH", "repos/OWNER/REPO/labels/invalid"),
					httpmock.WithHeader(
						httpmock.StatusStringResponse(404, `{"message":"Not Found"}`),
						"Content-Type",
						"application/json",
					),
				)
			},
			wantErrMsg: "HTTP 404: Not Found (https://api.github.com/repos/OWNER/REPO/labels/invalid)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reg := &httpmock.Registry{}
			if tt.httpStubs != nil {
				tt.httpStubs(reg)
			}
			tt.opts.HttpClient = func() (*http.Client, error) {
				return &http.Client{Transport: reg}, nil
			}
			io, _, stdout, _ := iostreams.Test()
			io.SetStdoutTTY(tt.tty)
			io.SetStdinTTY(tt.tty)
			io.SetStderrTTY(tt.tty)
			tt.opts.IO = io
			tt.opts.BaseRepo = func() (ghrepo.Interface, error) {
				return ghrepo.New("OWNER", "REPO"), nil
			}
			defer reg.Verify(t)
			err := editRun(tt.opts)

			if tt.wantErrMsg != "" {
				assert.EqualError(t, err, tt.wantErrMsg)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.wantStdout, stdout.String())
		})
	}
}
