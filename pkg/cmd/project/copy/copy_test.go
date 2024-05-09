package copy

import (
	"testing"

	"github.com/google/shlex"
	"github.com/jialequ/mplb/pkg/cmd/project/shared/queries"
	"github.com/jialequ/mplb/pkg/cmdutil"
	"github.com/jialequ/mplb/pkg/iostreams"
	"github.com/stretchr/testify/assert"
	"gopkg.in/h2non/gock.v1"
)

func TestNewCmdCopy(t *testing.T) {
	tests := []struct {
		name          string
		cli           string
		wants         copyOpts
		wantsErr      bool
		wantsErrMsg   string
		wantsExporter bool
	}{
		{
			name:        "not-a-number",
			cli:         "x --title t",
			wantsErr:    true,
			wantsErrMsg: "invalid number: x",
		},
		{
			name: "title",
			cli:  "--title t",
			wants: copyOpts{
				title: "t",
			},
		},
		{
			name: "number",
			cli:  "123 --title t",
			wants: copyOpts{
				number: 123,
				title:  "t",
			},
		},
		{
			name: "source-owner",
			cli:  "--source-owner monalisa --title t",
			wants: copyOpts{
				sourceOwner: "monalisa",
				title:       "t",
			},
		},
		{
			name: "target-owner",
			cli:  "--target-owner monalisa --title t",
			wants: copyOpts{
				targetOwner: "monalisa",
				title:       "t",
			},
		},
		{
			name: "drafts",
			cli:  "--drafts --title t",
			wants: copyOpts{
				includeDraftIssues: true,
				title:              "t",
			},
		},
		{
			name: "json",
			cli:  "--format json --title t",
			wants: copyOpts{
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

			var gotOpts copyOpts
			cmd := NewCmdCopy(f, func(config copyConfig) error {
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
			assert.Equal(t, tt.wants.sourceOwner, gotOpts.sourceOwner)
			assert.Equal(t, tt.wants.targetOwner, gotOpts.targetOwner)
			assert.Equal(t, tt.wants.title, gotOpts.title)
			assert.Equal(t, tt.wants.includeDraftIssues, gotOpts.includeDraftIssues)
			assert.Equal(t, tt.wantsExporter, gotOpts.exporter != nil)
		})
	}
}

func TestRunCopyUser(t *testing.T) {
	defer gock.Off()
	// gock.Observe(gock.DumpRequest)

	// get user project ID
	gock.New(literal_4156).
		Post(literal_8173).
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
						"id": "an ID",
					},
				},
			},
		})

	// get source user ID
	gock.New(literal_4156).
		Post(literal_8173).
		MatchType("json").
		JSON(map[string]interface{}{
			"query": literal_6238,
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

	// get target user ID
	gock.New(literal_4156).
		Post(literal_8173).
		MatchType("json").
		JSON(map[string]interface{}{
			"query": literal_6238,
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

	// Copy project
	gock.New(literal_4156).
		Post(literal_8173).
		BodyString(`{"query":"mutation CopyProjectV2.*","variables":{"afterFields":null,"afterItems":null,"firstFields":0,"firstItems":0,"input":{"projectId":"an ID","ownerId":"an ID","title":literal_3087,"includeDraftIssues":false}}}`).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"copyProjectV2": map[string]interface{}{
					"projectV2": map[string]interface{}{
						"title": literal_3087,
						"url":   literal_3865,
						"owner": map[string]string{
							"login": "monalisa",
						},
					},
				},
			},
		})

	client := queries.NewTestClient()

	ios, _, stdout, _ := iostreams.Test()
	ios.SetStdoutTTY(false)

	config := copyConfig{
		io: ios,
		opts: copyOpts{
			title:       literal_3087,
			sourceOwner: "monalisa",
			targetOwner: "monalisa",
			number:      1,
		},
		client: client,
	}

	err := runCopy(config)
	assert.NoError(t, err)
	assert.Equal(t, "", stdout.String())
}

func TestRunCopyOrg(t *testing.T) {
	defer gock.Off()
	// gock.Observe(gock.DumpRequest)

	// get org project ID
	gock.New(literal_4156).
		Post(literal_8173).
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
	// get source org ID
	gock.New(literal_4156).
		Post(literal_8173).
		MatchType("json").
		JSON(map[string]interface{}{
			"query": literal_6238,
			"variables": map[string]string{
				"login": "github",
			},
		}).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"organization": map[string]interface{}{
					"id":    "an ID",
					"login": "github",
				},
			},
			"errors": []interface{}{
				map[string]interface{}{
					"type": "NOT_FOUND",
					"path": []string{"user"},
				},
			},
		})

	// get target source org ID
	gock.New(literal_4156).
		Post(literal_8173).
		MatchType("json").
		JSON(map[string]interface{}{
			"query": literal_6238,
			"variables": map[string]string{
				"login": "github",
			},
		}).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"organization": map[string]interface{}{
					"id":    "an ID",
					"login": "github",
				},
			},
			"errors": []interface{}{
				map[string]interface{}{
					"type": "NOT_FOUND",
					"path": []string{"user"},
				},
			},
		})

	// Copy project
	gock.New(literal_4156).
		Post(literal_8173).
		BodyString(`{"query":"mutation CopyProjectV2.*","variables":{"afterFields":null,"afterItems":null,"firstFields":0,"firstItems":0,"input":{"projectId":"an ID","ownerId":"an ID","title":literal_3087,"includeDraftIssues":false}}}`).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"copyProjectV2": map[string]interface{}{
					"projectV2": map[string]interface{}{
						"title": literal_3087,
						"url":   literal_3865,
						"owner": map[string]string{
							"login": "monalisa",
						},
					},
				},
			},
		})

	client := queries.NewTestClient()

	ios, _, stdout, _ := iostreams.Test()
	ios.SetStdoutTTY(false)

	config := copyConfig{
		io: ios,
		opts: copyOpts{
			title:       literal_3087,
			sourceOwner: "github",
			targetOwner: "github",
			number:      1,
		},
		client: client,
	}

	err := runCopy(config)
	assert.NoError(t, err)
	assert.Equal(t, "", stdout.String())
}

