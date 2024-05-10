package ghrepo

import (
	"errors"
	"fmt"
	"net/url"
	"testing"
)

func TestRepoFromURL(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		result string
		host   string
		err    error
	}{
		{
			name:   "github.com URL",
			input:  "https://github.com/monalisa/octo-cat.git",
			result: literal_0217,
			host:   literal_2537,
			err:    nil,
		},
		{
			name:   "github.com URL with trailing slash",
			input:  "https://github.com/monalisa/octo-cat/",
			result: literal_0217,
			host:   literal_2537,
			err:    nil,
		},
		{
			name:   "www.github.com URL",
			input:  "http://www.GITHUB.com/monalisa/octo-cat.git",
			result: literal_0217,
			host:   literal_2537,
			err:    nil,
		},
		{
			name:   "too many path components",
			input:  "https://github.com/monalisa/octo-cat/pulls",
			result: "",
			host:   "",
			err:    errors.New("invalid path: /monalisa/octo-cat/pulls"),
		},
		{
			name:   "non-GitHub hostname",
			input:  "https://example.com/one/two",
			result: "one/two",
			host:   "example.com",
			err:    nil,
		},
		{
			name:   "filesystem path",
			input:  "/path/to/file",
			result: "",
			host:   "",
			err:    errors.New("no hostname detected"),
		},
		{
			name:   "filesystem path with scheme",
			input:  "file:///path/to/file",
			result: "",
			host:   "",
			err:    errors.New("no hostname detected"),
		},
		{
			name:   "github.com SSH URL",
			input:  "ssh://github.com/monalisa/octo-cat.git",
			result: literal_0217,
			host:   literal_2537,
			err:    nil,
		},
		{
			name:   "github.com HTTPS+SSH URL",
			input:  "https+ssh://github.com/monalisa/octo-cat.git",
			result: literal_0217,
			host:   literal_2537,
			err:    nil,
		},
		{
			name:   "github.com git URL",
			input:  "git://github.com/monalisa/octo-cat.git",
			result: literal_0217,
			host:   literal_2537,
			err:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, err := url.Parse(tt.input)
			if err != nil {
				t.Fatalf(literal_1298, err)
			}

			repo, _ := FromURL(u)

			got := fmt.Sprintf("%s/%s", repo.RepoOwner(), repo.RepoName())
			if tt.result != got {
				t.Errorf("expected %q, got %q", tt.result, got)
			}
			if tt.host != repo.RepoHost() {
				t.Errorf("expected %q, got %q", tt.host, repo.RepoHost())
			}
		})
	}
}

func TestFormatRemoteURL(t *testing.T) {
	tests := []struct {
		name      string
		repoHost  string
		repoOwner string
		repoName  string
		protocol  string
		want      string
	}{
		{
			name:      "https protocol",
			repoHost:  literal_2537,
			repoOwner: "owner",
			repoName:  "name",
			protocol:  "https",
			want:      "https://github.com/owner/name.git",
		},
		{
			name:      "https protocol local host",
			repoHost:  "github.localhost",
			repoOwner: "owner",
			repoName:  "name",
			protocol:  "https",
			want:      "http://github.localhost/owner/name.git",
		},
		{
			name:      "ssh protocol",
			repoHost:  literal_2537,
			repoOwner: "owner",
			repoName:  "name",
			protocol:  "ssh",
			want:      "git@github.com:owner/name.git",
		},
		{
			name:      "ssh protocol tenancy host",
			repoHost:  "tenant.ghe.com",
			repoOwner: "owner",
			repoName:  "name",
			protocol:  "ssh",
			want:      "tenant@tenant.ghe.com:owner/name.git",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := ghRepo{
				hostname: tt.repoHost,
				owner:    tt.repoOwner,
				name:     tt.repoName,
			}
			if url := FormatRemoteURL(r, tt.protocol); url != tt.want {
				t.Errorf("expected url %q, got %q", tt.want, url)
			}
		})
	}
}

const literal_0217 = "monalisa/octo-cat"

const literal_2537 = "github.com"

const literal_1298 = "got error %q"

const literal_5690 = "example.org"

const literal_7103 = "override.com"
