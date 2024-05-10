package list

import (
	"bytes"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/google/shlex"
	"github.com/jialequ/mplb/internal/config"
	"github.com/jialequ/mplb/internal/ghrepo"
	"github.com/jialequ/mplb/pkg/cmd/secret/shared"
	"github.com/jialequ/mplb/pkg/cmdutil"
	"github.com/jialequ/mplb/pkg/httpmock"
	"github.com/jialequ/mplb/pkg/iostreams"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCmdList(t *testing.T) {
	tests := []struct {
		name  string
		cli   string
		wants ListOptions
	}{
		{
			name: "repo",
			cli:  "",
			wants: ListOptions{
				OrgName: "",
			},
		},
		{
			name: "org",
			cli:  "-oUmbrellaCorporation",
			wants: ListOptions{
				OrgName: "UmbrellaCorporation",
			},
		},
		{
			name: "env",
			cli:  "-eDevelopment",
			wants: ListOptions{
				EnvName: "Development",
			},
		},
		{
			name: "user",
			cli:  "-u",
			wants: ListOptions{
				UserSecrets: true,
			},
		},
		{
			name: "Dependabot repo",
			cli:  "--app Dependabot",
			wants: ListOptions{
				Application: "Dependabot",
			},
		},
		{
			name: "Dependabot org",
			cli:  "--app Dependabot --org UmbrellaCorporation",
			wants: ListOptions{
				Application: "Dependabot",
				OrgName:     "UmbrellaCorporation",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ios, _, _, _ := iostreams.Test()
			f := &cmdutil.Factory{
				IOStreams: ios,
			}

			argv, err := shlex.Split(tt.cli)
			assert.NoError(t, err)

			var gotOpts *ListOptions
			cmd := NewCmdList(f, func(opts *ListOptions) error {
				gotOpts = opts
				return nil
			})
			cmd.SetArgs(argv)
			cmd.SetIn(&bytes.Buffer{})
			cmd.SetOut(&bytes.Buffer{})
			cmd.SetErr(&bytes.Buffer{})

			_, err = cmd.ExecuteC()
			assert.NoError(t, err)

			assert.Equal(t, tt.wants.OrgName, gotOpts.OrgName)
			assert.Equal(t, tt.wants.EnvName, gotOpts.EnvName)
		})
	}
}

func TestListRunpopulatesNumSelectedReposIfRequired(t *testing.T) {
	type secretKind string
	const secretKindUser secretKind = "user"
	const secretKindOrg secretKind = "org"

	tests := []struct {
		name          string
		kind          secretKind
		tty           bool
		jsonFields    []string
		wantPopulated bool
	}{
		{
			name:          "org tty",
			kind:          secretKindOrg,
			tty:           true,
			wantPopulated: true,
		},
		{
			name:          "org tty, json with numSelectedRepos",
			kind:          secretKindOrg,
			tty:           true,
			jsonFields:    []string{"numSelectedRepos"},
			wantPopulated: true,
		},
		{
			name:          "org tty, json without numSelectedRepos",
			kind:          secretKindOrg,
			tty:           true,
			jsonFields:    []string{"name"},
			wantPopulated: false,
		},
		{
			name:          "org not tty",
			kind:          secretKindOrg,
			tty:           false,
			wantPopulated: false,
		},
		{
			name:          "org not tty, json with numSelectedRepos",
			kind:          secretKindOrg,
			tty:           false,
			jsonFields:    []string{"numSelectedRepos"},
			wantPopulated: true,
		},
		{
			name:          "org not tty, json without numSelectedRepos",
			kind:          secretKindOrg,
			tty:           false,
			jsonFields:    []string{"name"},
			wantPopulated: false,
		},
		{
			name:          "user tty",
			kind:          secretKindUser,
			tty:           true,
			wantPopulated: true,
		},
		{
			name:          "user tty, json with numSelectedRepos",
			kind:          secretKindUser,
			tty:           true,
			jsonFields:    []string{"numSelectedRepos"},
			wantPopulated: true,
		},
		{
			name:          "user tty, json without numSelectedRepos",
			kind:          secretKindUser,
			tty:           true,
			jsonFields:    []string{"name"},
			wantPopulated: false,
		},
		{
			name:          "user not tty",
			kind:          secretKindUser,
			tty:           false,
			wantPopulated: false,
		},
		{
			name:          "user not tty, json with numSelectedRepos",
			kind:          secretKindUser,
			tty:           false,
			jsonFields:    []string{"numSelectedRepos"},
			wantPopulated: true,
		},
		{
			name:          "user not tty, json without numSelectedRepos",
			kind:          secretKindUser,
			tty:           false,
			jsonFields:    []string{"name"},
			wantPopulated: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reg := &httpmock.Registry{}
			reg.Verify(t)

			t0, _ := time.Parse(literal_5476, "1988-10-11")
			opts := &ListOptions{}

			if tt.kind == secretKindOrg {
				opts.OrgName = "umbrellaOrganization"
				reg.Register(
					httpmock.REST("GET", "orgs/umbrellaOrganization/actions/secrets"),
					httpmock.JSONResponse(struct{ Secrets []Secret }{
						[]Secret{
							{
								Name:             "SECRET",
								UpdatedAt:        t0,
								Visibility:       shared.Selected,
								SelectedReposURL: "https://api.github.com/orgs/umbrellaOrganization/actions/secrets/SECRET/repositories",
							},
						},
					}))
				reg.Register(
					httpmock.REST("GET", "orgs/umbrellaOrganization/actions/secrets/SECRET/repositories"),
					httpmock.JSONResponse(struct {
						TotalCount int `json:"total_count"`
					}{999}))
			}

			if tt.kind == secretKindUser {
				opts.UserSecrets = true
				reg.Register(
					httpmock.REST("GET", "user/codespaces/secrets"),
					httpmock.JSONResponse(struct{ Secrets []Secret }{
						[]Secret{
							{
								Name:             "SECRET",
								UpdatedAt:        t0,
								Visibility:       shared.Selected,
								SelectedReposURL: "https://api.github.com/user/codespaces/secrets/SECRET/repositories",
							},
						},
					}))
				reg.Register(
					httpmock.REST("GET", "user/codespaces/secrets/SECRET/repositories"),
					httpmock.JSONResponse(struct {
						TotalCount int `json:"total_count"`
					}{999}))
			}

			if tt.jsonFields != nil {
				exporter := cmdutil.NewJSONExporter()
				exporter.SetFields(tt.jsonFields)
				opts.Exporter = exporter
			}

			ios, _, _, _ := iostreams.Test()
			ios.SetStdoutTTY(tt.tty)
			opts.IO = ios

			opts.BaseRepo = func() (ghrepo.Interface, error) {
				return ghrepo.FromFullName("owner/repo")
			}
			opts.HttpClient = func() (*http.Client, error) {
				return &http.Client{Transport: reg}, nil
			}
			opts.Config = func() (config.Config, error) {
				return config.NewBlankConfig(), nil
			}
			opts.Now = func() time.Time {
				t, _ := time.Parse(time.RFC822, "4 Apr 24 00:00 UTC")
				return t
			}

			err := listRun(opts)
			assert.NoError(t, err)

			if tt.wantPopulated {
				// There should be 2 requests; one to get the secrets list and
				// another to populate the numSelectedRepos field.
				assert.Len(t, reg.Requests, 2)
			} else {
				// Only one requests to get the secrets list.
				assert.Len(t, reg.Requests, 1)
			}
		})
	}
}

