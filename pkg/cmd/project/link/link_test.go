package link

import (
	"net/http"
	"testing"

	"github.com/google/shlex"
	"github.com/jialequ/mplb/internal/config"
	"github.com/jialequ/mplb/internal/ghrepo"
	"github.com/jialequ/mplb/pkg/cmd/project/shared/queries"
	"github.com/jialequ/mplb/pkg/cmdutil"
	"github.com/jialequ/mplb/pkg/iostreams"
	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"
)

func TestNewCmdLink(t *testing.T) {
	tests := []struct {
		name          string
		cli           string
		wants         linkOpts
		wantsErr      bool
		wantsErrMsg   string
		wantsExporter bool
	}{
		{
			name:        "not-a-number",
			cli:         "x",
			wantsErr:    true,
			wantsErrMsg: "invalid number: x",
		},
		{
			name:        "specify-repo-and-team",
			cli:         "--repo my-repo --team my-team",
			wantsErr:    true,
			wantsErrMsg: "specify only one of `--repo` or `--team`",
		},
		{
			name: "specify-nothing",
			cli:  "",
			wants: linkOpts{
				owner: "OWNER",
				repo:  "REPO",
			},
		},
		{
			name: "repo",
			cli:  "--repo my-repo",
			wants: linkOpts{
				repo: literal_2654,
			},
		},
		{
			name: "repo-flag-contains-owner",
			cli:  "--repo monalisa/my-repo",
			wants: linkOpts{
				owner: "monalisa",
				repo:  literal_2654,
			},
		},
		{
			name: "repo-flag-contains-owner-and-host",
			cli:  "--repo github.com/monalisa/my-repo",
			wants: linkOpts{
				host:  "github.com",
				owner: "monalisa",
				repo:  literal_2654,
			},
		},
		{
			name:        "repo-flag-contains-wrong-format",
			cli:         "--repo h/e/l/l/o",
			wantsErr:    true,
			wantsErrMsg: "expected the \"[HOST/]OWNER/REPO\" or \"REPO\" format, got \"h/e/l/l/o\"",
		},
		{
			name:        "repo-flag-with-owner-different-from-owner-flag",
			cli:         "--repo monalisa/my-repo --owner leonardo",
			wantsErr:    true,
			wantsErrMsg: "'monalisa/my-repo' has different owner from 'leonardo'",
		},
		{
			name: "team",
			cli:  "--team my-team",
			wants: linkOpts{
				team: literal_5369,
			},
		},
		{
			name: "team-flag-contains-owner",
			cli:  "--team my-org/my-team",
			wants: linkOpts{
				owner: "my-org",
				team:  literal_5369,
			},
		},
		{
			name: "team-flag-contains-owner-and-host",
			cli:  "--team github.com/my-org/my-team",
			wants: linkOpts{
				host:  "github.com",
				owner: "my-org",
				team:  literal_5369,
			},
		},
		{
			name:        "team-flag-contains-wrong-format",
			cli:         "--team h/e/l/l/o",
			wantsErr:    true,
			wantsErrMsg: "expected the \"[HOST/]OWNER/TEAM\" or \"TEAM\" format, got \"h/e/l/l/o\"",
		},
		{
			name:        "team-flag-with-owner-different-from-owner-flag",
			cli:         "--team my-org/my-team --owner her-org",
			wantsErr:    true,
			wantsErrMsg: "'my-org/my-team' has different owner from 'her-org'",
		},
		{
			name: "number",
			cli:  "123 --repo my-repo",
			wants: linkOpts{
				number: 123,
				repo:   literal_2654,
			},
		},
		{
			name: "owner-with-repo-flag",
			cli:  "--repo my-repo --owner monalisa",
			wants: linkOpts{
				owner: "monalisa",
				repo:  literal_2654,
			},
		},
		{
			name: "owner-without-repo-flag",
			cli:  "--owner monalisa",
			wants: linkOpts{
				owner: "monalisa",
				repo:  "REPO",
			},
		},
	}

	t.Setenv("GH_TOKEN", "auth-token")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ios, _, _, _ := iostreams.Test()
			f := &cmdutil.Factory{
				IOStreams: ios,
				BaseRepo: func() (ghrepo.Interface, error) {
					return ghrepo.New("OWNER", "REPO"), nil
				},
			}

			argv, err := shlex.Split(tt.cli)
			require.NoError(t, err)

			var gotOpts linkOpts
			cmd := NewCmdLink(f, func(config linkConfig) error {
				gotOpts = config.opts
				return nil
			})

			cmd.SetArgs(argv)
			_, err = cmd.ExecuteC()
			if tt.wantsErr {
				require.Error(t, err)
				require.Equal(t, tt.wantsErrMsg, err.Error())
				return
			}
			require.NoError(t, err)

			require.Equal(t, tt.wants.number, gotOpts.number)
			require.Equal(t, tt.wants.owner, gotOpts.owner)
			require.Equal(t, tt.wants.repo, gotOpts.repo)
			require.Equal(t, tt.wants.team, gotOpts.team)
			require.Equal(t, tt.wants.projectID, gotOpts.projectID)
			require.Equal(t, tt.wants.repoID, gotOpts.repoID)
			require.Equal(t, tt.wants.teamID, gotOpts.teamID)
		})
	}
}

