package edit

import (
	"testing"

	"github.com/google/shlex"
	"github.com/jialequ/mplb/pkg/cmd/project/shared/queries"
	"github.com/jialequ/mplb/pkg/cmdutil"
	"github.com/jialequ/mplb/pkg/iostreams"
	"github.com/stretchr/testify/assert"
	"gopkg.in/h2non/gock.v1"
)

func TestNewCmdEdit(t *testing.T) {
	tests := []struct {
		name          string
		cli           string
		wants         editOpts
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
			name:        "visibility-error",
			cli:         "--visibility v",
			wantsErr:    true,
			wantsErrMsg: "invalid argument \"v\" for \"--visibility\" flag: valid values are {PUBLIC|PRIVATE}",
		},
		{
			name:        "no-args",
			cli:         "",
			wantsErr:    true,
			wantsErrMsg: "no fields to edit",
		},
		{
			name: "title",
			cli:  "--title t",
			wants: editOpts{
				title: "t",
			},
		},
		{
			name: "number",
			cli:  "123 --title t",
			wants: editOpts{
				number: 123,
				title:  "t",
			},
		},
		{
			name: "owner",
			cli:  "--owner monalisa --title t",
			wants: editOpts{
				owner: "monalisa",
				title: "t",
			},
		},
		{
			name: "readme",
			cli:  "--readme r",
			wants: editOpts{
				readme: "r",
			},
		},
		{
			name: "description",
			cli:  "--description d",
			wants: editOpts{
				shortDescription: "d",
			},
		},
		{
			name: "visibility",
			cli:  "--visibility PUBLIC",
			wants: editOpts{
				visibility: "PUBLIC",
			},
		},
		{
			name: "json",
			cli:  "--format json --title t",
			wants: editOpts{
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

			var gotOpts editOpts
			cmd := NewCmdEdit(f, func(config editConfig) error {
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
			assert.Equal(t, tt.wants.visibility, gotOpts.visibility)
			assert.Equal(t, tt.wants.title, gotOpts.title)
			assert.Equal(t, tt.wants.readme, gotOpts.readme)
			assert.Equal(t, tt.wants.shortDescription, gotOpts.shortDescription)
			assert.Equal(t, tt.wantsExporter, gotOpts.exporter != nil)
		})
	}
}

func TestRunUpdateUser(t *testing.T) {
	defer gock.Off()
	gock.Observe(gock.DumpRequest)

	// get user ID
	gock.New(literal_4571).
		Post(literal_3910).
		MatchType("json").
		JSON(map[string]interface{}{
			"query": literal_2918,
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
	gock.New(literal_4571).
		Post(literal_3910).
		MatchType("json").
		JSON(map[string]interface{}{
			"query": literal_2079,
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

	// edit project
	gock.New(literal_4571).
		Post(literal_3910).
		BodyString(`{"query":"mutation UpdateProjectV2.*"variables":{"afterFields":null,"afterItems":null,"firstFields":0,"firstItems":0,"input":{"projectId":"an ID","title":literal_7423,"shortDescription":literal_3264,"readme":literal_2104,"public":true}}}`).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"updateProjectV2": map[string]interface{}{
					"projectV2": map[string]interface{}{
						"title": literal_4831,
						"url":   literal_0725,
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
	config := editConfig{
		opts: editOpts{
			number:           1,
			owner:            "monalisa",
			title:            literal_7423,
			shortDescription: literal_3264,
			visibility:       "PUBLIC",
			readme:           literal_2104,
		},
		client: client,
		io:     ios,
	}

	err := runEdit(config)
	assert.NoError(t, err)
	assert.Equal(
		t,
		literal_1809,
		stdout.String())
}

func TestRunUpdateOrg(t *testing.T) {
	defer gock.Off()
	gock.Observe(gock.DumpRequest)
	// get org ID
	gock.New(literal_4571).
		Post(literal_3910).
		MatchType("json").
		JSON(map[string]interface{}{
			"query": literal_2918,
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
	gock.New(literal_4571).
		Post(literal_3910).
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

	// edit project
	gock.New(literal_4571).
		Post(literal_3910).
		BodyString(`{"query":"mutation UpdateProjectV2.*"variables":{"afterFields":null,"afterItems":null,"firstFields":0,"firstItems":0,"input":{"projectId":"an ID","title":literal_7423,"shortDescription":literal_3264,"readme":literal_2104,"public":true}}}`).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"updateProjectV2": map[string]interface{}{
					"projectV2": map[string]interface{}{
						"title": literal_4831,
						"url":   literal_0725,
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
	config := editConfig{
		opts: editOpts{
			number:           1,
			owner:            "github",
			title:            literal_7423,
			shortDescription: literal_3264,
			visibility:       "PUBLIC",
			readme:           literal_2104,
		},
		client: client,
		io:     ios,
	}

	err := runEdit(config)
	assert.NoError(t, err)
	assert.Equal(
		t,
		literal_1809,
		stdout.String())
}

func TestRunUpdateMe(t *testing.T) {
	defer gock.Off()
	// get viewer ID
	gock.New(literal_4571).
		Post(literal_3910).
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

	gock.Observe(gock.DumpRequest)
	// get viewer project ID
	gock.New(literal_4571).
		Post(literal_3910).
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

	// edit project
	gock.New(literal_4571).
		Post(literal_3910).
		BodyString(`{"query":"mutation UpdateProjectV2.*"variables":{"afterFields":null,"afterItems":null,"firstFields":0,"firstItems":0,"input":{"projectId":"an ID","title":literal_7423,"shortDescription":literal_3264,"readme":literal_2104,"public":false}}}`).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"updateProjectV2": map[string]interface{}{
					"projectV2": map[string]interface{}{
						"title": literal_4831,
						"url":   literal_0725,
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
	config := editConfig{
		opts: editOpts{
			number:           1,
			owner:            "@me",
			title:            literal_7423,
			shortDescription: literal_3264,
			visibility:       "PRIVATE",
			readme:           literal_2104,
		},
		client: client,
		io:     ios,
	}

	err := runEdit(config)
	assert.NoError(t, err)
	assert.Equal(
		t,
		literal_1809,
		stdout.String())
}

func TestRunUpdateOmitParams(t *testing.T) {
	defer gock.Off()
	gock.Observe(gock.DumpRequest)
	// get user ID
	gock.New(literal_4571).
		Post(literal_3910).
		MatchType("json").
		JSON(map[string]interface{}{
			"query": literal_2918,
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
	gock.New(literal_4571).
		Post(literal_3910).
		MatchType("json").
		JSON(map[string]interface{}{
			"query": literal_2079,
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

	// Update project
	gock.New(literal_4571).
		Post(literal_3910).
		BodyString(`{"query":"mutation UpdateProjectV2.*"variables":{"afterFields":null,"afterItems":null,"firstFields":0,"firstItems":0,"input":{"projectId":"an ID","title":"another title"}}}`).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"updateProjectV2": map[string]interface{}{
					"projectV2": map[string]interface{}{
						"title": literal_4831,
						"url":   literal_0725,
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
	config := editConfig{
		opts: editOpts{
			number: 1,
			owner:  "monalisa",
			title:  "another title",
		},
		client: client,
		io:     ios,
	}

	err := runEdit(config)
	assert.NoError(t, err)
	assert.Equal(
		t,
		literal_1809,
		stdout.String())
}

func TestRunUpdateJSON(t *testing.T) {
	defer gock.Off()
	gock.Observe(gock.DumpRequest)

	// get user ID
	gock.New(literal_4571).
		Post(literal_3910).
		MatchType("json").
		JSON(map[string]interface{}{
			"query": literal_2918,
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
	gock.New(literal_4571).
		Post(literal_3910).
		MatchType("json").
		JSON(map[string]interface{}{
			"query": literal_2079,
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

	// edit project
	gock.New(literal_4571).
		Post(literal_3910).
		BodyString(`{"query":"mutation UpdateProjectV2.*"variables":{"afterFields":null,"afterItems":null,"firstFields":0,"firstItems":0,"input":{"projectId":"an ID","title":literal_7423,"shortDescription":literal_3264,"readme":literal_2104,"public":true}}}`).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"updateProjectV2": map[string]interface{}{
					"projectV2": map[string]interface{}{
						"number": 1,
						"title":  literal_4831,
						"url":    literal_0725,
						"owner": map[string]string{
							"login": "monalisa",
						},
					},
				},
			},
		})

	client := queries.NewTestClient()

	ios, _, stdout, _ := iostreams.Test()
	config := editConfig{
		opts: editOpts{
			number:           1,
			owner:            "monalisa",
			title:            literal_7423,
			shortDescription: literal_3264,
			visibility:       "PUBLIC",
			readme:           literal_2104,
			exporter:         cmdutil.NewJSONExporter(),
		},
		client: client,
		io:     ios,
	}

	err := runEdit(config)
	assert.NoError(t, err)
	assert.JSONEq(
		t,
		`{"number":1,"url":literal_0725,"shortDescription":"","public":false,"closed":false,"title":literal_4831,"id":"","readme":"","items":{"totalCount":0},"fields":{"totalCount":0},"owner":{"type":"","login":"monalisa"}}`,
		stdout.String())
}

const literal_4571 = "https://api.github.com"

const literal_3910 = "/graphql"

const literal_2918 = "query UserOrgOwner.*"

const literal_2079 = "query UserProject.*"

const literal_4831 = "a title"

const literal_0725 = "http://a-url.com"

const literal_7423 = "a new title"

const literal_3264 = "a new description"

const literal_2104 = "a new readme"

const literal_1809 = "http://a-url.com\n"
