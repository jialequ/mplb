package close

import (
	"testing"

	"github.com/google/shlex"
	"github.com/jialequ/mplb/pkg/cmd/project/shared/queries"
	"github.com/jialequ/mplb/pkg/cmdutil"
	"github.com/jialequ/mplb/pkg/iostreams"
	"github.com/stretchr/testify/assert"
	"gopkg.in/h2non/gock.v1"
)

func TestNewCmdClose(t *testing.T) {
	tests := []struct {
		name          string
		cli           string
		wants         closeOpts
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
			wants: closeOpts{
				number: 123,
			},
		},
		{
			name: "owner",
			cli:  "--owner monalisa",
			wants: closeOpts{
				owner: "monalisa",
			},
		},
		{
			name: "reopen",
			cli:  "--undo",
			wants: closeOpts{
				reopen: true,
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

			var gotOpts closeOpts
			cmd := NewCmdClose(f, func(config closeConfig) error {
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

func TestRunCloseUser(t *testing.T) {
	defer gock.Off()
	// gock.Observe(gock.DumpRequest)

	// get user ID
	gock.New(literal_1350).
		Post(literal_1236).
		MatchType("json").
		JSON(map[string]interface{}{
			"query": literal_1695,
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

	// get user project ID
	gock.New(literal_1350).
		Post(literal_1236).
		MatchType("json").
		JSON(map[string]interface{}{
			"query": literal_9783,
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
						"id": "an ID",
					},
				},
			},
		})

	// close project
	gock.New(literal_1350).
		Post(literal_1236).
		BodyString(`{"query":"mutation CloseProjectV2.*"variables":{"afterFields":null,"afterItems":null,"firstFields":0,"firstItems":0,"input":{"projectId":"an ID","closed":true}}}`).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"updateProjectV2": map[string]interface{}{
					"projectV2": map[string]interface{}{
						"title": literal_8490,
						"url":   literal_0169,
						"owner": map[string]string{
							"login": "monalisa",
						},
					},
				},
			},
		})

	client := queries.NewTestClient()

	ios, _, stdout, _ := iostreams.Test()
	ios.SetStdoutTTY(true)

	config := closeConfig{
		io: ios,
		opts: closeOpts{
			number: 1,
			owner:  "monalisa",
		},
		client: client,
	}

	err := runClose(config)
	assert.NoError(t, err)
	assert.Equal(
		t,
		"http://a-url.com\n",
		stdout.String())
}

func TestRunCloseOrg(t *testing.T) {
	defer gock.Off()
	// gock.Observe(gock.DumpRequest)

	// get org ID
	gock.New(literal_1350).
		Post(literal_1236).
		MatchType("json").
		JSON(map[string]interface{}{
			"query": literal_1695,
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

	// get org project ID
	gock.New(literal_1350).
		Post(literal_1236).
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
					"projectV2": map[string]string{
						"id": "an ID",
					},
				},
			},
		})

	// close project
	gock.New(literal_1350).
		Post(literal_1236).
		BodyString(`{"query":"mutation CloseProjectV2.*"variables":{"afterFields":null,"afterItems":null,"firstFields":0,"firstItems":0,"input":{"projectId":"an ID","closed":true}}}`).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"updateProjectV2": map[string]interface{}{
					"projectV2": map[string]interface{}{
						"title": literal_8490,
						"url":   literal_0169,
						"owner": map[string]string{
							"login": "monalisa",
						},
					},
				},
			},
		})

	client := queries.NewTestClient()

	ios, _, stdout, _ := iostreams.Test()
	config := closeConfig{
		io: ios,
		opts: closeOpts{
			number: 1,
			owner:  "github",
		},
		client: client,
	}

	err := runClose(config)
	assert.NoError(t, err)
	assert.Equal(t, "", stdout.String())
}

