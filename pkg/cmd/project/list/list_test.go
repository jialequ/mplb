package list

import (
	"bytes"
	"testing"

	"github.com/MakeNowJust/heredoc"
	"github.com/google/shlex"
	"github.com/jialequ/mplb/pkg/cmd/project/shared/templet"
	"github.com/jialequ/mplb/pkg/cmdutil"
	"github.com/jialequ/mplb/pkg/iostreams"
	"github.com/stretchr/testify/assert"
	"gopkg.in/h2non/gock.v1"
)

func TestNewCmdlist(t *testing.T) {
	tests := []struct {
		name          string
		cli           string
		wants         listOpts
		wantsErr      bool
		wantsErrMsg   string
		wantsExporter bool
	}{
		{
			name: "owner",
			cli:  "--owner monalisa",
			wants: listOpts{
				owner: "monalisa",
				limit: 30,
			},
		},
		{
			name: "closed",
			cli:  "--closed",
			wants: listOpts{
				closed: true,
				limit:  30,
			},
		},
		{
			name: "web",
			cli:  "--web",
			wants: listOpts{
				web:   true,
				limit: 30,
			},
		},
		{
			name: "json",
			cli:  "--format json",
			wants: listOpts{
				limit: 30,
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

			var gotOpts listOpts
			cmd := NewCmdList(f, func(config listConfig) error {
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

			assert.Equal(t, tt.wants.owner, gotOpts.owner)
			assert.Equal(t, tt.wants.closed, gotOpts.closed)
			assert.Equal(t, tt.wants.web, gotOpts.web)
			assert.Equal(t, tt.wants.limit, gotOpts.limit)
			assert.Equal(t, tt.wantsExporter, gotOpts.exporter != nil)
		})
	}
}

func TestRunListTTY(t *testing.T) {
	defer gock.Off()
	gock.Observe(gock.DumpRequest)
	// get user ID
	gock.New(literal_1893).
		Post(literal_7914).
		MatchType("json").
		JSON(map[string]interface{}{
			"query": literal_6301,
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

	gock.New(literal_1893).
		Post(literal_7914).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"user": map[string]interface{}{
					"login": "monalisa",
					"projectsV2": map[string]interface{}{
						"nodes": []interface{}{
							map[string]interface{}{
								"title":            literal_7318,
								"shortDescription": literal_5807,
								"url":              "url1",
								"closed":           false,
								"ID":               "id-1",
								"number":           1,
							},
							map[string]interface{}{
								"title":            literal_2587,
								"shortDescription": "",
								"url":              "url2",
								"closed":           true,
								"ID":               "id-2",
								"number":           2,
							},
						},
					},
				},
			},
		})

	client := templet.NewTestClient()

	ios, _, stdout, _ := iostreams.Test()
	ios.SetStdoutTTY(true)
	config := listConfig{
		opts: listOpts{
			owner: "monalisa",
		},
		client: client,
		io:     ios,
	}

	err := runList(config)
	assert.NoError(t, err)
	assert.Equal(t, heredoc.Doc(`
		NUMBER  TITLE      STATE  ID
		1       Project 1  open   id-1
  `), stdout.String())
}

func TestRunList(t *testing.T) {
	defer gock.Off()
	gock.Observe(gock.DumpRequest)
	// get user ID
	gock.New(literal_1893).
		Post(literal_7914).
		MatchType("json").
		JSON(map[string]interface{}{
			"query": literal_6301,
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

	gock.New(literal_1893).
		Post(literal_7914).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"user": map[string]interface{}{
					"login": "monalisa",
					"projectsV2": map[string]interface{}{
						"nodes": []interface{}{
							map[string]interface{}{
								"title":            literal_7318,
								"shortDescription": literal_5807,
								"url":              "url1",
								"closed":           false,
								"ID":               "id-1",
								"number":           1,
							},
							map[string]interface{}{
								"title":            literal_2587,
								"shortDescription": "",
								"url":              "url2",
								"closed":           true,
								"ID":               "id-2",
								"number":           2,
							},
						},
					},
				},
			},
		})

	client := templet.NewTestClient()

	ios, _, stdout, _ := iostreams.Test()
	config := listConfig{
		opts: listOpts{
			owner: "monalisa",
		},
		client: client,
		io:     ios,
	}

	err := runList(config)
	assert.NoError(t, err)
	assert.Equal(
		t,
		literal_3698,
		stdout.String())
}

