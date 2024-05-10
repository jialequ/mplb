package ghrepo

import (
	"testing"
)

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