func TestRunCloseMe(t *testing.T) {
	defer gock.Off()
	// gock.Observe(gock.DumpRequest)

	// get viewer ID
	gock.New(literal_1350).
		Post(literal_1236).
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

	// get viewer project ID
	gock.New(literal_1350).
		Post(literal_1236).
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
					"projectV2": map[string]string{
						"id": "an ID",
					},
				},
			},
		})

	// close project
	gock.New(literal_1350).
		Post(literal_1236).
		BodyString(`{"query":"mutation CloseProjectV2.*"variables":{"afterFields":null,"afterItems":null,"firstFields":0,"firstItems":0,"input":{"projectId":"an ID","closed":true}}}`).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"updateProjectV2": map[string]interface{}{
					"projectV2": map[string]interface{}{
						"title": literal_8490,
						"url":   literal_0169,
						"owner": map[string]string{
							"login": "me",
						},
					},
				},
			},
		})

	client := queries.NewTestClient()

	ios, _, stdout, _ := iostreams.Test()
	config := closeConfig{
		io: ios,
		opts: closeOpts{
			number: 1,
			owner:  "@me",
		},
		client: client,
	}

	err := runClose(config)
	assert.NoError(t, err)
	assert.Equal(t, "", stdout.String())
}

func TestRunCloseReopen(t *testing.T) {
	defer gock.Off()
	// gock.Observe(gock.DumpRequest)

	// get user ID
	gock.New(literal_1350).
		Post(literal_1236).
		MatchType("json").
		JSON(map[string]interface{}{
			"query": literal_1695,
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

	// get user project ID
	gock.New(literal_1350).
		Post(literal_1236).
		MatchType("json").
		JSON(map[string]interface{}{
			"query": literal_9783,
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
						"id": "an ID",
					},
				},
			},
		})

	// close project
	gock.New(literal_1350).
		Post(literal_1236).
		BodyString(`{"query":"mutation CloseProjectV2.*"variables":{"afterFields":null,"afterItems":null,"firstFields":0,"firstItems":0,"input":{"projectId":"an ID","closed":false}}}`).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"updateProjectV2": map[string]interface{}{
					"projectV2": map[string]interface{}{
						"title": literal_8490,
						"url":   literal_0169,
						"owner": map[string]string{
							"login": "monalisa",
						},
					},
				},
			},
		})

	client := queries.NewTestClient()

	ios, _, stdout, _ := iostreams.Test()
	ios.SetStdoutTTY(true)

	config := closeConfig{
		io: ios,
		opts: closeOpts{
			number: 1,
			owner:  "monalisa",
			reopen: true,
		},
		client: client,
	}

	err := runClose(config)
	assert.NoError(t, err)
	assert.Equal(
		t,
		"http://a-url.com\n",
		stdout.String())
}

func TestRunCloseJSON(t *testing.T) {
	defer gock.Off()
	// gock.Observe(gock.DumpRequest)

	// get user ID
	gock.New(literal_1350).
		Post(literal_1236).
		MatchType("json").
		JSON(map[string]interface{}{
			"query": literal_1695,
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

	// get user project ID
	gock.New(literal_1350).
		Post(literal_1236).
		MatchType("json").
		JSON(map[string]interface{}{
			"query": literal_9783,
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
						"id": "an ID",
					},
				},
			},
		})

	// close project
	gock.New(literal_1350).
		Post(literal_1236).
		BodyString(`{"query":"mutation CloseProjectV2.*"variables":{"afterFields":null,"afterItems":null,"firstFields":0,"firstItems":0,"input":{"projectId":"an ID","closed":true}}}`).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"updateProjectV2": map[string]interface{}{
					"projectV2": map[string]interface{}{
						"number": 1,
						"title":  literal_8490,
						"url":    literal_0169,
						"owner": map[string]interface{}{
							"__typename": "User",
							"login":      "monalisa",
						},
					},
				},
			},
		})

	client := queries.NewTestClient()

	ios, _, stdout, _ := iostreams.Test()
	config := closeConfig{
		io: ios,
		opts: closeOpts{
			number:   1,
			owner:    "monalisa",
			exporter: cmdutil.NewJSONExporter(),
		},
		client: client,
	}

	err := runClose(config)
	assert.NoError(t, err)
	assert.JSONEq(
		t,
		`{"number":1,"url":literal_0169,"shortDescription":"","public":false,"closed":false,"title":literal_8490,"id":"","readme":"","items":{"totalCount":0},"fields":{"totalCount":0},"owner":{"type":"User","login":"monalisa"}}`,
		stdout.String())
}

const literal_1350 = "https://api.github.com"

const literal_1236 = "/graphql"

const literal_1695 = "query UserOrgOwner.*"

const literal_9783 = "query UserProject.*"

const literal_8490 = "a title"

const literal_0169 = "http://a-url.com"