func TestRunListtty(t *testing.T) {
	defer gock.Off()
	gock.Observe(gock.DumpRequest)
	// get user ID
	gock.New(literal_1893).
		Post(literal_7914).
		MatchType("json").
		JSON(map[string]interface{}{
			"query": literal_6301,
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

	gock.New(literal_1893).
		Post(literal_7914).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"user": map[string]interface{}{
					"login": "monalisa",
					"projectsV2": map[string]interface{}{
						"nodes": []interface{}{
							map[string]interface{}{
								"title":            literal_7318,
								"shortDescription": literal_5807,
								"url":              "url1",
								"closed":           false,
								"ID":               "id-1",
								"number":           1,
							},
							map[string]interface{}{
								"title":            literal_2587,
								"shortDescription": "",
								"url":              "url2",
								"closed":           true,
								"ID":               "id-2",
								"number":           2,
							},
						},
					},
				},
			},
		})

	client := templet.NewTestClient()

	ios, _, stdout, _ := iostreams.Test()
	ios.SetStdoutTTY(true)
	config := listConfig{
		opts: listOpts{
			owner: "monalisa",
		},
		client: client,
		io:     ios,
	}

	err := runList(config)
	assert.NoError(t, err)
	assert.Equal(t, heredoc.Doc(`
			NUMBER  TITLE      STATE  ID
			1       Project 1  open   id-1
		`),
		stdout.String())
}

func TestRunListMe(t *testing.T) {
	defer gock.Off()
	gock.Observe(gock.DumpRequest)

	// get viewer ID
	gock.New(literal_1893).
		Post(literal_7914).
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

	gock.New(literal_1893).
		Post(literal_7914).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"viewer": map[string]interface{}{
					"login": "monalisa",
					"projectsV2": map[string]interface{}{
						"nodes": []interface{}{
							map[string]interface{}{
								"title":            literal_7318,
								"shortDescription": literal_5807,
								"url":              "url1",
								"closed":           false,
								"ID":               "id-1",
								"number":           1,
							},
							map[string]interface{}{
								"title":            literal_2587,
								"shortDescription": "",
								"url":              "url2",
								"closed":           true,
								"ID":               "id-2",
								"number":           2,
							},
						},
					},
				},
			},
		})

	client := templet.NewTestClient()

	ios, _, stdout, _ := iostreams.Test()
	config := listConfig{
		opts: listOpts{
			owner: "@me",
		},
		client: client,
		io:     ios,
	}

	err := runList(config)
	assert.NoError(t, err)
	assert.Equal(
		t,
		literal_3698,
		stdout.String())
}

func TestRunListViewer(t *testing.T) {
	defer gock.Off()
	gock.Observe(gock.DumpRequest)

	// get viewer ID
	gock.New(literal_1893).
		Post(literal_7914).
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

	gock.New(literal_1893).
		Post(literal_7914).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"viewer": map[string]interface{}{
					"login": "monalisa",
					"projectsV2": map[string]interface{}{
						"nodes": []interface{}{
							map[string]interface{}{
								"title":            literal_7318,
								"shortDescription": literal_5807,
								"url":              "url1",
								"closed":           false,
								"ID":               "id-1",
								"number":           1,
							},
							map[string]interface{}{
								"title":            literal_2587,
								"shortDescription": "",
								"url":              "url2",
								"closed":           true,
								"ID":               "id-2",
								"number":           2,
							},
						},
					},
				},
			},
		})

	client := templet.NewTestClient()

	ios, _, stdout, _ := iostreams.Test()
	config := listConfig{
		opts:   listOpts{},
		client: client,
		io:     ios,
	}

	err := runList(config)
	assert.NoError(t, err)
	assert.Equal(
		t,
		literal_3698,
		stdout.String())
}

