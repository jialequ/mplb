package itemlist

import (
	"testing"

	"github.com/MakeNowJust/heredoc"
	"github.com/google/shlex"
	"github.com/jialequ/mplb/pkg/cmd/project/shared/queries"
	"github.com/jialequ/mplb/pkg/cmdutil"
	"github.com/jialequ/mplb/pkg/iostreams"

	"github.com/stretchr/testify/assert"
	"gopkg.in/h2non/gock.v1"
)

func TestNewCmdList(t *testing.T) {
	tests := []struct {
		name          string
		cli           string
		wants         listOpts
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
			name: "number",
			cli:  "123",
			wants: listOpts{
				number: 123,
				limit:  30,
			},
		},
		{
			name: "owner",
			cli:  "--owner monalisa",
			wants: listOpts{
				owner: "monalisa",
				limit: 30,
			},
		},
		{
			name: "json",
			cli:  "--format json",
			wants: listOpts{
				limit: 30,
			},
			wantsExporter: true,
		},
	}

	t.Setenv("GH_TOKEN", "auth-token")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ios, _, _, _ := iostreams.Test()
			f := &cmdutil.Factory{
				IOStreams: ios,
			}

			argv, err := shlex.Split(tt.cli)
			assert.NoError(t, err)

			var gotOpts listOpts
			cmd := NewCmdList(f, func(config listConfig) error {
				gotOpts = config.opts
				return nil
			})

			cmd.SetArgs(argv)
			_, err = cmd.ExecuteC()
			if tt.wantsErr {
				assert.Error(t, err)
				assert.Equal(t, tt.wantsErrMsg, err.Error())
				return
			}
			assert.NoError(t, err)

			assert.Equal(t, tt.wants.number, gotOpts.number)
			assert.Equal(t, tt.wants.owner, gotOpts.owner)
			assert.Equal(t, tt.wantsExporter, gotOpts.exporter != nil)
			assert.Equal(t, tt.wants.limit, gotOpts.limit)
		})
	}
}

func TestRunListUsertty(t *testing.T) {
	defer gock.Off()
	// gock.Observe(gock.DumpRequest)

	// get user ID
	gock.New(literal_4380).
		Post(literal_0539).
		MatchType("json").
		JSON(map[string]interface{}{
			"query": literal_9123,
			"variables": map[string]interface{}{
				"login": "monalisa",
			},
		}).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"user": map[string]interface{}{
					"id": "an ID",
				},
			},
			"errors": []interface{}{
				map[string]interface{}{
					"type": "NOT_FOUND",
					"path": []string{"organization"},
				},
			},
		})

	// list project items
	gock.New(literal_4380).
		Post(literal_0539).
		JSON(map[string]interface{}{
			"query": literal_2079,
			"variables": map[string]interface{}{
				"firstItems":  queries.LimitDefault,
				"afterItems":  nil,
				"firstFields": queries.LimitMax,
				"afterFields": nil,
				"login":       "monalisa",
				"number":      1,
			},
		}).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"user": map[string]interface{}{
					"projectV2": map[string]interface{}{
						"items": map[string]interface{}{
							"nodes": []map[string]interface{}{
								{
									"id": literal_5026,
									"content": map[string]interface{}{
										"__typename": "Issue",
										"title":      literal_6781,
										"number":     1,
										"repository": map[string]string{
											"nameWithOwner": literal_4231,
										},
									},
								},
								{
									"id": literal_4327,
									"content": map[string]interface{}{
										"__typename": "PullRequest",
										"title":      literal_9617,
										"number":     2,
										"repository": map[string]string{
											"nameWithOwner": literal_4231,
										},
									},
								},
								{
									"id": literal_3951,
									"content": map[string]interface{}{
										"id":         literal_3951,
										"title":      "draft issue1",
										"__typename": "DraftIssue",
									},
								},
							},
						},
					},
				},
			},
		})

	client := queries.NewTestClient()

	ios, _, stdout, _ := iostreams.Test()
	ios.SetStdoutTTY(true)
	config := listConfig{
		opts: listOpts{
			number: 1,
			owner:  "monalisa",
		},
		client: client,
		io:     ios,
	}

	err := runList(config)
	assert.NoError(t, err)
	assert.Equal(t, heredoc.Doc(`
		TYPE         TITLE           NUMBER  REPOSITORY  ID
		Issue        an issue        1       cli/go-gh   issue ID
		PullRequest  a pull request  2       cli/go-gh   pull request ID
		DraftIssue   draft issue                         draft issue ID
  `), stdout.String())
}

