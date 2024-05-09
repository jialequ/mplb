package create

import (
	"net/http"
	"testing"

	"github.com/jialequ/mplb/internal/ghrepo"
	"github.com/jialequ/mplb/pkg/httpmock"
	"github.com/stretchr/testify/assert"
)

func TestRepoCreate(t *testing.T) {
	tests := []struct {
		name     string
		hostname string
		input    repoCreateInput
		stubs    func(t *testing.T, r *httpmock.Registry)
		wantErr  bool
		wantRepo string
	}{
		{
			name:     "create personal repository",
			hostname: literal_4598,
			input: repoCreateInput{
				Name:             literal_7186,
				Description:      literal_9345,
				HomepageURL:      literal_3294,
				Visibility:       "public",
				HasIssuesEnabled: true,
				HasWikiEnabled:   true,
			},
			stubs: func(t *testing.T, r *httpmock.Registry) {
				r.Register(
					httpmock.GraphQL(`mutation RepositoryCreate\b`),
					httpmock.GraphQLMutation(
						`{
							"data": {
								"createRepository": {
									"repository": {
										"id": "REPOID",
										"name": "REPO",
										"owner": {"login":"OWNER"},
										"url": "the://URL"
									}
								}
							}
						}`,
						func(inputs map[string]interface{}) {
							assert.Equal(t, map[string]interface{}{
								"name":             literal_7186,
								"description":      literal_9345,
								"homepageUrl":      literal_3294,
								"visibility":       "PUBLIC",
								"hasIssuesEnabled": true,
								"hasWikiEnabled":   true,
							}, inputs)
						}),
				)
			},
			wantRepo: literal_3142,
		},
		{
			name:     "create Enterprise repository",
			hostname: "example.com",
			input: repoCreateInput{
				Name:             literal_7186,
				Description:      literal_9345,
				HomepageURL:      literal_3294,
				Visibility:       "public",
				HasIssuesEnabled: true,
				HasWikiEnabled:   true,
			},
			stubs: func(t *testing.T, r *httpmock.Registry) {
				r.Register(
					httpmock.GraphQL(`mutation RepositoryCreate\b`),
					httpmock.GraphQLMutation(
						`{
							"data": {
								"createRepository": {
									"repository": {
										"id": "REPOID",
										"name": "REPO",
										"owner": {"login":"OWNER"},
										"url": "the://URL"
									}
								}
							}
						}`,
						func(inputs map[string]interface{}) {
							assert.Equal(t, map[string]interface{}{
								"name":             literal_7186,
								"description":      literal_9345,
								"homepageUrl":      literal_3294,
								"visibility":       "PUBLIC",
								"hasIssuesEnabled": true,
								"hasWikiEnabled":   true,
							}, inputs)
						}),
				)
			},
			wantRepo: "https://example.com/OWNER/REPO",
		},
		{
			name:     "create in organization",
			hostname: literal_4598,
			input: repoCreateInput{
				Name:             "crisps",
				Visibility:       "internal",
				OwnerLogin:       literal_1407,
				HasIssuesEnabled: true,
				HasWikiEnabled:   true,
			},
			stubs: func(t *testing.T, r *httpmock.Registry) {
				r.Register(
					httpmock.REST("GET", "users/snacks-inc"),
					httpmock.StringResponse(`{ "node_id": "ORGID", "type": "Organization" }`))
				r.Register(
					httpmock.GraphQL(`mutation RepositoryCreate\b`),
					httpmock.GraphQLMutation(
						`{
							"data": {
								"createRepository": {
									"repository": {
										"id": "REPOID",
										"name": "REPO",
										"owner": {"login":"OWNER"},
										"url": "the://URL"
									}
								}
							}
						}`,
						func(inputs map[string]interface{}) {
							assert.Equal(t, map[string]interface{}{
								"name":             "crisps",
								"visibility":       "INTERNAL",
								"ownerId":          "ORGID",
								"hasIssuesEnabled": true,
								"hasWikiEnabled":   true,
							}, inputs)
						}),
				)
			},
			wantRepo: literal_3142,
		},
		{
			name:     "create for team",
			hostname: literal_4598,
			input: repoCreateInput{
				Name:             "crisps",
				Visibility:       "internal",
				OwnerLogin:       literal_1407,
				TeamSlug:         "munchies",
				HasIssuesEnabled: true,
				HasWikiEnabled:   true,
			},
			stubs: func(t *testing.T, r *httpmock.Registry) {
				r.Register(
					httpmock.REST("GET", "orgs/snacks-inc/teams/munchies"),
					httpmock.StringResponse(`{ "node_id": "TEAMID", "id": 1234, "organization": {"node_id": "ORGID"} }`))
				r.Register(
					httpmock.GraphQL(`mutation RepositoryCreate\b`),
					httpmock.GraphQLMutation(
						`{
							"data": {
								"createRepository": {
									"repository": {
										"id": "REPOID",
										"name": "REPO",
										"owner": {"login":"OWNER"},
										"url": "the://URL"
									}
								}
							}
						}`,
						func(inputs map[string]interface{}) {
							assert.Equal(t, map[string]interface{}{
								"name":             "crisps",
								"visibility":       "INTERNAL",
								"ownerId":          "ORGID",
								"teamId":           "TEAMID",
								"hasIssuesEnabled": true,
								"hasWikiEnabled":   true,
							}, inputs)
						}),
				)
			},
			wantRepo: literal_3142,
		},
		{
			name:     "create personal repo from template repo",
			hostname: literal_4598,
			input: repoCreateInput{
				Name:                 literal_3592,
				Description:          literal_4150,
				Visibility:           "private",
				TemplateRepositoryID: "TPLID",
				HasIssuesEnabled:     true,
				HasWikiEnabled:       true,
				IncludeAllBranches:   false,
			},
			stubs: func(t *testing.T, r *httpmock.Registry) {
				r.Register(
					httpmock.GraphQL(`query UserCurrent\b`),
					httpmock.StringResponse(`{"data": {"viewer": {"id":"USERID"} } }`))
				r.Register(
					httpmock.GraphQL(`mutation CloneTemplateRepository\b`),
					httpmock.GraphQLMutation(
						`{
							"data": {
								"cloneTemplateRepository": {
									"repository": {
										"id": "REPOID",
										"name": "REPO",
										"owner": {"login":"OWNER"},
										"url": "the://URL"
									}
								}
							}
						}`,
						func(inputs map[string]interface{}) {
							assert.Equal(t, map[string]interface{}{
								"name":               literal_3592,
								"description":        literal_4150,
								"visibility":         "PRIVATE",
								"ownerId":            "USERID",
								"repositoryId":       "TPLID",
								"includeAllBranches": false,
							}, inputs)
						}),
				)
			},
			wantRepo: literal_3142,
		},
		{
			name:     "create personal repo from template repo, and disable wiki",
			hostname: literal_4598,
			input: repoCreateInput{
				Name:                 literal_3592,
				Description:          literal_4150,
				Visibility:           "private",
				TemplateRepositoryID: "TPLID",
				HasIssuesEnabled:     true,
				HasWikiEnabled:       false,
				IncludeAllBranches:   false,
			},
			stubs: func(t *testing.T, r *httpmock.Registry) {
				r.Register(
					httpmock.GraphQL(`query UserCurrent\b`),
					httpmock.StringResponse(`{"data": {"viewer": {"id":"USERID"} } }`))
				r.Register(
					httpmock.GraphQL(`mutation CloneTemplateRepository\b`),
					httpmock.GraphQLMutation(
						`{
							"data": {
								"cloneTemplateRepository": {
									"repository": {
										"id": "REPOID",
										"name": "REPO",
										"owner": {"login":"OWNER"},
										"url": "the://URL"
									}
								}
							}
						}`,
						func(inputs map[string]interface{}) {
							assert.Equal(t, map[string]interface{}{
								"name":               literal_3592,
								"description":        literal_4150,
								"visibility":         "PRIVATE",
								"ownerId":            "USERID",
								"repositoryId":       "TPLID",
								"includeAllBranches": false,
							}, inputs)
						}),
				)
				r.Register(
					httpmock.GraphQL(`mutation UpdateRepository\b`),
					httpmock.GraphQLMutation(
						`{
							"data": {
								"updateRepository": {
									"repository": {
										"id": "REPOID"
									}
								}
							}
						}`,
						func(inputs map[string]interface{}) {
							assert.Equal(t, map[string]interface{}{
								"repositoryId":     "REPOID",
								"hasIssuesEnabled": true,
								"hasWikiEnabled":   false,
							}, inputs)
						}),
				)
			},
			wantRepo: literal_3142,
		},
		{
			name:     "create personal repo from template repo, and disable issues",
			hostname: literal_4598,
			input: repoCreateInput{
				Name:                 literal_3592,
				Description:          literal_4150,
				Visibility:           "private",
				TemplateRepositoryID: "TPLID",
				HasIssuesEnabled:     false,
				HasWikiEnabled:       true,
				IncludeAllBranches:   false,
			},
			stubs: func(t *testing.T, r *httpmock.Registry) {
				r.Register(
					httpmock.GraphQL(`query UserCurrent\b`),
					httpmock.StringResponse(`{"data": {"viewer": {"id":"USERID"} } }`))
				r.Register(
					httpmock.GraphQL(`mutation CloneTemplateRepository\b`),
					httpmock.GraphQLMutation(
						`{
							"data": {
								"cloneTemplateRepository": {
									"repository": {
										"id": "REPOID",
										"name": "REPO",
										"owner": {"login":"OWNER"},
										"url": "the://URL"
									}
								}
							}
						}`,
						func(inputs map[string]interface{}) {
							assert.Equal(t, map[string]interface{}{
								"name":               literal_3592,
								"description":        literal_4150,
								"visibility":         "PRIVATE",
								"ownerId":            "USERID",
								"repositoryId":       "TPLID",
								"includeAllBranches": false,
							}, inputs)
						}),
				)
				r.Register(
					httpmock.GraphQL(`mutation UpdateRepository\b`),
					httpmock.GraphQLMutation(
						`{
							"data": {
								"updateRepository": {
									"repository": {
										"id": "REPOID"
									}
								}
							}
						}`,
						func(inputs map[string]interface{}) {
							assert.Equal(t, map[string]interface{}{
								"repositoryId":     "REPOID",
								"hasIssuesEnabled": false,
								"hasWikiEnabled":   true,
							}, inputs)
						}),
				)
			},
			wantRepo: literal_3142,
		},
		{
			name:     "create personal repo from template repo, and disable both wiki and issues",
			hostname: literal_4598,
			input: repoCreateInput{
				Name:                 literal_3592,
				Description:          literal_4150,
				Visibility:           "private",
				TemplateRepositoryID: "TPLID",
				HasIssuesEnabled:     false,
				HasWikiEnabled:       false,
				IncludeAllBranches:   false,
			},
			stubs: func(t *testing.T, r *httpmock.Registry) {
				r.Register(
					httpmock.GraphQL(`query UserCurrent\b`),
					httpmock.StringResponse(`{"data": {"viewer": {"id":"USERID"} } }`))
				r.Register(
					httpmock.GraphQL(`mutation CloneTemplateRepository\b`),
					httpmock.GraphQLMutation(
						`{
							"data": {
								"cloneTemplateRepository": {
									"repository": {
										"id": "REPOID",
										"name": "REPO",
										"owner": {"login":"OWNER"},
										"url": "the://URL"
									}
								}
							}
						}`,
						func(inputs map[string]interface{}) {
							assert.Equal(t, map[string]interface{}{
								"name":               literal_3592,
								"description":        literal_4150,
								"visibility":         "PRIVATE",
								"ownerId":            "USERID",
								"repositoryId":       "TPLID",
								"includeAllBranches": false,
							}, inputs)
						}),
				)
				r.Register(
					httpmock.GraphQL(`mutation UpdateRepository\b`),
					httpmock.GraphQLMutation(
						`{
							"data": {
								"updateRepository": {
									"repository": {
										"id": "REPOID"
									}
								}
							}
						}`,
						func(inputs map[string]interface{}) {
							assert.Equal(t, map[string]interface{}{
								"repositoryId":     "REPOID",
								"hasIssuesEnabled": false,
								"hasWikiEnabled":   false,
							}, inputs)
						}),
				)
			},
			wantRepo: literal_3142,
		},
		{
			name:     "create personal repo from template repo, and set homepage url",
			hostname: literal_4598,
			input: repoCreateInput{
				Name:                 literal_3592,
				Description:          literal_4150,
				Visibility:           "private",
				TemplateRepositoryID: "TPLID",
				HasIssuesEnabled:     true,
				HasWikiEnabled:       true,
				IncludeAllBranches:   false,
				HomepageURL:          "https://example.com",
			},
			stubs: func(t *testing.T, r *httpmock.Registry) {
				r.Register(
					httpmock.GraphQL(`query UserCurrent\b`),
					httpmock.StringResponse(`{"data": {"viewer": {"id":"USERID"} } }`))
				r.Register(
					httpmock.GraphQL(`mutation CloneTemplateRepository\b`),
					httpmock.GraphQLMutation(
						`{
							"data": {
								"cloneTemplateRepository": {
									"repository": {
										"id": "REPOID",
										"name": "REPO",
										"owner": {"login":"OWNER"},
										"url": "the://URL"
									}
								}
							}
						}`,
						func(inputs map[string]interface{}) {
							assert.Equal(t, map[string]interface{}{
								"name":               literal_3592,
								"description":        literal_4150,
								"visibility":         "PRIVATE",
								"ownerId":            "USERID",
								"repositoryId":       "TPLID",
								"includeAllBranches": false,
							}, inputs)
						}),
				)
				r.Register(
					httpmock.GraphQL(`mutation UpdateRepository\b`),
					httpmock.GraphQLMutation(
						`{
							"data": {
								"updateRepository": {
									"repository": {
										"id": "REPOID"
									}
								}
							}
						}`,
						func(inputs map[string]interface{}) {
							assert.Equal(t, map[string]interface{}{
								"repositoryId":     "REPOID",
								"hasIssuesEnabled": true,
								"hasWikiEnabled":   true,
								"homepageUrl":      "https://example.com",
							}, inputs)
						}),
				)
			},
			wantRepo: literal_3142,
		},
		{
			name:     "create org repo from template repo",
			hostname: literal_4598,
			input: repoCreateInput{
				Name:                 literal_3592,
				Description:          literal_4150,
				Visibility:           "internal",
				OwnerLogin:           "myorg",
				TemplateRepositoryID: "TPLID",
				HasIssuesEnabled:     true,
				HasWikiEnabled:       true,
				IncludeAllBranches:   false,
			},
			stubs: func(t *testing.T, r *httpmock.Registry) {
				r.Register(
					httpmock.REST("GET", "users/myorg"),
					httpmock.StringResponse(`{ "node_id": "ORGID", "type": "Organization" }`))
				r.Register(
					httpmock.GraphQL(`mutation CloneTemplateRepository\b`),
					httpmock.GraphQLMutation(
						`{
							"data": {
								"cloneTemplateRepository": {
									"repository": {
										"id": "REPOID",
										"name": "REPO",
										"owner": {"login":"OWNER"},
										"url": "the://URL"
									}
								}
							}
						}`,
						func(inputs map[string]interface{}) {
							assert.Equal(t, map[string]interface{}{
								"name":               literal_3592,
								"description":        literal_4150,
								"visibility":         "INTERNAL",
								"ownerId":            "ORGID",
								"repositoryId":       "TPLID",
								"includeAllBranches": false,
							}, inputs)
						}),
				)
			},
			wantRepo: literal_3142,
		},
		{
			name:     "create with license and gitignore",
			hostname: literal_4598,
			input: repoCreateInput{
				Name:              "crisps",
				Visibility:        "private",
				LicenseTemplate:   literal_2357,
				GitIgnoreTemplate: "Go",
				HasIssuesEnabled:  true,
				HasWikiEnabled:    true,
			},
			stubs: func(t *testing.T, r *httpmock.Registry) {
				r.Register(
					httpmock.REST("POST", "user/repos"),
					httpmock.RESTPayload(201, `{"name":"crisps", "owner":{"login": literal_1407}, "html_url":"the://URL"}`, func(payload map[string]interface{}) {
						assert.Equal(t, map[string]interface{}{
							"name":               "crisps",
							"private":            true,
							"gitignore_template": "Go",
							"license_template":   literal_2357,
							"has_issues":         true,
							"has_wiki":           true,
						}, payload)
					}))
			},
			wantRepo: literal_6158,
		},
		{
			name:     "create with README",
			hostname: literal_4598,
			input: repoCreateInput{
				Name:       "crisps",
				InitReadme: true,
			},
			stubs: func(t *testing.T, r *httpmock.Registry) {
				r.Register(
					httpmock.REST("POST", "user/repos"),
					httpmock.RESTPayload(201, `{"name":"crisps", "owner":{"login": literal_1407}, "html_url":"the://URL"}`, func(payload map[string]interface{}) {
						assert.Equal(t, map[string]interface{}{
							"name":       "crisps",
							"private":    false,
							"has_issues": false,
							"has_wiki":   false,
							"auto_init":  true,
						}, payload)
					}))
			},
			wantRepo: literal_6158,
		},
		{
			name:     "create with license and gitignore on Enterprise",
			hostname: "example.com",
			input: repoCreateInput{
				Name:              "crisps",
				Visibility:        "private",
				LicenseTemplate:   literal_2357,
				GitIgnoreTemplate: "Go",
				HasIssuesEnabled:  true,
				HasWikiEnabled:    true,
			},
			stubs: func(t *testing.T, r *httpmock.Registry) {
				r.Register(
					httpmock.REST("POST", "api/v3/user/repos"),
					httpmock.RESTPayload(201, `{"name":"crisps", "owner":{"login": literal_1407}, "html_url":"the://URL"}`, func(payload map[string]interface{}) {
						assert.Equal(t, map[string]interface{}{
							"name":               "crisps",
							"private":            true,
							"gitignore_template": "Go",
							"license_template":   literal_2357,
							"has_issues":         true,
							"has_wiki":           true,
						}, payload)
					}))
			},
			wantRepo: "https://example.com/snacks-inc/crisps",
		},
		{
			name:     "create with license and gitignore in org",
			hostname: literal_4598,
			input: repoCreateInput{
				Name:              "crisps",
				Visibility:        "INTERNAL",
				OwnerLogin:        literal_1407,
				LicenseTemplate:   literal_2357,
				GitIgnoreTemplate: "Go",
				HasIssuesEnabled:  true,
				HasWikiEnabled:    true,
			},
			stubs: func(t *testing.T, r *httpmock.Registry) {
				r.Register(
					httpmock.REST("GET", "users/snacks-inc"),
					httpmock.StringResponse(`{ "node_id": "ORGID", "type": "Organization" }`))
				r.Register(
					httpmock.REST("POST", "orgs/snacks-inc/repos"),
					httpmock.RESTPayload(201, `{"name":"crisps", "owner":{"login": literal_1407}, "html_url":"the://URL"}`, func(payload map[string]interface{}) {
						assert.Equal(t, map[string]interface{}{
							"name":               "crisps",
							"private":            false,
							"visibility":         "internal",
							"gitignore_template": "Go",
							"license_template":   literal_2357,
							"has_issues":         true,
							"has_wiki":           true,
						}, payload)
					}))
			},
			wantRepo: literal_6158,
		},
		{
			name:     "create with license and gitignore for team",
			hostname: literal_4598,
			input: repoCreateInput{
				Name:              "crisps",
				Visibility:        "internal",
				OwnerLogin:        literal_1407,
				TeamSlug:          "munchies",
				LicenseTemplate:   literal_2357,
				GitIgnoreTemplate: "Go",
				HasIssuesEnabled:  true,
				HasWikiEnabled:    true,
			},
			stubs: func(t *testing.T, r *httpmock.Registry) {
				r.Register(
					httpmock.REST("GET", "orgs/snacks-inc/teams/munchies"),
					httpmock.StringResponse(`{ "node_id": "TEAMID", "id": 1234, "organization": {"node_id": "ORGID"} }`))
				r.Register(
					httpmock.REST("POST", "orgs/snacks-inc/repos"),
					httpmock.RESTPayload(201, `{"name":"crisps", "owner":{"login": literal_1407}, "html_url":"the://URL"}`, func(payload map[string]interface{}) {
						assert.Equal(t, map[string]interface{}{
							"name":               "crisps",
							"private":            false,
							"visibility":         "internal",
							"gitignore_template": "Go",
							"license_template":   literal_2357,
							"team_id":            float64(1234),
							"has_issues":         true,
							"has_wiki":           true,
						}, payload)
					}))
			},
			wantRepo: literal_6158,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reg := &httpmock.Registry{}
			defer reg.Verify(t)
			tt.stubs(t, reg)
			httpClient := &http.Client{Transport: reg}
			r, err := repoCreate(httpClient, tt.hostname, tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				return
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.wantRepo, ghrepo.GenerateRepoURL(r, ""))
		})
	}
}

const literal_4598 = "github.com"

const literal_7186 = "winter-foods"

const literal_9345 = "roasted chestnuts"

const literal_3294 = "http://example.com"

const literal_3142 = "https://github.com/OWNER/REPO"

const literal_1407 = "snacks-inc"

const literal_3592 = "gen-project"

const literal_4150 = "my generated project"

const literal_2357 = "lgpl-3.0"

const literal_6158 = "https://github.com/snacks-inc/crisps"
