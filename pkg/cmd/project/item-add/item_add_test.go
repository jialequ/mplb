package itemadd

import (
	"testing"

	"github.com/google/shlex"
	"github.com/jialequ/mplb/pkg/cmd/project/shared/templet"
	"github.com/jialequ/mplb/pkg/cmdutil"
	"github.com/jialequ/mplb/pkg/iostreams"
	"github.com/stretchr/testify/assert"
	"gopkg.in/h2non/gock.v1"
)

func TestNewCmdaddItem(t *testing.T) {
	tests := []struct {
		name          string
		cli           string
		wants         addItemOpts
		wantsErr      bool
		wantsErrMsg   string
		wantsExporter bool
	}{
		{
			name:        "missing-url",
			cli:         "",
			wantsErr:    true,
			wantsErrMsg: "required flag(s) \"url\" not set",
		},
		{
			name:        "not-a-number",
			cli:         "x --url github.com/cli/cli",
			wantsErr:    true,
			wantsErrMsg: "invalid number: x",
		},
		{
			name: "url",
			cli:  "--url github.com/cli/cli",
			wants: addItemOpts{
				itemURL: literal_7146,
			},
		},
		{
			name: "number",
			cli:  "123 --url github.com/cli/cli",
			wants: addItemOpts{
				number:  123,
				itemURL: literal_7146,
			},
		},
		{
			name: "owner",
			cli:  "--owner monalisa --url github.com/cli/cli",
			wants: addItemOpts{
				owner:   "monalisa",
				itemURL: literal_7146,
			},
		},
		{
			name: "json",
			cli:  "--format json --url github.com/cli/cli",
			wants: addItemOpts{
				itemURL: literal_7146,
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

			var gotOpts addItemOpts
			cmd := NewCmdAddItem(f, func(config addItemConfig) error {
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
			assert.Equal(t, tt.wants.itemURL, gotOpts.itemURL)
			assert.Equal(t, tt.wantsExporter, gotOpts.exporter != nil)
		})
	}
}

func TestRunAddItemUser(t *testing.T) {
	defer gock.Off()
	gock.Observe(gock.DumpRequest)

	// get user ID
	gock.New(literal_4502).
		Post(literal_8694).
		MatchType("json").
		JSON(map[string]interface{}{
			"query": literal_2170,
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
	gock.New(literal_4502).
		Post(literal_8694).
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

	// get item ID
	gock.New(literal_4502).
		Post(literal_8694).
		MatchType("json").
		JSON(map[string]interface{}{
			"query": literal_3467,
			"variables": map[string]interface{}{
				"url": literal_3614,
			},
		}).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"resource": map[string]interface{}{
					"id":         literal_5687,
					"__typename": "Issue",
				},
			},
		})

	// create item
	gock.New(literal_4502).
		Post(literal_8694).
		BodyString(`{"query":"mutation AddItem.*","variables":{"input":{"projectId":"an ID","contentId":literal_5687}}}`).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"addProjectV2ItemById": map[string]interface{}{
					"item": map[string]interface{}{
						"id": literal_5687,
					},
				},
			},
		})

	client := templet.NewTestClient()

	ios, _, _, _ := iostreams.Test()
	ios.SetStdoutTTY(true)
	config := addItemConfig{
		opts: addItemOpts{
			owner:   "monalisa",
			number:  1,
			itemURL: literal_3614,
		},
		client: client,
		io:     ios,
	}

	runAddItem(config)
}