func TestRunListUser(t *testing.T) {
	defer gock.Off()
	// gock.Observe(gock.DumpRequest)

	// get user ID
	gock.New(literal_4380).
		Post(literal_0539).
		MatchType("json").
		JSON(map[string]interface{}{
			"query": literal_9123,
			"variables": map[string]interface{}{
				"login": "monalisa",
			},
		}).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"user": map[string]interface{}{
					"id": "an ID",
				},
			},
			"errors": []interface{}{
				map[string]interface{}{
					"type": "NOT_FOUND",
					"path": []string{"organization"},
				},
			},
		})

	// list project items
	gock.New(literal_4380).
		Post(literal_0539).
		JSON(map[string]interface{}{
			"query": literal_2079,
			"variables": map[string]interface{}{
				"firstItems":  queries.LimitDefault,
				"afterItems":  nil,
				"firstFields": queries.LimitMax,
				"afterFields": nil,
				"login":       "monalisa",
				"number":      1,
			},
		}).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"user": map[string]interface{}{
					"projectV2": map[string]interface{}{
						"items": map[string]interface{}{
							"nodes": []map[string]interface{}{
								{
									"id": literal_5026,
									"content": map[string]interface{}{
										"__typename": "Issue",
										"title":      literal_6781,
										"number":     1,
										"repository": map[string]string{
											"nameWithOwner": literal_4231,
										},
									},
								},
								{
									"id": literal_4327,
									"content": map[string]interface{}{
										"__typename": "PullRequest",
										"title":      literal_9617,
										"number":     2,
										"repository": map[string]string{
											"nameWithOwner": literal_4231,
										},
									},
								},
								{
									"id": literal_3951,
									"content": map[string]interface{}{
										"id":         literal_3951,
										"title":      "draft issue2",
										"__typename": "DraftIssue",
									},
								},
							},
						},
					},
				},
			},
		})

	client := queries.NewTestClient()

	ios, _, stdout, _ := iostreams.Test()
	config := listConfig{
		opts: listOpts{
			number: 1,
			owner:  "monalisa",
		},
		client: client,
		io:     ios,
	}

	err := runList(config)
	assert.NoError(t, err)
	assert.Equal(
		t,
		literal_4513,
		stdout.String())
}

func TestRunListOrg(t *testing.T) {
	defer gock.Off()
	// gock.Observe(gock.DumpRequest)

	// get org ID
	gock.New(literal_4380).
		Post(literal_0539).
		MatchType("json").
		JSON(map[string]interface{}{
			"query": literal_9123,
			"variables": map[string]interface{}{
				"login": "github",
			},
		}).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"organization": map[string]interface{}{
					"id": "an ID",
				},
			},
			"errors": []interface{}{
				map[string]interface{}{
					"type": "NOT_FOUND",
					"path": []string{"user"},
				},
			},
		})

	// list project items
	gock.New(literal_4380).
		Post(literal_0539).
		JSON(map[string]interface{}{
			"query": "query OrgProjectWithItems.*",
			"variables": map[string]interface{}{
				"firstItems":  queries.LimitDefault,
				"afterItems":  nil,
				"firstFields": queries.LimitMax,
				"afterFields": nil,
				"login":       "github",
				"number":      1,
			},
		}).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"organization": map[string]interface{}{
					"projectV2": map[string]interface{}{
						"items": map[string]interface{}{
							"nodes": []map[string]interface{}{
								{
									"id": literal_5026,
									"content": map[string]interface{}{
										"__typename": "Issue",
										"title":      literal_6781,
										"number":     1,
										"repository": map[string]string{
											"nameWithOwner": literal_4231,
										},
									},
								},
								{
									"id": literal_4327,
									"content": map[string]interface{}{
										"__typename": "PullRequest",
										"title":      literal_9617,
										"number":     2,
										"repository": map[string]string{
											"nameWithOwner": literal_4231,
										},
									},
								},
								{
									"id": literal_3951,
									"content": map[string]interface{}{
										"id":         literal_3951,
										"title":      "draft issue3",
										"__typename": "DraftIssue",
									},
								},
							},
						},
					},
				},
			},
		})

	client := queries.NewTestClient()

	ios, _, stdout, _ := iostreams.Test()
	config := listConfig{
		opts: listOpts{
			number: 1,
			owner:  "github",
		},
		client: client,
		io:     ios,
	}

	err := runList(config)
	assert.NoError(t, err)
	assert.Equal(
		t,
		literal_4513,
		stdout.String())
}

const literal_4380 = "https://api.github.com"

const literal_0539 = "/graphql"

const literal_9123 = "query UserOrgOwner.*"

const literal_2079 = "query UserProjectWithItems.*"

const literal_5026 = "issue ID"

const literal_6781 = "an issue"

const literal_4231 = "cli/go-gh"

const literal_4327 = "pull request ID"

const literal_9617 = "a pull request"

const literal_3951 = "draft issue ID"

const literal_4513 = "Issue\tan issue\t1\tcli/go-gh\tissue ID\nPullRequest\ta pull request\t2\tcli/go-gh\tpull request ID\nDraftIssue\tdraft issue\t\t\tdraft issue ID\n"