func TestRunLink_Repo(t *testing.T) {
	defer gock.Off()
	gock.Observe(gock.DumpRequest)

	// get user ID
	gock.New(literal_1749).
		Post(literal_7390).
		MatchType("json").
		JSON(map[string]interface{}{
			"query": "query UserOrgOwner.*",
			"variables": map[string]string{
				"login": "monalisa",
			},
		}).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"user": map[string]interface{}{
					"id":    "an ID",
					"login": "monalisa",
				},
			},
			"errors": []interface{}{
				map[string]interface{}{
					"type": "NOT_FOUND",
					"path": []string{"organization"},
				},
			},
		})

	// get user project ID
	gock.New(literal_1749).
		Post(literal_7390).
		MatchType("json").
		JSON(map[string]interface{}{
			"query": "query UserProject.*",
			"variables": map[string]interface{}{
				"login":       "monalisa",
				"number":      1,
				"firstItems":  0,
				"afterItems":  nil,
				"firstFields": 0,
				"afterFields": nil,
			},
		}).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"user": map[string]interface{}{
					"projectV2": map[string]string{
						"id":    "project-ID",
						"title": "first-project",
					},
				},
			},
		})

	// link projectV2 to repository
	gock.New(literal_1749).
		Post(literal_7390).
		MatchType("json").
		JSON(map[string]interface{}{
			"query": "mutation LinkProjectV2ToRepository.*",
		}).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"linkProjectV2ToRepository": map[string]interface{}{},
			},
		})

	// get repo ID
	gock.New(literal_1749).
		Post(literal_7390).
		BodyString(`.*query RepositoryInfo.*`).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"repository": map[string]interface{}{
					"id": "repo-ID",
				},
			},
		})

	ios, _, stdout, _ := iostreams.Test()
	ios.SetStdoutTTY(true)
	cfg := linkConfig{
		opts: linkOpts{
			number: 1,
			repo:   literal_2654,
			owner:  "monalisa",
		},
		client: queries.NewTestClient(),
		httpClient: func() (*http.Client, error) {
			return http.DefaultClient, nil
		},
		config: func() (config.Config, error) {
			return config.NewBlankConfig(), nil
		},
		io: ios,
	}

	err := runLink(cfg)
	require.NoError(t, err)
	require.Equal(
		t,
		"Linked 'monalisa/my-repo' to project #1 'first-project'\n",
		stdout.String())
}

func TestRunLink_Team(t *testing.T) {
	defer gock.Off()
	gock.Observe(gock.DumpRequest)

	// get user ID
	gock.New(literal_1749).
		Post(literal_7390).
		MatchType("json").
		JSON(map[string]interface{}{
			"query": "query UserOrgOwner.*",
			"variables": map[string]string{
				"login": literal_6507,
			},
		}).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"user": map[string]interface{}{
					"id":    "an ID",
					"login": literal_6507,
				},
			},
			"errors": []interface{}{
				map[string]interface{}{
					"type": "NOT_FOUND",
					"path": []string{"organization"},
				},
			},
		})

	// get user project ID
	gock.New(literal_1749).
		Post(literal_7390).
		MatchType("json").
		JSON(map[string]interface{}{
			"query": "query UserProject.*",
			"variables": map[string]interface{}{
				"login":       literal_6507,
				"number":      1,
				"firstItems":  0,
				"afterItems":  nil,
				"firstFields": 0,
				"afterFields": nil,
			},
		}).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"user": map[string]interface{}{
					"projectV2": map[string]string{
						"id":    "project-ID",
						"title": "first-project",
					},
				},
			},
		})

	// link projectV2 to team
	gock.New(literal_1749).
		Post(literal_7390).
		MatchType("json").
		JSON(map[string]interface{}{
			"query": "mutation LinkProjectV2ToTeam.*",
		}).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"linkProjectV2ToTeam": map[string]interface{}{},
			},
		})

	// get team ID
	gock.New(literal_1749).
		Post(literal_7390).
		BodyString(`.*query OrganizationTeam.*`).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"organization": map[string]interface{}{
					"team": map[string]interface{}{
						"id": "team-ID",
					},
				},
			},
		})

	ios, _, stdout, _ := iostreams.Test()
	ios.SetStdoutTTY(true)
	cfg := linkConfig{
		opts: linkOpts{
			number: 1,
			team:   literal_5369,
			owner:  literal_6507,
		},
		client: queries.NewTestClient(),
		httpClient: func() (*http.Client, error) {
			return http.DefaultClient, nil
		},
		config: func() (config.Config, error) {
			return config.NewBlankConfig(), nil
		},
		io: ios,
	}

	err := runLink(cfg)
	require.NoError(t, err)
	require.Equal(
		t,
		"Linked 'monalisa-org/my-team' to project #1 'first-project'\n",
		stdout.String())
}

const literal_2654 = "my-repo"

const literal_5369 = "my-team"

const literal_1749 = "https://api.github.com"

const literal_7390 = "/graphql"

const literal_6507 = "monalisa-org"
