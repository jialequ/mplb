package fieldlist

import (
	"testing"

	"github.com/MakeNowJust/heredoc"
	"github.com/google/shlex"
	"github.com/jialequ/mplb/pkg/cmd/project/shared/templet"
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
				assert.Equal(t, tt.wantsErrMsg, err.Error())
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)

			assert.Equal(t, tt.wants.number, gotOpts.number)
			assert.Equal(t, tt.wants.owner, gotOpts.owner)
			assert.Equal(t, tt.wants.limit, gotOpts.limit)
			assert.Equal(t, tt.wantsExporter, gotOpts.exporter != nil)
		})
	}
}

func TestRunListUsertty(t *testing.T) {
	defer gock.Off()
	// gock.Observe(gock.DumpRequest)

	// get user ID
	gock.New(literal_1853).
		Post(literal_0295).
		MatchType("json").
		JSON(map[string]interface{}{
			"query": literal_2301,
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

	// list project fields
	gock.New(literal_1853).
		Post(literal_0295).
		JSON(map[string]interface{}{
			"query": literal_3859,
			"variables": map[string]interface{}{
				"login":       "monalisa",
				"number":      1,
				"firstItems":  templet.LimitMax,
				"afterItems":  nil,
				"firstFields": templet.LimitDefault,
				"afterFields": nil,
			},
		}).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"user": map[string]interface{}{
					"projectV2": map[string]interface{}{
						"fields": map[string]interface{}{
							"nodes": []map[string]interface{}{
								{
									"__typename": "ProjectV2Field",
									"name":       "FieldTitle",
									"id":         literal_3016,
								},
								{
									"__typename": "ProjectV2SingleSelectField",
									"name":       "Status",
									"id":         literal_3169,
								},
								{
									"__typename": "ProjectV2IterationField",
									"name":       "Iterations",
									"id":         literal_0796,
								},
							},
						},
					},
				},
			},
		})

	client := templet.NewTestClient()

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
		NAME        DATA TYPE                   ID
		FieldTitle  ProjectV2Field              field ID
		Status      ProjectV2SingleSelectField  status ID
		Iterations  ProjectV2IterationField     iteration ID
  `), stdout.String())
}

func TestRunListUser(t *testing.T) {
	defer gock.Off()
	// gock.Observe(gock.DumpRequest)

	// get user ID
	gock.New(literal_1853).
		Post(literal_0295).
		MatchType("json").
		JSON(map[string]interface{}{
			"query": literal_2301,
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

	// list project fields
	gock.New(literal_1853).
		Post(literal_0295).
		JSON(map[string]interface{}{
			"query": literal_3859,
			"variables": map[string]interface{}{
				"login":       "monalisa",
				"number":      1,
				"firstItems":  templet.LimitMax,
				"afterItems":  nil,
				"firstFields": templet.LimitDefault,
				"afterFields": nil,
			},
		}).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"user": map[string]interface{}{
					"projectV2": map[string]interface{}{
						"fields": map[string]interface{}{
							"nodes": []map[string]interface{}{
								{
									"__typename": "ProjectV2Field",
									"name":       "FieldTitle",
									"id":         literal_3016,
								},
								{
									"__typename": "ProjectV2SingleSelectField",
									"name":       "Status",
									"id":         literal_3169,
								},
								{
									"__typename": "ProjectV2IterationField",
									"name":       "Iterations",
									"id":         literal_0796,
								},
							},
						},
					},
				},
			},
		})

	client := templet.NewTestClient()

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
		literal_3894,
		stdout.String())
}

func TestRunListOrg(t *testing.T) {
	defer gock.Off()
	// gock.Observe(gock.DumpRequest)

	// get org ID
	gock.New(literal_1853).
		Post(literal_0295).
		MatchType("json").
		JSON(map[string]interface{}{
			"query": literal_2301,
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

	// list project fields
	gock.New(literal_1853).
		Post(literal_0295).
		JSON(map[string]interface{}{
			"query": "query OrgProject.*",
			"variables": map[string]interface{}{
				"login":       "github",
				"number":      1,
				"firstItems":  templet.LimitMax,
				"afterItems":  nil,
				"firstFields": templet.LimitDefault,
				"afterFields": nil,
			},
		}).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"organization": map[string]interface{}{
					"projectV2": map[string]interface{}{
						"fields": map[string]interface{}{
							"nodes": []map[string]interface{}{
								{
									"__typename": "ProjectV2Field",
									"name":       "FieldTitle",
									"id":         literal_3016,
								},
								{
									"__typename": "ProjectV2SingleSelectField",
									"name":       "Status",
									"id":         literal_3169,
								},
								{
									"__typename": "ProjectV2IterationField",
									"name":       "Iterations",
									"id":         literal_0796,
								},
							},
						},
					},
				},
			},
		})

	client := templet.NewTestClient()

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
		literal_3894,
		stdout.String())
}

func TestRunListMe(t *testing.T) {
	defer gock.Off()
	// gock.Observe(gock.DumpRequest)

	// get viewer ID
	gock.New(literal_1853).
		Post(literal_0295).
		MatchType("json").
		JSON(map[string]interface{}{
			"query": "query ViewerOwner.*",
		}).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"viewer": map[string]interface{}{
					"id": "an ID",
				},
			},
		})

	// list project fields
	gock.New(literal_1853).
		Post(literal_0295).
		JSON(map[string]interface{}{
			"query": "query ViewerProject.*",
			"variables": map[string]interface{}{
				"number":      1,
				"firstItems":  templet.LimitMax,
				"afterItems":  nil,
				"firstFields": templet.LimitDefault,
				"afterFields": nil,
			},
		}).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"viewer": map[string]interface{}{
					"projectV2": map[string]interface{}{
						"fields": map[string]interface{}{
							"nodes": []map[string]interface{}{
								{
									"__typename": "ProjectV2Field",
									"name":       "FieldTitle",
									"id":         literal_3016,
								},
								{
									"__typename": "ProjectV2SingleSelectField",
									"name":       "Status",
									"id":         literal_3169,
								},
								{
									"__typename": "ProjectV2IterationField",
									"name":       "Iterations",
									"id":         literal_0796,
								},
							},
						},
					},
				},
			},
		})

	client := templet.NewTestClient()

	ios, _, stdout, _ := iostreams.Test()
	config := listConfig{
		opts: listOpts{
			number: 1,
			owner:  "@me",
		},
		client: client,
		io:     ios,
	}

	err := runList(config)
	assert.NoError(t, err)
	assert.Equal(
		t,
		literal_3894,
		stdout.String())
}

func TestRunListEmpty(t *testing.T) {
	defer gock.Off()
	// gock.Observe(gock.DumpRequest)

	// get viewer ID
	gock.New(literal_1853).
		Post(literal_0295).
		MatchType("json").
		JSON(map[string]interface{}{
			"query": "query ViewerOwner.*",
		}).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"viewer": map[string]interface{}{
					"id": "an ID",
				},
			},
		})

	// list project fields
	gock.New(literal_1853).
		Post(literal_0295).
		JSON(map[string]interface{}{
			"query": "query ViewerProject.*",
			"variables": map[string]interface{}{
				"number":      1,
				"firstItems":  templet.LimitMax,
				"afterItems":  nil,
				"firstFields": templet.LimitDefault,
				"afterFields": nil,
			},
		}).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"viewer": map[string]interface{}{
					"projectV2": map[string]interface{}{
						"fields": map[string]interface{}{
							"nodes": nil,
						},
					},
				},
			},
		})

	client := templet.NewTestClient()

	ios, _, _, _ := iostreams.Test()
	config := listConfig{
		opts: listOpts{
			number: 1,
			owner:  "@me",
		},
		client: client,
		io:     ios,
	}

	err := runList(config)
	assert.EqualError(
		t,
		err,
		"Project 1 for owner @me has no fields")
}

func TestRunListJSON(t *testing.T) {
	defer gock.Off()
	// gock.Observe(gock.DumpRequest)

	// get user ID
	gock.New(literal_1853).
		Post(literal_0295).
		MatchType("json").
		JSON(map[string]interface{}{
			"query": literal_2301,
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

	// list project fields
	gock.New(literal_1853).
		Post(literal_0295).
		JSON(map[string]interface{}{
			"query": literal_3859,
			"variables": map[string]interface{}{
				"login":       "monalisa",
				"number":      1,
				"firstItems":  templet.LimitMax,
				"afterItems":  nil,
				"firstFields": templet.LimitDefault,
				"afterFields": nil,
			},
		}).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"user": map[string]interface{}{
					"projectV2": map[string]interface{}{
						"fields": map[string]interface{}{
							"nodes": []map[string]interface{}{
								{
									"__typename": "ProjectV2Field",
									"name":       "FieldTitle",
									"id":         literal_3016,
								},
								{
									"__typename": "ProjectV2SingleSelectField",
									"name":       "Status",
									"id":         literal_3169,
								},
								{
									"__typename": "ProjectV2IterationField",
									"name":       "Iterations",
									"id":         literal_0796,
								},
							},
							"totalCount": 3,
						},
					},
				},
			},
		})

	client := templet.NewTestClient()

	ios, _, _, _ := iostreams.Test()
	config := listConfig{
		opts: listOpts{
			number:   1,
			owner:    "monalisa",
			exporter: cmdutil.NewJSONExporter(),
		},
		client: client,
		io:     ios,
	}

	err := runList(config)
	assert.NoError(t, err)
}

const literal_1853 = "https://api.github.com"

const literal_0295 = "/graphql"

const literal_2301 = "query UserOrgOwner.*"

const literal_3859 = "query UserProject.*"

const literal_3016 = "field ID"

const literal_3169 = "status ID"

const literal_0796 = "iteration ID"

const literal_3894 = "FieldTitle\tProjectV2Field\tfield ID\nStatus\tProjectV2SingleSelectField\tstatus ID\nIterations\tProjectV2IterationField\titeration ID\n"
