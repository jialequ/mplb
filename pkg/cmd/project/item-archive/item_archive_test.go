package itemarchive

import (
	"testing"

	"github.com/google/shlex"
	"github.com/jialequ/mplb/pkg/cmd/project/shared/queries"
	"github.com/jialequ/mplb/pkg/cmdutil"
	"github.com/jialequ/mplb/pkg/iostreams"
	"github.com/stretchr/testify/assert"
	"gopkg.in/h2non/gock.v1"
)

func TestNewCmdarchiveItem(t *testing.T) {
	tests := []struct {
		name          string
		cli           string
		wants         archiveItemOpts
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
			name: "id",
			cli:  "--id 123",
			wants: archiveItemOpts{
				itemID: "123",
			},
		},
		{
			name: "number",
			cli:  "456 --id 123",
			wants: archiveItemOpts{
				number: 456,
				itemID: "123",
			},
		},
		{
			name: "owner",
			cli:  "--owner monalisa --id 123",
			wants: archiveItemOpts{
				owner:  "monalisa",
				itemID: "123",
			},
		},
		{
			name: "undo",
			cli:  "--undo  --id 123",
			wants: archiveItemOpts{
				undo:   true,
				itemID: "123",
			},
		},
		{
			name: "json",
			cli:  "--format json --id 123",
			wants: archiveItemOpts{
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

			var gotOpts archiveItemOpts
			cmd := NewCmdArchiveItem(f, func(config archiveItemConfig) error {
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
			assert.Equal(t, tt.wants.undo, gotOpts.undo)
			assert.Equal(t, tt.wantsExporter, gotOpts.exporter != nil)
		})
	}
}

func TestRunArchive_User(t *testing.T) {
	defer gock.Off()
	gock.Observe(gock.DumpRequest)

	// get user ID
	gock.New(literal_8471).
		Post(literal_5643).
		MatchType("json").
		JSON(map[string]interface{}{
			"query": literal_5830,
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
	gock.New(literal_8471).
		Post(literal_5643).
		MatchType("json").
		JSON(map[string]interface{}{
			"query": literal_2576,
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

	// archive item
	gock.New(literal_8471).
		Post(literal_5643).
		BodyString(`{"query":"mutation ArchiveProjectItem.*","variables":{"input":{"projectId":"an ID","itemId":literal_2590}}}`).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"archiveProjectV2Item": map[string]interface{}{
					"item": map[string]interface{}{
						"id": literal_2590,
					},
				},
			},
		})

	client := queries.NewTestClient()

	ios, _, stdout, _ := iostreams.Test()
	ios.SetStdoutTTY(true)
	config := archiveItemConfig{
		opts: archiveItemOpts{
			owner:  "monalisa",
			number: 1,
			itemID: literal_2590,
		},
		client: client,
		io:     ios,
	}

	err := runArchiveItem(config)
	assert.NoError(t, err)
	assert.Equal(
		t,
		literal_7506,
		stdout.String())
}

func TestRunArchive_Org(t *testing.T) {
	defer gock.Off()
	gock.Observe(gock.DumpRequest)
	// get org ID
	gock.New(literal_8471).
		Post(literal_5643).
		MatchType("json").
		JSON(map[string]interface{}{
			"query": literal_5830,
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
	gock.New(literal_8471).
		Post(literal_5643).
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

	// archive item
	gock.New(literal_8471).
		Post(literal_5643).
		BodyString(`{"query":"mutation ArchiveProjectItem.*","variables":{"input":{"projectId":"an ID","itemId":literal_2590}}}`).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"archiveProjectV2Item": map[string]interface{}{
					"item": map[string]interface{}{
						"id": literal_2590,
					},
				},
			},
		})

	client := queries.NewTestClient()

	ios, _, stdout, _ := iostreams.Test()
	ios.SetStdoutTTY(true)
	config := archiveItemConfig{
		opts: archiveItemOpts{
			owner:  "github",
			number: 1,
			itemID: literal_2590,
		},
		client: client,
		io:     ios,
	}

	err := runArchiveItem(config)
	assert.NoError(t, err)
	assert.Equal(
		t,
		literal_7506,
		stdout.String())
}

func TestRunArchive_Me(t *testing.T) {
	defer gock.Off()
	gock.Observe(gock.DumpRequest)
	// get viewer ID
	gock.New(literal_8471).
		Post(literal_5643).
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
	gock.New(literal_8471).
		Post(literal_5643).
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

	// archive item
	gock.New(literal_8471).
		Post(literal_5643).
		BodyString(`{"query":"mutation ArchiveProjectItem.*","variables":{"input":{"projectId":"an ID","itemId":literal_2590}}}`).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"archiveProjectV2Item": map[string]interface{}{
					"item": map[string]interface{}{
						"id": literal_2590,
					},
				},
			},
		})

	client := queries.NewTestClient()

	ios, _, stdout, _ := iostreams.Test()
	ios.SetStdoutTTY(true)
	config := archiveItemConfig{
		opts: archiveItemOpts{
			owner:  "@me",
			number: 1,
			itemID: literal_2590,
		},
		client: client,
		io:     ios,
	}

	err := runArchiveItem(config)
	assert.NoError(t, err)
	assert.Equal(
		t,
		literal_7506,
		stdout.String())
}