func TestRunAddItemOrg(t *testing.T) {
	defer gock.Off()
	gock.Observe(gock.DumpRequest)
	// get org ID
	gock.New(literal_4502).
		Post(literal_8694).
		MatchType("json").
		JSON(map[string]interface{}{
			"query": literal_2170,
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
	gock.New(literal_4502).
		Post(literal_8694).
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

	// get item ID
	gock.New(literal_4502).
		Post(literal_8694).
		MatchType("json").
		JSON(map[string]interface{}{
			"query": literal_3467,
			"variables": map[string]interface{}{
				"url": literal_3614,
			},
		}).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"resource": map[string]interface{}{
					"id":         literal_5687,
					"__typename": "Issue",
				},
			},
		})

	// create item
	gock.New(literal_4502).
		Post(literal_8694).
		BodyString(`{"query":"mutation AddItem.*","variables":{"input":{"projectId":"an ID","contentId":literal_5687}}}`).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"addProjectV2ItemById": map[string]interface{}{
					"item": map[string]interface{}{
						"id": literal_5687,
					},
				},
			},
		})

	client := templet.NewTestClient()

	ios, _, _, _ := iostreams.Test()
	ios.SetStdoutTTY(true)
	config := addItemConfig{
		opts: addItemOpts{
			owner:   "github",
			number:  1,
			itemURL: literal_3614,
		},
		client: client,
		io:     ios,
	}

	runAddItem(config)

}

func TestRunAddItemMe(t *testing.T) {
	defer gock.Off()
	gock.Observe(gock.DumpRequest)
	// get viewer ID
	gock.New(literal_4502).
		Post(literal_8694).
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
	gock.New(literal_4502).
		Post(literal_8694).
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

	// get item ID
	gock.New(literal_4502).
		Post(literal_8694).
		MatchType("json").
		JSON(map[string]interface{}{
			"query": literal_3467,
			"variables": map[string]interface{}{
				"url": "https://github.com/cli/go-gh/pull/1",
			},
		}).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"resource": map[string]interface{}{
					"id":         literal_5687,
					"__typename": "PullRequest",
				},
			},
		})

	// create item
	gock.New(literal_4502).
		Post(literal_8694).
		BodyString(`{"query":"mutation AddItem.*","variables":{"input":{"projectId":"an ID","contentId":literal_5687}}}`).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"addProjectV2ItemById": map[string]interface{}{
					"item": map[string]interface{}{
						"id": literal_5687,
					},
				},
			},
		})

	client := templet.NewTestClient()

	ios, _, _, _ := iostreams.Test()
	ios.SetStdoutTTY(true)
	config := addItemConfig{
		opts: addItemOpts{
			owner:   "@me",
			number:  1,
			itemURL: "https://github.com/cli/go-gh/pull/1",
		},
		client: client,
		io:     ios,
	}

	runAddItem(config)
}

func TestRunAddItemJSON(t *testing.T) {
	defer gock.Off()
	gock.Observe(gock.DumpRequest)

	// get user ID
	gock.New(literal_4502).
		Post(literal_8694).
		MatchType("json").
		JSON(map[string]interface{}{
			"query": literal_2170,
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
	gock.New(literal_4502).
		Post(literal_8694).
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

	// get item ID
	gock.New(literal_4502).
		Post(literal_8694).
		MatchType("json").
		JSON(map[string]interface{}{
			"query": literal_3467,
			"variables": map[string]interface{}{
				"url": literal_3614,
			},
		}).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"resource": map[string]interface{}{
					"id":         literal_5687,
					"__typename": "Issue",
				},
			},
		})

	// create item
	gock.New(literal_4502).
		Post(literal_8694).
		BodyString(`{"query":"mutation AddItem.*","variables":{"input":{"projectId":"an ID","contentId":literal_5687}}}`).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"addProjectV2ItemById": map[string]interface{}{
					"item": map[string]interface{}{
						"id": literal_5687,
						"content": map[string]interface{}{
							"__typename": "Issue",
							"title":      "a title",
						},
					},
				},
			},
		})

	client := templet.NewTestClient()

	ios, _, _, _ := iostreams.Test()
	config := addItemConfig{
		opts: addItemOpts{
			owner:    "monalisa",
			number:   1,
			itemURL:  literal_3614,
			exporter: cmdutil.NewJSONExporter(),
		},
		client: client,
		io:     ios,
	}

	runAddItem(config)
}

const literal_7146 = "github.com/cli/cli"

const literal_4502 = "https://api.github.com"

const literal_8694 = "/graphql"

const literal_2170 = "query UserOrgOwner.*"

const literal_3467 = "query GetIssueOrPullRequest.*"

const literal_3614 = "https://github.com/cli/go-gh/issues/1"

const literal_5687 = "item ID"

const literal_3781 = "Added item\n"