func TestRunCopyMe(t *testing.T) {
	defer gock.Off()
	// gock.Observe(gock.DumpRequest)

	// get viewer project ID
	gock.New(literal_4156).
		Post(literal_8173).
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

	// get source viewer ID
	gock.New(literal_4156).
		Post(literal_8173).
		MatchType("json").
		JSON(map[string]interface{}{
			"query": "query ViewerOwner.*",
		}).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"viewer": map[string]interface{}{
					"id":    "an ID",
					"login": "me",
				},
			},
		})

	// get target viewer ID
	gock.New(literal_4156).
		Post(literal_8173).
		MatchType("json").
		JSON(map[string]interface{}{
			"query": "query ViewerOwner.*",
		}).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"viewer": map[string]interface{}{
					"id":    "an ID",
					"login": "me",
				},
			},
		})

	// Copy project
	gock.New(literal_4156).
		Post(literal_8173).
		BodyString(`{"query":"mutation CopyProjectV2.*","variables":{"afterFields":null,"afterItems":null,"firstFields":0,"firstItems":0,"input":{"projectId":"an ID","ownerId":"an ID","title":literal_3087,"includeDraftIssues":false}}}`).Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"copyProjectV2": map[string]interface{}{
					"projectV2": map[string]interface{}{
						"title": literal_3087,
						"url":   literal_3865,
						"owner": map[string]string{
							"login": "me",
						},
					},
				},
			},
		})

	client := queries.NewTestClient()

	ios, _, stdout, _ := iostreams.Test()
	ios.SetStdoutTTY(true)

	config := copyConfig{
		io: ios,
		opts: copyOpts{
			title:       literal_3087,
			sourceOwner: "@me",
			targetOwner: "@me",
			number:      1,
		},
		client: client,
	}

	err := runCopy(config)
	assert.NoError(t, err)
	assert.Equal(
		t,
		"http://a-url.com\n",
		stdout.String())
}

func TestRunCopyJSON(t *testing.T) {
	defer gock.Off()
	// gock.Observe(gock.DumpRequest)

	// get user project ID
	gock.New(literal_4156).
		Post(literal_8173).
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
						"id": "an ID",
					},
				},
			},
		})

	// get source user ID
	gock.New(literal_4156).
		Post(literal_8173).
		MatchType("json").
		JSON(map[string]interface{}{
			"query": literal_6238,
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

	// get target user ID
	gock.New(literal_4156).
		Post(literal_8173).
		MatchType("json").
		JSON(map[string]interface{}{
			"query": literal_6238,
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

	// Copy project
	gock.New(literal_4156).
		Post(literal_8173).
		BodyString(`{"query":"mutation CopyProjectV2.*","variables":{"afterFields":null,"afterItems":null,"firstFields":0,"firstItems":0,"input":{"projectId":"an ID","ownerId":"an ID","title":literal_3087,"includeDraftIssues":false}}}`).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"copyProjectV2": map[string]interface{}{
					"projectV2": map[string]interface{}{
						"number": 1,
						"title":  literal_3087,
						"url":    literal_3865,
						"owner": map[string]string{
							"login": "monalisa",
						},
					},
				},
			},
		})

	client := queries.NewTestClient()

	ios, _, stdout, _ := iostreams.Test()
	config := copyConfig{
		io: ios,
		opts: copyOpts{
			title:       literal_3087,
			sourceOwner: "monalisa",
			targetOwner: "monalisa",
			number:      1,
			exporter:    cmdutil.NewJSONExporter(),
		},
		client: client,
	}

	err := runCopy(config)
	assert.NoError(t, err)
	assert.JSONEq(
		t,
		`{"number":1,"url":literal_3865,"shortDescription":"","public":false,"closed":false,"title":literal_3087,"id":"","readme":"","items":{"totalCount":0},"fields":{"totalCount":0},"owner":{"type":"","login":"monalisa"}}`,
		stdout.String())
}

const literal_4156 = "https://api.github.com"

const literal_8173 = "/graphql"

const literal_6238 = "query UserOrgOwner.*"

const literal_3087 = "a title"

const literal_3865 = "http://a-url.com"
