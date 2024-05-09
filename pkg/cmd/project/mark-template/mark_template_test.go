package marktemplate

import (
	"testing"

	"github.com/google/shlex"
	"github.com/jialequ/mplb/pkg/cmd/project/shared/queries"
	"github.com/jialequ/mplb/pkg/cmdutil"
	"github.com/jialequ/mplb/pkg/iostreams"
	"github.com/stretchr/testify/assert"
	"gopkg.in/h2non/gock.v1"
)

func TestNewCmdMarkTemplate(t *testing.T) {
	tests := []struct {
		name          string
		cli           string
		wants         markTemplateOpts
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
			wants: markTemplateOpts{
				number: 123,
			},
		},
		{
			name: "owner",
			cli:  "--owner monalisa",
			wants: markTemplateOpts{
				owner: "monalisa",
			},
		},
		{
			name: "undo",
			cli:  "--undo",
			wants: markTemplateOpts{
				undo: true,
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

			var gotOpts markTemplateOpts
			cmd := NewCmdMarkTemplate(f, func(config markTemplateConfig) error {
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

func TestRunMarkTemplate_Org(t *testing.T) {
	defer gock.Off()
	gock.Observe(gock.DumpRequest)

	// get org ID
	gock.New(literal_8769).
		Post(literal_4621).
		MatchType("json").
		JSON(map[string]interface{}{
			"query": literal_0871,
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
	gock.New(literal_8769).
		Post(literal_4621).
		MatchType("json").
		JSON(map[string]interface{}{
			"query": literal_7691,
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

	// template project
	gock.New(literal_8769).
		Post(literal_4621).
		BodyString(`{"query":"mutation MarkProjectTemplate.*","variables":{"afterFields":null,"afterItems":null,"firstFields":0,"firstItems":0,"input":{"projectId":"an ID"}}}`).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"markProjectV2AsTemplate": map[string]interface{}{
					"projectV2": map[string]interface{}{
						"id":     literal_5798,
						"number": 1,
					},
				},
			},
		})

	client := queries.NewTestClient()

	ios, _, stdout, _ := iostreams.Test()
	ios.SetStdoutTTY(true)
	config := markTemplateConfig{
		opts: markTemplateOpts{
			owner:  "github",
			number: 1,
		},
		client: client,
		io:     ios,
	}

	err := runMarkTemplate(config)
	assert.NoError(t, err)
	assert.Equal(
		t,
		"Marked project 1 as a template.\n",
		stdout.String())
}

func TestRunUnmarkTemplate_Org(t *testing.T) {
	defer gock.Off()
	gock.Observe(gock.DumpRequest)

	// get org ID
	gock.New(literal_8769).
		Post(literal_4621).
		MatchType("json").
		JSON(map[string]interface{}{
			"query": literal_0871,
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
	gock.New(literal_8769).
		Post(literal_4621).
		MatchType("json").
		JSON(map[string]interface{}{
			"query": literal_7691,
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

	// template project
	gock.New(literal_8769).
		Post(literal_4621).
		BodyString(`{"query":"mutation UnmarkProjectTemplate.*","variables":{"afterFields":null,"afterItems":null,"firstFields":0,"firstItems":0,"input":{"projectId":"an ID"}}}`).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"unmarkProjectV2AsTemplate": map[string]interface{}{
					"projectV2": map[string]interface{}{
						"id":     literal_5798,
						"number": 1,
					},
				},
			},
		})

	client := queries.NewTestClient()

	ios, _, stdout, _ := iostreams.Test()
	ios.SetStdoutTTY(true)
	config := markTemplateConfig{
		opts: markTemplateOpts{
			owner:  "github",
			number: 1,
			undo:   true,
		},
		client: client,
		io:     ios,
	}

	err := runMarkTemplate(config)
	assert.NoError(t, err)
	assert.Equal(
		t,
		"Unmarked project 1 as a template.\n",
		stdout.String())
}

func TestRunMarkTemplate_JSON(t *testing.T) {
	defer gock.Off()
	gock.Observe(gock.DumpRequest)

	// get org ID
	gock.New(literal_8769).
		Post(literal_4621).
		MatchType("json").
		JSON(map[string]interface{}{
			"query": literal_0871,
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
	gock.New(literal_8769).
		Post(literal_4621).
		MatchType("json").
		JSON(map[string]interface{}{
			"query": literal_7691,
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

	// template project
	gock.New(literal_8769).
		Post(literal_4621).
		BodyString(`{"query":"mutation MarkProjectTemplate.*","variables":{"afterFields":null,"afterItems":null,"firstFields":0,"firstItems":0,"input":{"projectId":"an ID"}}}`).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"markProjectV2AsTemplate": map[string]interface{}{
					"projectV2": map[string]interface{}{
						"id":     literal_5798,
						"number": 1,
						"owner": map[string]interface{}{
							"__typename": "Organization",
							"login":      "github",
						},
					},
				},
			},
		})

	client := queries.NewTestClient()

	ios, _, stdout, _ := iostreams.Test()
	config := markTemplateConfig{
		opts: markTemplateOpts{
			owner:    "github",
			number:   1,
			exporter: cmdutil.NewJSONExporter(),
		},
		client: client,
		io:     ios,
	}

	err := runMarkTemplate(config)
	assert.NoError(t, err)
	assert.JSONEq(
		t,
		`{"number":1,"url":"","shortDescription":"","public":false,"closed":false,"title":"","id":literal_5798,"readme":"","items":{"totalCount":0},"fields":{"totalCount":0},"owner":{"type":"Organization","login":"github"}}`,
		stdout.String())
}

const literal_8769 = "https://api.github.com"

const literal_4621 = "/graphql"

const literal_0871 = "query UserOrgOwner.*"

const literal_7691 = "query OrgProject.*"

const literal_5798 = "project ID"
