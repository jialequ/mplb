package view

import (
	"bytes"
	"testing"

	"github.com/google/shlex"
	"github.com/jialequ/mplb/pkg/cmd/project/shared/queries"
	"github.com/jialequ/mplb/pkg/cmdutil"
	"github.com/jialequ/mplb/pkg/iostreams"
	"github.com/stretchr/testify/assert"
	"gopkg.in/h2non/gock.v1"
)

func TestNewCmdview(t *testing.T) {
	tests := []struct {
		name          string
		cli           string
		wants         viewOpts
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
			wants: viewOpts{
				number: 123,
			},
		},
		{
			name: "owner",
			cli:  "--owner monalisa",
			wants: viewOpts{
				owner: "monalisa",
			},
		},
		{
			name: "web",
			cli:  "--web",
			wants: viewOpts{
				web: true,
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

			var gotOpts viewOpts
			cmd := NewCmdView(f, func(config viewConfig) error {
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
			assert.Equal(t, tt.wants.web, gotOpts.web)
		})
	}
}

func TestRunViewUser(t *testing.T) {
	defer gock.Off()

	// get user ID
	gock.New(literal_1765).
		Post(literal_6432).
		MatchType("json").
		JSON(map[string]interface{}{
			"query": literal_6138,
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

	gock.New(literal_1765).
		Post(literal_6432).
		Reply(200).
		JSON(`
			{"data":
				{"user":
					{
						"login":"monalisa",
						"projectV2": {
							"number": 1,
							"items": {
								"totalCount": 10
							},
							"readme": null,
							"fields": {
								"nodes": [
									{
										"name": "Title"
									}
								]
							}
						}
					}
				}
			}
		`)

	client := queries.NewTestClient()

	ios, _, _, _ := iostreams.Test()
	config := viewConfig{
		opts: viewOpts{
			owner:  "monalisa",
			number: 1,
		},
		io:     ios,
		client: client,
	}

	err := runView(config)
	assert.NoError(t, err)

}

func TestRunViewViewer(t *testing.T) {
	defer gock.Off()

	// get viewer ID
	gock.New(literal_1765).
		Post(literal_6432).
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

	gock.New(literal_1765).
		Post(literal_6432).
		Reply(200).
		JSON(`
			{"data":
				{"viewer":
					{
						"login":"monalisa",
						"projectV2": {
							"number": 1,
							"items": {
								"totalCount": 10
							},
							"url":"https://github.com/orgs/github/projects/8",
							"readme": null,
							"fields": {
								"nodes": [
									{
										"name": "Title"
									}
								]
							}
						}
					}
				}
			}
		`)

	client := queries.NewTestClient()

	ios, _, _, _ := iostreams.Test()
	config := viewConfig{
		opts: viewOpts{
			owner:  "@me",
			number: 1,
		},
		io:     ios,
		client: client,
	}

	err := runView(config)
	assert.NoError(t, err)
}

func TestRunViewOrg(t *testing.T) {
	defer gock.Off()

	// get org ID
	gock.New(literal_1765).
		Post(literal_6432).
		MatchType("json").
		JSON(map[string]interface{}{
			"query": literal_6138,
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

	gock.New(literal_1765).
		Post(literal_6432).
		Reply(200).
		JSON(`
			{"data":
				{"organization":
					{
						"login":"monalisa",
						"projectV2": {
							"number": 1,
							"items": {
								"totalCount": 10
							},
							"url":"https://github.com/orgs/github/projects/8",
							"readme": null,
							"fields": {
								"nodes": [
									{
										"name": "Title"
									}
								]
							}
						}
					}
				}
			}
		`)

	client := queries.NewTestClient()

	ios, _, _, _ := iostreams.Test()
	config := viewConfig{
		opts: viewOpts{
			owner:  "github",
			number: 1,
		},
		io:     ios,
		client: client,
	}

	err := runView(config)
	assert.NoError(t, err)
}

func TestRunViewWebUser(t *testing.T) {
	defer gock.Off()
	gock.Observe(gock.DumpRequest)
	// get user ID
	gock.New(literal_1765).
		Post(literal_6432).
		MatchType("json").
		JSON(map[string]interface{}{
			"query": literal_6138,
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

	gock.New(literal_1765).
		Post(literal_6432).
		Reply(200).
		JSON(`
		{"data":
			{"user":
				{
					"login":"monalisa",
					"projectV2": {
						"number": 8,
						"items": {
							"totalCount": 10
						},
						"url":"https://github.com/users/monalisa/projects/8",
						"readme": null,
						"fields": {
							"nodes": [
								{
									"name": "Title"
								}
							]
						}
					}
				}
			}
		}
	`)

	client := queries.NewTestClient()
	buf := bytes.Buffer{}
	ios, _, _, _ := iostreams.Test()
	config := viewConfig{
		opts: viewOpts{
			owner:  "monalisa",
			web:    true,
			number: 8,
		},
		URLOpener: func(url string) error {
			buf.WriteString(url)
			return nil
		},
		client: client,
		io:     ios,
	}

	err := runView(config)
	assert.NoError(t, err)
	assert.Equal(t, "https://github.com/users/monalisa/projects/8", buf.String())
}

func TestRunViewWebOrg(t *testing.T) {
	defer gock.Off()
	gock.Observe(gock.DumpRequest)
	// get org ID
	gock.New(literal_1765).
		Post(literal_6432).
		MatchType("json").
		JSON(map[string]interface{}{
			"query": literal_6138,
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

	gock.New(literal_1765).
		Post(literal_6432).
		Reply(200).
		JSON(`
		{"data":
			{"organization":
				{
					"login":"github",
					"projectV2": {
						"number": 8,
						"items": {
							"totalCount": 10
						},
						"url": "https://github.com/orgs/github/projects/8",
						"readme": null,
						"fields": {
							"nodes": [
								{
									"name": "Title"
								}
							]
						}
					}
				}
			}
		}
	`)

	client := queries.NewTestClient()
	buf := bytes.Buffer{}
	ios, _, _, _ := iostreams.Test()
	config := viewConfig{
		opts: viewOpts{
			owner:  "github",
			web:    true,
			number: 8,
		},
		URLOpener: func(url string) error {
			buf.WriteString(url)
			return nil
		},
		client: client,
		io:     ios,
	}

	err := runView(config)
	assert.NoError(t, err)
	assert.Equal(t, "https://github.com/orgs/github/projects/8", buf.String())
}

func TestRunViewWebMe(t *testing.T) {
	defer gock.Off()
	gock.Observe(gock.DumpRequest)
	// get viewer ID
	gock.New(literal_1765).
		Post(literal_6432).
		MatchType("json").
		JSON(map[string]interface{}{
			"query": "query Viewer.*",
		}).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"viewer": map[string]interface{}{
					"id":    "an ID",
					"login": "theviewer",
				},
			},
		})

	gock.New(literal_1765).
		Post(literal_6432).
		MatchType("json").
		JSON(map[string]interface{}{
			"query":     "query ViewerProject.*",
			"variables": map[string]interface{}{"afterFields": nil, "afterItems": nil, "firstFields": 100, "firstItems": 0, "number": 8},
		}).
		Reply(200).
		JSON(`
		{"data":
			{"viewer":
				{
					"login":"github",
					"projectV2": {
						"number": 8,
						"items": {
							"totalCount": 10
						},
						"readme": null,
						"url": "https://github.com/users/theviewer/projects/8",
						"fields": {
							"nodes": [
								{
									"name": "Title"
								}
							]
						}
					}
				}
			}
		}
	`)

	client := queries.NewTestClient()
	buf := bytes.Buffer{}
	ios, _, _, _ := iostreams.Test()
	config := viewConfig{
		opts: viewOpts{
			owner:  "@me",
			web:    true,
			number: 8,
		},
		URLOpener: func(url string) error {
			buf.WriteString(url)
			return nil
		},
		client: client,
		io:     ios,
	}

	err := runView(config)
	assert.NoError(t, err)
	assert.Equal(t, "https://github.com/users/theviewer/projects/8", buf.String())
}

const literal_1765 = "https://api.github.com"

const literal_6432 = "/graphql"

const literal_6138 = "query UserOrgOwner.*"