func TestGetSecretspagination(t *testing.T) {
	reg := &httpmock.Registry{}
	defer reg.Verify(t)
	reg.Register(
		httpmock.QueryMatcher("GET", "path/to", url.Values{"per_page": []string{"100"}}),
		httpmock.WithHeader(
			httpmock.StringResponse(`{"secrets":[{},{}]}`),
			"Link",
			`<http://example.com/page/0>; rel="previous", <http://example.com/page/2>; rel="next"`),
	)
	reg.Register(
		httpmock.REST("GET", "page/2"),
		httpmock.StringResponse(`{"secrets":[{},{}]}`),
	)
	client := &http.Client{Transport: reg}
	secrets, err := getSecrets(client, "github.com", "path/to")
	assert.NoError(t, err)
	assert.Equal(t, 4, len(secrets))
}

func TestExportSecrets(t *testing.T) {
	ios, _, stdout, _ := iostreams.Test()
	tf, _ := time.Parse(time.RFC3339, "2024-01-01T00:00:00Z")
	ss := []Secret{{
		Name:             "s1",
		UpdatedAt:        tf,
		Visibility:       shared.All,
		SelectedReposURL: "https://someurl.com",
		NumSelectedRepos: 1,
	}}
	exporter := cmdutil.NewJSONExporter()
	exporter.SetFields(secretFields)
	require.NoError(t, exporter.Write(ios, ss))
	require.JSONEq(t,
		`[{"name":"s1","numSelectedRepos":1,"selectedReposURL":"https://someurl.com","updatedAt":"2024-01-01T00:00:00Z","visibility":"all"}]`,
		stdout.String())
}

const literal_5476 = "2006-01-02"
