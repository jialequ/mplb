package delete

import (
	"testing"

	"github.com/google/shlex"
	"github.com/jialequ/mplb/pkg/cmd/project/shared/queries"
	"github.com/jialequ/mplb/pkg/cmdutil"
	"github.com/jialequ/mplb/pkg/iostreams"
	"github.com/stretchr/testify/assert"
	"gopkg.in/h2non/gock.v1"
)

func TestNewCmdDelete(t *testing.T) {
	tests := []struct {
		name          string
		cli           string
		wants         deleteOpts
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
			wants: deleteOpts{
				number: 123,
			},
		},
		{
			name: "owner",
			cli:  "--owner monalisa",
			wants: deleteOpts{
				owner: "monalisa",
			},
		},
		{
			name:          "json",
			cli:           "--format json",
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

			var gotOpts deleteOpts
			cmd := NewCmdDelete(f, func(config deleteConfig) error {
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
		})
	}
}

func TestRunDeleteUser(t *testing.T) {
	defer gock.Off()
	gock.Observe(gock.DumpRequest)

	// get user ID
	gock.New(literal_5193).
		Post(literal_4172).
		MatchType("json").
		JSON(map[string]interface{}{
			"query": literal_8403,
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

	// get project ID
	gock.New(literal_5193).
		Post(literal_4172).
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
					"projectV2": map[string]interface{}{
						"id": "an ID",
					},
				},
			},
		})

	// delete project
	gock.New(literal_5193).
		Post(literal_4172).
		BodyString(`{"query":"mutation DeleteProject.*","variables":{"afterFields":null,"afterItems":null,"firstFields":0,"firstItems":0,"input":{"projectId":"an ID"}}}`).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"deleteProjectV2": map[string]interface{}{
					"projectV2": map[string]interface{}{
						"id":     literal_8307,
						"number": 1,
					},
				},
			},
		})

	client := queries.NewTestClient()

	ios, _, stdout, _ := iostreams.Test()
	ios.SetStdoutTTY(true)
	config := deleteConfig{
		opts: deleteOpts{
			owner:  "monalisa",
			number: 1,
		},
		client: client,
		io:     ios,
	}

	err := runDelete(config)
	assert.NoError(t, err)
	assert.Equal(
		t,
		literal_5673,
		stdout.String())
}

func TestRunDeleteOrg(t *testing.T) {
	defer gock.Off()
	gock.Observe(gock.DumpRequest)

	// get org ID
	gock.New(literal_5193).
		Post(literal_4172).
		MatchType("json").
		JSON(map[string]interface{}{
			"query": literal_8403,
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

	// get project ID
	gock.New(literal_5193).
		Post(literal_4172).
		MatchType("json").
		JSON(map[string]interface{}{
			"query": "query OrgProject.*",
			"variables": map[string]interface{}{
				"login":       "github",
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
				"organization": map[string]interface{}{
					"projectV2": map[string]interface{}{
						"id": "an ID",
					},
				},
			},
		})

	// delete project
	gock.New(literal_5193).
		Post(literal_4172).
		BodyString(`{"query":"mutation DeleteProject.*","variables":{"afterFields":null,"afterItems":null,"firstFields":0,"firstItems":0,"input":{"projectId":"an ID"}}}`).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"deleteProjectV2": map[string]interface{}{
					"projectV2": map[string]interface{}{
						"id":     literal_8307,
						"number": 1,
					},
				},
			},
		})

	client := queries.NewTestClient()

	ios, _, stdout, _ := iostreams.Test()
	ios.SetStdoutTTY(true)
	config := deleteConfig{
		opts: deleteOpts{
			owner:  "github",
			number: 1,
		},
		client: client,
		io:     ios,
	}

	err := runDelete(config)
	assert.NoError(t, err)
	assert.Equal(
		t,
		literal_5673,
		stdout.String())
}

func TestRunDeleteMe(t *testing.T) {
	defer gock.Off()
	gock.Observe(gock.DumpRequest)

	// get viewer ID
	gock.New(literal_5193).
		Post(literal_4172).
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

	// get project ID
	gock.New(literal_5193).
		Post(literal_4172).
		MatchType("json").
		JSON(map[string]interface{}{
			"query": "query ViewerProject.*",
			"variables": map[string]interface{}{
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
				"viewer": map[string]interface{}{
					"projectV2": map[string]interface{}{
						"id": "an ID",
					},
				},
			},
		})

	// delete project
	gock.New(literal_5193).
		Post(literal_4172).
		BodyString(`{"query":"mutation DeleteProject.*","variables":{"afterFields":null,"afterItems":null,"firstFields":0,"firstItems":0,"input":{"projectId":"an ID"}}}`).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"deleteProjectV2": map[string]interface{}{
					"projectV2": map[string]interface{}{
						"id":     literal_8307,
						"number": 1,
					},
				},
			},
		})

	client := queries.NewTestClient()

	ios, _, stdout, _ := iostreams.Test()
	ios.SetStdoutTTY(true)
	config := deleteConfig{
		opts: deleteOpts{
			owner:  "@me",
			number: 1,
		},
		client: client,
		io:     ios,
	}

	err := runDelete(config)
	assert.NoError(t, err)
	assert.Equal(
		t,
		literal_5673,
		stdout.String())
}

func TestRunDeleteJSON(t *testing.T) {
	defer gock.Off()
	gock.Observe(gock.DumpRequest)

	// get user ID
	gock.New(literal_5193).
		Post(literal_4172).
		MatchType("json").
		JSON(map[string]interface{}{
			"query": literal_8403,
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

	// get project ID
	gock.New(literal_5193).
		Post(literal_4172).
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
					"projectV2": map[string]interface{}{
						"id": "an ID",
					},
				},
			},
		})

	// delete project
	gock.New(literal_5193).
		Post(literal_4172).
		BodyString(`{"query":"mutation DeleteProject.*","variables":{"afterFields":null,"afterItems":null,"firstFields":0,"firstItems":0,"input":{"projectId":"an ID"}}}`).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"deleteProjectV2": map[string]interface{}{
					"projectV2": map[string]interface{}{
						"id":     literal_8307,
						"number": 1,
					},
				},
			},
		})

	client := queries.NewTestClient()

	ios, _, stdout, _ := iostreams.Test()
	config := deleteConfig{
		opts: deleteOpts{
			owner:    "monalisa",
			number:   1,
			exporter: cmdutil.NewJSONExporter(),
		},
		client: client,
		io:     ios,
	}

	err := runDelete(config)
	assert.NoError(t, err)
	assert.JSONEq(
		t,
		`{"number":1,"url":"","shortDescription":"","public":false,"closed":false,"title":"","id":literal_8307,"readme":"","items":{"totalCount":0},"fields":{"totalCount":0},"owner":{"type":"","login":""}}`,
		stdout.String())
}

const literal_5193 = "https://api.github.com"

const literal_4172 = "/graphql"

const literal_8403 = "query UserOrgOwner.*"

const literal_8307 = "project ID"

const literal_5673 = "Deleted project 1\n"