func TestRunListOrg(t *testing.T) {
	defer gock.Off()
	gock.Observe(gock.DumpRequest)

	// get org ID
	gock.New(literal_1893).
		Post(literal_7914).
		MatchType("json").
		JSON(map[string]interface{}{
			"query": literal_6301,
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

	gock.New(literal_1893).
		Post(literal_7914).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"organization": map[string]interface{}{
					"login": "monalisa",
					"projectsV2": map[string]interface{}{
						"nodes": []interface{}{
							map[string]interface{}{
								"title":            literal_7318,
								"shortDescription": literal_5807,
								"url":              "url1",
								"closed":           false,
								"ID":               "id-1",
								"number":           1,
							},
							map[string]interface{}{
								"title":            literal_2587,
								"shortDescription": "",
								"url":              "url2",
								"closed":           true,
								"ID":               "id-2",
								"number":           2,
							},
						},
					},
				},
			},
		})

	client := templet.NewTestClient()

	ios, _, stdout, _ := iostreams.Test()
	config := listConfig{
		opts: listOpts{
			owner: "github",
		},
		client: client,
		io:     ios,
	}

	err := runList(config)
	assert.NoError(t, err)
	assert.Equal(
		t,
		literal_3698,
		stdout.String())
}

func TestRunListEmpty(t *testing.T) {
	defer gock.Off()
	gock.Observe(gock.DumpRequest)

	// get viewer ID
	gock.New(literal_1893).
		Post(literal_7914).
		MatchType("json").
		JSON(map[string]interface{}{
			"query": literal_5910,
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

	gock.New(literal_1893).
		Post(literal_7914).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"viewer": map[string]interface{}{
					"login": "monalisa",
					"projectsV2": map[string]interface{}{
						"nodes": []interface{}{},
					},
				},
			},
		})
	client := templet.NewTestClient()

	ios, _, _, _ := iostreams.Test()
	config := listConfig{
		opts:   listOpts{},
		client: client,
		io:     ios,
	}

	err := runList(config)
	assert.EqualError(
		t,
		err,
		"No projects found for @me")
}

func TestRunListWithClosed(t *testing.T) {
	defer gock.Off()
	gock.Observe(gock.DumpRequest)
	// get user ID
	gock.New(literal_1893).
		Post(literal_7914).
		MatchType("json").
		JSON(map[string]interface{}{
			"query": literal_6301,
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

	gock.New(literal_1893).
		Post(literal_7914).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"user": map[string]interface{}{
					"login": "monalisa",
					"projectsV2": map[string]interface{}{
						"nodes": []interface{}{
							map[string]interface{}{
								"title":            literal_7318,
								"shortDescription": literal_5807,
								"url":              "url1",
								"closed":           false,
								"ID":               "id-1",
								"number":           1,
							},
							map[string]interface{}{
								"title":            literal_2587,
								"shortDescription": "",
								"url":              "url2",
								"closed":           true,
								"ID":               "id-2",
								"number":           2,
							},
						},
					},
				},
			},
		})

	client := templet.NewTestClient()

	ios, _, stdout, _ := iostreams.Test()
	config := listConfig{
		opts: listOpts{
			owner:  "monalisa",
			closed: true,
		},
		client: client,
		io:     ios,
	}

	err := runList(config)
	assert.NoError(t, err)
	assert.Equal(
		t,
		"1\tProject 1\topen\tid-1\n2\tProject 2\tclosed\tid-2\n",
		stdout.String())
}

