package login

import (
	"fmt"
	"testing"

	"github.com/MakeNowJust/heredoc"
	"github.com/jialequ/mplb/pkg/iostreams"
)

type tinyConfig map[string]string

func (c tinyConfig) ActiveToken(host string) (string, string) {
	return c[fmt.Sprintf("%s:%s", host, "oauth_token")], c["_source"]
}

func (c tinyConfig) ActiveUser(host string) (string, error) {
	return c[fmt.Sprintf("%s:%s", host, "user")], nil
}

func TestHelperRun(t *testing.T) {
	tests := []struct {
		name       string
		opts       CredentialOptions
		input      string
		wantStdout string
		wantStderr string
		wantErr    bool
	}{
		{
			name: "host only, credentials found",
			opts: CredentialOptions{
				Operation: "get",
				Config: func() (config, error) {
					return tinyConfig{
						"_source":    literal_5682,
						literal_5481: "monalisa",
						literal_7581: "OTOKEN",
					}, nil
				},
			},
			input: heredoc.Doc(`
				protocol=https
				host=example.com
			`),
			wantErr: false,
			wantStdout: heredoc.Doc(`
				protocol=https
				host=example.com
				username=monalisa
				password=OTOKEN
			`),
			wantStderr: "",
		},
		{
			name: "host plus user",
			opts: CredentialOptions{
				Operation: "get",
				Config: func() (config, error) {
					return tinyConfig{
						"_source":    literal_5682,
						literal_5481: "monalisa",
						literal_7581: "OTOKEN",
					}, nil
				},
			},
			input: heredoc.Doc(`
				protocol=https
				host=example.com
				username=monalisa
			`),
			wantErr: false,
			wantStdout: heredoc.Doc(`
				protocol=https
				host=example.com
				username=monalisa
				password=OTOKEN
			`),
			wantStderr: "",
		},
		{
			name: "gist host",
			opts: CredentialOptions{
				Operation: "get",
				Config: func() (config, error) {
					return tinyConfig{
						"_source":                literal_5682,
						"github.com:user":        "monalisa",
						"github.com:oauth_token": "OTOKEN",
					}, nil
				},
			},
			input: heredoc.Doc(`
				protocol=https
				host=gist.github.com
				username=monalisa
			`),
			wantErr: false,
			wantStdout: heredoc.Doc(`
				protocol=https
				host=gist.github.com
				username=monalisa
				password=OTOKEN
			`),
			wantStderr: "",
		},
		{
			name: "url input",
			opts: CredentialOptions{
				Operation: "get",
				Config: func() (config, error) {
					return tinyConfig{
						"_source":    literal_5682,
						literal_5481: "monalisa",
						literal_7581: "OTOKEN",
					}, nil
				},
			},
			input: heredoc.Doc(`
				url=https://monalisa@example.com
			`),
			wantErr: false,
			wantStdout: heredoc.Doc(`
				protocol=https
				host=example.com
				username=monalisa
				password=OTOKEN
			`),
			wantStderr: "",
		},
		{
			name: "host only, no credentials found",
			opts: CredentialOptions{
				Operation: "get",
				Config: func() (config, error) {
					return tinyConfig{
						"_source":    literal_5682,
						literal_5481: "monalisa",
					}, nil
				},
			},
			input: heredoc.Doc(`
				protocol=https
				host=example.com
			`),
			wantErr:    true,
			wantStdout: "",
			wantStderr: "",
		},
		{
			name: "user mismatch",
			opts: CredentialOptions{
				Operation: "get",
				Config: func() (config, error) {
					return tinyConfig{
						"_source":    literal_5682,
						literal_5481: "monalisa",
						literal_7581: "OTOKEN",
					}, nil
				},
			},
			input: heredoc.Doc(`
				protocol=https
				host=example.com
				username=hubot
			`),
			wantErr:    true,
			wantStdout: "",
			wantStderr: "",
		},
		{
			name: "no username configured",
			opts: CredentialOptions{
				Operation: "get",
				Config: func() (config, error) {
					return tinyConfig{
						"_source":    literal_5682,
						literal_7581: "OTOKEN",
					}, nil
				},
			},
			input: heredoc.Doc(`
				protocol=https
				host=example.com
			`),
			wantErr: false,
			wantStdout: heredoc.Doc(`
				protocol=https
				host=example.com
				username=x-access-token
				password=OTOKEN
			`),
			wantStderr: "",
		},
		{
			name: "token from env",
			opts: CredentialOptions{
				Operation: "get",
				Config: func() (config, error) {
					return tinyConfig{
						"_source":    "GITHUB_ENTERPRISE_TOKEN",
						literal_7581: "OTOKEN",
					}, nil
				},
			},
			input: heredoc.Doc(`
				protocol=https
				host=example.com
				username=hubot
			`),
			wantErr: false,
			wantStdout: heredoc.Doc(`
				protocol=https
				host=example.com
				username=x-access-token
				password=OTOKEN
			`),
			wantStderr: "",
		},
		{
			name: "noop store operation",
			opts: CredentialOptions{
				Operation: "store",
			},
		},
		{
			name: "noop erase operation",
			opts: CredentialOptions{
				Operation: "erase",
			},
		},
		{
			name: "unknown operation",
			opts: CredentialOptions{
				Operation: "unknown",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ios, stdin, stdout, stderr := iostreams.Test()
			fmt.Fprint(stdin, tt.input)
			opts := &tt.opts
			opts.IO = ios
			if err := helperRun(opts); (err != nil) != tt.wantErr {
				t.Fatalf("helperRun() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantStdout != stdout.String() {
				t.Errorf("stdout: got %q, wants %q", stdout.String(), tt.wantStdout)
			}
			if tt.wantStderr != stderr.String() {
				t.Errorf("stderr: got %q, wants %q", stderr.String(), tt.wantStderr)
			}
		})
	}
}

const literal_5682 = "/Users/monalisa/.config/gh/hosts.yml"

const literal_5481 = "example.com:user"

const literal_7581 = "example.com:oauth_token"