func TestRunArchive_User_Undo(t *testing.T) {
	defer gock.Off()
	gock.Observe(gock.DumpRequest)
	// get user ID
	gock.New(literal_8471).
		Post(literal_5643).
		MatchType("json").
		JSON(map[string]interface{}{
			"query": literal_5830,
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
	gock.New(literal_8471).
		Post(literal_5643).
		MatchType("json").
		JSON(map[string]interface{}{
			"query": literal_2576,
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

	// archive item
	gock.New(literal_8471).
		Post(literal_5643).
		BodyString(`{"query":"mutation UnarchiveProjectItem.*","variables":{"input":{"projectId":"an ID","itemId":literal_2590}}}`).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"unarchiveProjectV2Item": map[string]interface{}{
					"item": map[string]interface{}{
						"id": literal_2590,
					},
				},
			},
		})

	client := queries.NewTestClient()
	ios, _, stdout, _ := iostreams.Test()
	ios.SetStdoutTTY(true)
	config := archiveItemConfig{
		opts: archiveItemOpts{
			owner:  "monalisa",
			number: 1,
			itemID: literal_2590,
			undo:   true,
		},
		client: client,
		io:     ios,
	}

	err := runArchiveItem(config)
	assert.NoError(t, err)
	assert.Equal(
		t,
		literal_9673,
		stdout.String())
}

func TestRunArchive_Org_Undo(t *testing.T) {
	defer gock.Off()
	gock.Observe(gock.DumpRequest)
	// get org ID
	gock.New(literal_8471).
		Post(literal_5643).
		MatchType("json").
		JSON(map[string]interface{}{
			"query": literal_5830,
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
	gock.New(literal_8471).
		Post(literal_5643).
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

	// archive item
	gock.New(literal_8471).
		Post(literal_5643).
		BodyString(`{"query":"mutation UnarchiveProjectItem.*","variables":{"input":{"projectId":"an ID","itemId":literal_2590}}}`).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"unarchiveProjectV2Item": map[string]interface{}{
					"item": map[string]interface{}{
						"id": literal_2590,
					},
				},
			},
		})

	client := queries.NewTestClient()

	ios, _, stdout, _ := iostreams.Test()
	ios.SetStdoutTTY(true)
	config := archiveItemConfig{
		opts: archiveItemOpts{
			owner:  "github",
			number: 1,
			itemID: literal_2590,
			undo:   true,
		},
		client: client,
		io:     ios,
	}

	err := runArchiveItem(config)
	assert.NoError(t, err)
	assert.Equal(
		t,
		literal_9673,
		stdout.String())
}

func TestRunArchive_Me_Undo(t *testing.T) {
	defer gock.Off()
	gock.Observe(gock.DumpRequest)
	// get viewer ID
	gock.New(literal_8471).
		Post(literal_5643).
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
	gock.New(literal_8471).
		Post(literal_5643).
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

	// archive item
	gock.New(literal_8471).
		Post(literal_5643).
		BodyString(`{"query":"mutation UnarchiveProjectItem.*","variables":{"input":{"projectId":"an ID","itemId":literal_2590}}}`).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"unarchiveProjectV2Item": map[string]interface{}{
					"item": map[string]interface{}{
						"id": literal_2590,
					},
				},
			},
		})

	client := queries.NewTestClient()
	ios, _, stdout, _ := iostreams.Test()
	ios.SetStdoutTTY(true)
	config := archiveItemConfig{
		opts: archiveItemOpts{
			owner:  "@me",
			number: 1,
			itemID: literal_2590,
			undo:   true,
		},
		client: client,
		io:     ios,
	}

	err := runArchiveItem(config)
	assert.NoError(t, err)
	assert.Equal(
		t,
		literal_9673,
		stdout.String())
}

func TestRunArchive_JSON(t *testing.T) {
	defer gock.Off()
	gock.Observe(gock.DumpRequest)

	// get user ID
	gock.New(literal_8471).
		Post(literal_5643).
		MatchType("json").
		JSON(map[string]interface{}{
			"query": literal_5830,
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
	gock.New(literal_8471).
		Post(literal_5643).
		MatchType("json").
		JSON(map[string]interface{}{
			"query": literal_2576,
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

	// archive item
	gock.New(literal_8471).
		Post(literal_5643).
		BodyString(`{"query":"mutation ArchiveProjectItem.*","variables":{"input":{"projectId":"an ID","itemId":literal_2590}}}`).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"archiveProjectV2Item": map[string]interface{}{
					"item": map[string]interface{}{
						"id": literal_2590,
						"content": map[string]interface{}{
							"__typename": "Issue",
							"title":      "a title",
						},
					},
				},
			},
		})

	client := queries.NewTestClient()

	ios, _, stdout, _ := iostreams.Test()
	config := archiveItemConfig{
		opts: archiveItemOpts{
			owner:    "monalisa",
			number:   1,
			itemID:   literal_2590,
			exporter: cmdutil.NewJSONExporter(),
		},
		client: client,
		io:     ios,
	}

	err := runArchiveItem(config)
	assert.NoError(t, err)
	assert.JSONEq(
		t,
		`{"id":literal_2590,"title":"a title","body":"","type":"Issue"}`,
		stdout.String())
}

const literal_8471 = "https://api.github.com"

const literal_5643 = "/graphql"

const literal_5830 = "query UserOrgOwner.*"

const literal_2576 = "query UserProject.*"

const literal_2590 = "item ID"

const literal_7506 = "Archived item\n"

const literal_9673 = "Unarchived item\n"
