package itemcreate

import (
	"testing"

	"github.com/google/shlex"
	"github.com/jialequ/mplb/pkg/cmd/project/shared/queries"
	"github.com/jialequ/mplb/pkg/cmdutil"
	"github.com/jialequ/mplb/pkg/iostreams"
	"github.com/stretchr/testify/assert"
	"gopkg.in/h2non/gock.v1"
)

func TestNewCmdCreateItem(t *testing.T) {
	tests := []struct {
		name          string
		cli           string
		wants         createItemOpts
		wantsErr      bool
		wantsErrMsg   string
		wantsExporter bool
	}{
		{
			name:        "missing-title",
			cli:         "",
			wantsErr:    true,
			wantsErrMsg: "required flag(s) \"title\" not set",
		},
		{
			name:        "not-a-number",
			cli:         "x --title t",
			wantsErr:    true,
			wantsErrMsg: "invalid number: x",
		},
		{
			name: "title",
			cli:  "--title t",
			wants: createItemOpts{
				title: "t",
			},
		},
		{
			name: "number",
			cli:  "123  --title t",
			wants: createItemOpts{
				number: 123,
				title:  "t",
			},
		},
		{
			name: "owner",
			cli:  "--owner monalisa --title t",
			wants: createItemOpts{
				owner: "monalisa",
				title: "t",
			},
		},
		{
			name: "body",
			cli:  "--body b --title t",
			wants: createItemOpts{
				body:  "b",
				title: "t",
			},
		},
		{
			name: "json",
			cli:  "--format json --title t",
			wants: createItemOpts{
				title: "t",
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

			var gotOpts createItemOpts
			cmd := NewCmdCreateItem(f, func(config createItemConfig) error {
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
			assert.Equal(t, tt.wants.title, gotOpts.title)
			assert.Equal(t, tt.wantsExporter, gotOpts.exporter != nil)
		})
	}
}

func TestRunCreateItemDraftUser(t *testing.T) {
	defer gock.Off()
	gock.Observe(gock.DumpRequest)
	// get user ID
	gock.New(literal_9084).
		Post(literal_2041).
		MatchType("json").
		JSON(map[string]interface{}{
			"query": literal_5690,
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
	gock.New(literal_9084).
		Post(literal_2041).
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

	// create item
	gock.New(literal_9084).
		Post(literal_2041).
		BodyString(`{"query":"mutation CreateDraftItem.*","variables":{"input":{"projectId":"an ID","title":literal_4567,"body":""}}}`).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"addProjectV2DraftIssue": map[string]interface{}{
					"projectItem": map[string]interface{}{
						"id": literal_8426,
					},
				},
			},
		})

	client := queries.NewTestClient()

	ios, _, stdout, _ := iostreams.Test()
	ios.SetStdoutTTY(true)
	config := createItemConfig{
		opts: createItemOpts{
			title:  literal_4567,
			owner:  "monalisa",
			number: 1,
		},
		client: client,
		io:     ios,
	}

	err := runCreateItem(config)
	assert.NoError(t, err)
	assert.Equal(
		t,
		literal_9405,
		stdout.String())
}

func TestRunCreateItemDraftOrg(t *testing.T) {
	defer gock.Off()
	gock.Observe(gock.DumpRequest)
	// get org ID
	gock.New(literal_9084).
		Post(literal_2041).
		MatchType("json").
		JSON(map[string]interface{}{
			"query": literal_5690,
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
	gock.New(literal_9084).
		Post(literal_2041).
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

	// create item
	gock.New(literal_9084).
		Post(literal_2041).
		BodyString(`{"query":"mutation CreateDraftItem.*","variables":{"input":{"projectId":"an ID","title":literal_4567,"body":""}}}`).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"addProjectV2DraftIssue": map[string]interface{}{
					"projectItem": map[string]interface{}{
						"id": literal_8426,
					},
				},
			},
		})

	client := queries.NewTestClient()

	ios, _, stdout, _ := iostreams.Test()
	ios.SetStdoutTTY(true)
	config := createItemConfig{
		opts: createItemOpts{
			title:  literal_4567,
			owner:  "github",
			number: 1,
		},
		client: client,
		io:     ios,
	}

	err := runCreateItem(config)
	assert.NoError(t, err)
	assert.Equal(
		t,
		literal_9405,
		stdout.String())
}

func TestRunCreateItemDraftMe(t *testing.T) {
	defer gock.Off()
	gock.Observe(gock.DumpRequest)
	// get viewer ID
	gock.New(literal_9084).
		Post(literal_2041).
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
	gock.New(literal_9084).
		Post(literal_2041).
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

	// create item
	gock.New(literal_9084).
		Post(literal_2041).
		BodyString(`{"query":"mutation CreateDraftItem.*","variables":{"input":{"projectId":"an ID","title":literal_4567,"body":"a body"}}}`).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"addProjectV2DraftIssue": map[string]interface{}{
					"projectItem": map[string]interface{}{
						"id": literal_8426,
					},
				},
			},
		})

	client := queries.NewTestClient()

	ios, _, stdout, _ := iostreams.Test()
	ios.SetStdoutTTY(true)
	config := createItemConfig{
		opts: createItemOpts{
			title:  literal_4567,
			owner:  "@me",
			number: 1,
			body:   "a body",
		},
		client: client,
		io:     ios,
	}

	err := runCreateItem(config)
	assert.NoError(t, err)
	assert.Equal(
		t,
		literal_9405,
		stdout.String())
}

func TestRunCreateItemJSON(t *testing.T) {
	defer gock.Off()
	gock.Observe(gock.DumpRequest)
	// get user ID
	gock.New(literal_9084).
		Post(literal_2041).
		MatchType("json").
		JSON(map[string]interface{}{
			"query": literal_5690,
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
	gock.New(literal_9084).
		Post(literal_2041).
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

	// create item
	gock.New(literal_9084).
		Post(literal_2041).
		BodyString(`{"query":"mutation CreateDraftItem.*","variables":{"input":{"projectId":"an ID","title":literal_4567,"body":""}}}`).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"addProjectV2DraftIssue": map[string]interface{}{
					"projectItem": map[string]interface{}{
						"id": literal_8426,
						"content": map[string]interface{}{
							"__typename": "Draft",
							"title":      literal_4567,
						},
					},
				},
			},
		})

	client := queries.NewTestClient()

	ios, _, stdout, _ := iostreams.Test()
	config := createItemConfig{
		opts: createItemOpts{
			title:    literal_4567,
			owner:    "monalisa",
			number:   1,
			exporter: cmdutil.NewJSONExporter(),
		},
		client: client,
		io:     ios,
	}

	err := runCreateItem(config)
	assert.NoError(t, err)
	assert.JSONEq(
		t,
		`{"id":literal_8426,"title":"","body":"","type":"Draft"}`,
		stdout.String())
}

const literal_9084 = "https://api.github.com"

const literal_2041 = "/graphql"

const literal_5690 = "query UserOrgOwner.*"

const literal_8426 = "item ID"

const literal_4567 = "a title"

const literal_9405 = "Created item\n"