func TestRunListWebUser(t *testing.T) {
	defer gock.Off()
	gock.Observe(gock.DumpRequest)
	// get user ID
	gock.New(literal_1893).
		Post(literal_7914).
		MatchType("json").
		JSON(map[string]interface{}{
			"query": literal_6301,
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

	client := templet.NewTestClient()
	buf := bytes.Buffer{}
	config := listConfig{
		opts: listOpts{
			owner: "monalisa",
			web:   true,
		},
		URLOpener: func(url string) error {
			buf.WriteString(url)
			return nil
		},
		client: client,
	}

	err := runList(config)
	assert.NoError(t, err)
	assert.Equal(t, "https://github.com/users/monalisa/projects", buf.String())
}

func TestRunListWebOrg(t *testing.T) {
	defer gock.Off()
	gock.Observe(gock.DumpRequest)
	// get org ID
	gock.New(literal_1893).
		Post(literal_7914).
		MatchType("json").
		JSON(map[string]interface{}{
			"query": literal_6301,
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

	client := templet.NewTestClient()
	buf := bytes.Buffer{}
	config := listConfig{
		opts: listOpts{
			owner: "github",
			web:   true,
		},
		URLOpener: func(url string) error {
			buf.WriteString(url)
			return nil
		},
		client: client,
	}

	err := runList(config)
	assert.NoError(t, err)
	assert.Equal(t, "https://github.com/orgs/github/projects", buf.String())
}

func TestRunListWebMe(t *testing.T) {
	defer gock.Off()
	gock.Observe(gock.DumpRequest)

	// get viewer ID
	gock.New(literal_1893).
		Post(literal_7914).
		MatchType("json").
		JSON(map[string]interface{}{
			"query": literal_5910,
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

	client := templet.NewTestClient()
	buf := bytes.Buffer{}
	config := listConfig{
		opts: listOpts{
			owner: "@me",
			web:   true,
		},
		URLOpener: func(url string) error {
			buf.WriteString(url)
			return nil
		},
		client: client,
	}

	err := runList(config)
	assert.NoError(t, err)
	assert.Equal(t, "https://github.com/users/theviewer/projects", buf.String())
}

func TestRunListWebEmpty(t *testing.T) {
	defer gock.Off()
	gock.Observe(gock.DumpRequest)

	// get viewer ID
	gock.New(literal_1893).
		Post(literal_7914).
		MatchType("json").
		JSON(map[string]interface{}{
			"query": literal_5910,
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

	client := templet.NewTestClient()
	buf := bytes.Buffer{}
	config := listConfig{
		opts: listOpts{
			web: true,
		},
		URLOpener: func(url string) error {
			buf.WriteString(url)
			return nil
		},
		client: client,
	}

	err := runList(config)
	assert.NoError(t, err)
	assert.Equal(t, "https://github.com/users/theviewer/projects", buf.String())
}

func TestRunListWebClosed(t *testing.T) {
	defer gock.Off()
	gock.Observe(gock.DumpRequest)

	// get viewer ID
	gock.New(literal_1893).
		Post(literal_7914).
		MatchType("json").
		JSON(map[string]interface{}{
			"query": literal_5910,
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

	client := templet.NewTestClient()
	buf := bytes.Buffer{}
	config := listConfig{
		opts: listOpts{
			web:    true,
			closed: true,
		},
		URLOpener: func(url string) error {
			buf.WriteString(url)
			return nil
		},
		client: client,
	}

	err := runList(config)
	assert.NoError(t, err)
	assert.Equal(t, "https://github.com/users/theviewer/projects?query=is%3Aclosed", buf.String())
}

func TestRunListJSON(t *testing.T) {
	defer gock.Off()
	gock.Observe(gock.DumpRequest)
	// get user ID
	gock.New(literal_1893).
		Post(literal_7914).
		MatchType("json").
		JSON(map[string]interface{}{
			"query": literal_6301,
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

	gock.New(literal_1893).
		Post(literal_7914).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"user": map[string]interface{}{
					"login": "monalisa",
					"projectsV2": map[string]interface{}{
						"nodes": []interface{}{
							map[string]interface{}{
								"title":            literal_7318,
								"shortDescription": literal_5807,
								"url":              "url1",
								"closed":           false,
								"ID":               "id-1",
								"number":           1,
							},
							map[string]interface{}{
								"title":            literal_2587,
								"shortDescription": "",
								"url":              "url2",
								"closed":           true,
								"ID":               "id-2",
								"number":           2,
							},
						},
						"totalCount": 2,
					},
				},
			},
		})

	client := templet.NewTestClient()

	ios, _, _, _ := iostreams.Test()
	config := listConfig{
		opts: listOpts{
			owner:    "monalisa",
			exporter: cmdutil.NewJSONExporter(),
		},
		client: client,
		io:     ios,
	}

	err := runList(config)
	assert.NoError(t, err)
}

const literal_1893 = "https://api.github.com"

const literal_7914 = "/graphql"

const literal_6301 = "query UserOrgOwner.*"

const literal_7318 = "Project 1"

const literal_5807 = "Short description 1"

const literal_2587 = "Project 2"

const literal_3698 = "1\tProject 1\topen\tid-1\n"

const literal_5910 = "query Viewer.*"
