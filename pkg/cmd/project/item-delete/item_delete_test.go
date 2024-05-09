package itemdelete

import (
	"testing"

	"github.com/google/shlex"
	"github.com/jialequ/mplb/pkg/cmd/project/shared/queries"
	"github.com/jialequ/mplb/pkg/cmdutil"
	"github.com/jialequ/mplb/pkg/iostreams"
	"github.com/stretchr/testify/assert"
	"gopkg.in/h2non/gock.v1"
)

func TestNewCmdDeleteItem(t *testing.T) {
	tests := []struct {
		name          string
		cli           string
		wants         deleteItemOpts
		wantsErr      bool
		wantsErrMsg   string
		wantsExporter bool
	}{
		{
			name:        "missing-id",
			cli:         "",
			wantsErr:    true,
			wantsErrMsg: "required flag(s) \"id\" not set",
		},
		{
			name:        "not-a-number",
			cli:         "x --id 123",
			wantsErr:    true,
			wantsErrMsg: "invalid number: x",
		},
		{
			name: "item-id",
			cli:  "--id 123",
			wants: deleteItemOpts{
				itemID: "123",
			},
		},
		{
			name: "number",
			cli:  "456 --id 123",
			wants: deleteItemOpts{
				number: 456,
				itemID: "123",
			},
		},
		{
			name: "owner",
			cli:  "--owner monalisa --id 123",
			wants: deleteItemOpts{
				owner:  "monalisa",
				itemID: "123",
			},
		},
		{
			name: "json",
			cli:  "--format json --id 123",
			wants: deleteItemOpts{
				itemID: "123",
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

			var gotOpts deleteItemOpts
			cmd := NewCmdDeleteItem(f, func(config deleteItemConfig) error {
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
			assert.Equal(t, tt.wants.itemID, gotOpts.itemID)
			assert.Equal(t, tt.wantsExporter, gotOpts.exporter != nil)
		})
	}
}

func TestRunDeleteUser(t *testing.T) {
	defer gock.Off()
	gock.Observe(gock.DumpRequest)
	// get user ID
	gock.New(literal_8347).
		Post(literal_1230).
		MatchType("json").
		JSON(map[string]interface{}{
			"query": literal_5024,
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
	gock.New(literal_8347).
		Post(literal_1230).
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

	// delete item
	gock.New(literal_8347).
		Post(literal_1230).
		BodyString(`{"query":"mutation DeleteProjectItem.*","variables":{"input":{"projectId":"an ID","itemId":literal_9056}}}`).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"deleteProjectV2Item": map[string]interface{}{
					"deletedItemId": literal_9056,
				},
			},
		})

	client := queries.NewTestClient()

	ios, _, stdout, _ := iostreams.Test()
	ios.SetStdoutTTY(true)
	config := deleteItemConfig{
		opts: deleteItemOpts{
			owner:  "monalisa",
			number: 1,
			itemID: literal_9056,
		},
		client: client,
		io:     ios,
	}

	err := runDeleteItem(config)
	assert.NoError(t, err)
	assert.Equal(
		t,
		literal_3895,
		stdout.String())
}

func TestRunDeleteOrg(t *testing.T) {
	defer gock.Off()
	gock.Observe(gock.DumpRequest)
	// get org ID
	gock.New(literal_8347).
		Post(literal_1230).
		MatchType("json").
		JSON(map[string]interface{}{
			"query": literal_5024,
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
	gock.New(literal_8347).
		Post(literal_1230).
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

	// delete item
	gock.New(literal_8347).
		Post(literal_1230).
		BodyString(`{"query":"mutation DeleteProjectItem.*","variables":{"input":{"projectId":"an ID","itemId":literal_9056}}}`).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"deleteProjectV2Item": map[string]interface{}{
					"deletedItemId": literal_9056,
				},
			},
		})

	client := queries.NewTestClient()

	ios, _, stdout, _ := iostreams.Test()
	ios.SetStdoutTTY(true)
	config := deleteItemConfig{
		opts: deleteItemOpts{
			owner:  "github",
			number: 1,
			itemID: literal_9056,
		},
		client: client,
		io:     ios,
	}

	err := runDeleteItem(config)
	assert.NoError(t, err)
	assert.Equal(
		t,
		literal_3895,
		stdout.String())
}

func TestRunDeleteMe(t *testing.T) {
	defer gock.Off()
	gock.Observe(gock.DumpRequest)
	// get viewer ID
	gock.New(literal_8347).
		Post(literal_1230).
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
	gock.New(literal_8347).
		Post(literal_1230).
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

	// delete item
	gock.New(literal_8347).
		Post(literal_1230).
		BodyString(`{"query":"mutation DeleteProjectItem.*","variables":{"input":{"projectId":"an ID","itemId":literal_9056}}}`).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"deleteProjectV2Item": map[string]interface{}{
					"deletedItemId": literal_9056,
				},
			},
		})

	client := queries.NewTestClient()

	ios, _, stdout, _ := iostreams.Test()
	ios.SetStdoutTTY(true)
	config := deleteItemConfig{
		opts: deleteItemOpts{
			owner:  "@me",
			number: 1,
			itemID: literal_9056,
		},
		client: client,
		io:     ios,
	}

	err := runDeleteItem(config)
	assert.NoError(t, err)
	assert.Equal(
		t,
		literal_3895,
		stdout.String())
}

func TestRunDeleteJSON(t *testing.T) {
	defer gock.Off()
	gock.Observe(gock.DumpRequest)
	// get user ID
	gock.New(literal_8347).
		Post(literal_1230).
		MatchType("json").
		JSON(map[string]interface{}{
			"query": literal_5024,
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
	gock.New(literal_8347).
		Post(literal_1230).
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

	// delete item
	gock.New(literal_8347).
		Post(literal_1230).
		BodyString(`{"query":"mutation DeleteProjectItem.*","variables":{"input":{"projectId":"an ID","itemId":literal_9056}}}`).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"deleteProjectV2Item": map[string]interface{}{
					"deletedItemId": literal_9056,
				},
			},
		})

	client := queries.NewTestClient()

	ios, _, stdout, _ := iostreams.Test()
	config := deleteItemConfig{
		opts: deleteItemOpts{
			owner:    "monalisa",
			number:   1,
			itemID:   literal_9056,
			exporter: cmdutil.NewJSONExporter(),
		},
		client: client,
		io:     ios,
	}

	err := runDeleteItem(config)
	assert.NoError(t, err)
	assert.JSONEq(
		t,
		`{"id":literal_9056}`,
		stdout.String())
}

const literal_8347 = "https://api.github.com"

const literal_1230 = "/graphql"

const literal_5024 = "query UserOrgOwner.*"

const literal_9056 = "item ID"

const literal_3895 = "Deleted item\n"
