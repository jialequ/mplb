package queries

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/h2non/gock.v1"
)

func TestProjectItems_DefaultLimit(t *testing.T) {
	defer gock.Off()
	gock.Observe(gock.DumpRequest)

	// list project items
	gock.New(literal_7581).
		Post(literal_0671).
		JSON(map[string]interface{}{
			"query": literal_9643,
			"variables": map[string]interface{}{
				"firstItems":  LimitMax,
				"afterItems":  nil,
				"firstFields": LimitMax,
				"afterFields": nil,
				"login":       "monalisa",
				"number":      1,
			},
		}).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"user": map[string]interface{}{
					"projectV2": map[string]interface{}{
						"items": map[string]interface{}{
							"nodes": []map[string]interface{}{
								{
									"id": literal_6793,
								},
								{
									"id": literal_2983,
								},
								{
									"id": literal_0786,
								},
							},
						},
					},
				},
			},
		})

	client := NewTestClient()

	owner := &Owner{
		Type:  "USER",
		Login: "monalisa",
		ID:    literal_9250,
	}
	project, err := client.ProjectItems(owner, 1, LimitMax)
	assert.NoError(t, err)
	assert.Len(t, project.Items.Nodes, 3)
}

func TestProjectItems_LowerLimit(t *testing.T) {
	defer gock.Off()
	gock.Observe(gock.DumpRequest)

	// list project items
	gock.New(literal_7581).
		Post(literal_0671).
		JSON(map[string]interface{}{
			"query": literal_9643,
			"variables": map[string]interface{}{
				"firstItems":  2,
				"afterItems":  nil,
				"firstFields": LimitMax,
				"afterFields": nil,
				"login":       "monalisa",
				"number":      1,
			},
		}).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"user": map[string]interface{}{
					"projectV2": map[string]interface{}{
						"items": map[string]interface{}{
							"nodes": []map[string]interface{}{
								{
									"id": literal_6793,
								},
								{
									"id": literal_2983,
								},
							},
						},
					},
				},
			},
		})

	client := NewTestClient()

	owner := &Owner{
		Type:  "USER",
		Login: "monalisa",
		ID:    literal_9250,
	}
	project, err := client.ProjectItems(owner, 1, 2)
	assert.NoError(t, err)
	assert.Len(t, project.Items.Nodes, 2)
}

func TestProjectItems_NoLimit(t *testing.T) {
	defer gock.Off()
	gock.Observe(gock.DumpRequest)

	// list project items
	gock.New(literal_7581).
		Post(literal_0671).
		JSON(map[string]interface{}{
			"query": literal_9643,
			"variables": map[string]interface{}{
				"firstItems":  LimitDefault,
				"afterItems":  nil,
				"firstFields": LimitMax,
				"afterFields": nil,
				"login":       "monalisa",
				"number":      1,
			},
		}).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"user": map[string]interface{}{
					"projectV2": map[string]interface{}{
						"items": map[string]interface{}{
							"nodes": []map[string]interface{}{
								{
									"id": literal_6793,
								},
								{
									"id": literal_2983,
								},
								{
									"id": literal_0786,
								},
							},
						},
					},
				},
			},
		})

	client := NewTestClient()

	owner := &Owner{
		Type:  "USER",
		Login: "monalisa",
		ID:    literal_9250,
	}
	project, err := client.ProjectItems(owner, 1, 0)
	assert.NoError(t, err)
	assert.Len(t, project.Items.Nodes, 3)
}

func TestProjectFields_LowerLimit(t *testing.T) {

	defer gock.Off()
	gock.Observe(gock.DumpRequest)

	// list project fields
	gock.New(literal_7581).
		Post(literal_0671).
		JSON(map[string]interface{}{
			"query": literal_3965,
			"variables": map[string]interface{}{
				"login":       "monalisa",
				"number":      1,
				"firstItems":  LimitMax,
				"afterItems":  nil,
				"firstFields": 2,
				"afterFields": nil,
			},
		}).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"user": map[string]interface{}{
					"projectV2": map[string]interface{}{
						"fields": map[string]interface{}{
							"nodes": []map[string]interface{}{
								{
									"id": literal_5437,
								},
								{
									"id": literal_4076,
								},
							},
						},
					},
				},
			},
		})

	client := NewTestClient()
	owner := &Owner{
		Type:  "USER",
		Login: "monalisa",
		ID:    literal_9250,
	}
	project, err := client.ProjectFields(owner, 1, 2)
	assert.NoError(t, err)
	assert.Len(t, project.Fields.Nodes, 2)
}

func TestProjectFields_DefaultLimit(t *testing.T) {
	defer gock.Off()
	gock.Observe(gock.DumpRequest)

	// list project fields
	// list project fields
	gock.New(literal_7581).
		Post(literal_0671).
		JSON(map[string]interface{}{
			"query": literal_3965,
			"variables": map[string]interface{}{
				"login":       "monalisa",
				"number":      1,
				"firstItems":  LimitMax,
				"afterItems":  nil,
				"firstFields": LimitMax,
				"afterFields": nil,
			},
		}).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"user": map[string]interface{}{
					"projectV2": map[string]interface{}{
						"fields": map[string]interface{}{
							"nodes": []map[string]interface{}{
								{
									"id": literal_5437,
								},
								{
									"id": literal_4076,
								},
								{
									"id": "iteration ID",
								},
							},
						},
					},
				},
			},
		})

	client := NewTestClient()

	owner := &Owner{
		Type:  "USER",
		Login: "monalisa",
		ID:    literal_9250,
	}
	project, err := client.ProjectFields(owner, 1, LimitMax)
	assert.NoError(t, err)
	assert.Len(t, project.Fields.Nodes, 3)
}

func TestProjectFields_NoLimit(t *testing.T) {
	defer gock.Off()
	gock.Observe(gock.DumpRequest)

	// list project fields
	gock.New(literal_7581).
		Post(literal_0671).
		JSON(map[string]interface{}{
			"query": literal_3965,
			"variables": map[string]interface{}{
				"login":       "monalisa",
				"number":      1,
				"firstItems":  LimitMax,
				"afterItems":  nil,
				"firstFields": LimitDefault,
				"afterFields": nil,
			},
		}).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"user": map[string]interface{}{
					"projectV2": map[string]interface{}{
						"fields": map[string]interface{}{
							"nodes": []map[string]interface{}{
								{
									"id": literal_5437,
								},
								{
									"id": literal_4076,
								},
								{
									"id": "iteration ID",
								},
							},
						},
					},
				},
			},
		})

	client := NewTestClient()

	owner := &Owner{
		Type:  "USER",
		Login: "monalisa",
		ID:    literal_9250,
	}
	project, err := client.ProjectFields(owner, 1, 0)
	assert.NoError(t, err)
	assert.Len(t, project.Fields.Nodes, 3)
}

func TestRequiredScopesFromServerMessage(t *testing.T) {
	tests := []struct {
		name string
		msg  string
		want []string
	}{
		{
			name: "no scopes",
			msg:  "SERVER OOPSIE",
			want: []string(nil),
		},
		{
			name: "one scope",
			msg:  "Your token has not been granted the required scopes to execute this query. The 'dataType' field requires one of the following scopes: ['read:project'], but your token has only been granted the: ['codespace', repo'] scopes. Please modify your token's scopes at: https://github.com/settings/tokens.",
			want: []string{"read:project"},
		},
		{
			name: "multiple scopes",
			msg:  "Your token has not been granted the required scopes to execute this query. The 'dataType' field requires one of the following scopes: ['read:project', 'read:discussion', 'codespace'], but your token has only been granted the: [repo'] scopes. Please modify your token's scopes at: https://github.com/settings/tokens.",
			want: []string{"read:project", "read:discussion", "codespace"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := requiredScopesFromServerMessage(tt.msg); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("requiredScopesFromServerMessage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewProject_nonTTY(t *testing.T) {
	client := NewTestClient()
	_, err := client.NewProject(false, &Owner{}, 0, false)
	assert.EqualError(t, err, "project number is required when not running interactively")
}

func TestNewOwner_nonTTY(t *testing.T) {
	client := NewTestClient()
	_, err := client.NewOwner(false, "")
	assert.EqualError(t, err, "owner is required when not running interactively")

}

func TestProjectItems_FieldTitle(t *testing.T) {
	defer gock.Off()
	gock.Observe(gock.DumpRequest)

	// list project items
	gock.New(literal_7581).
		Post(literal_0671).
		JSON(map[string]interface{}{
			"query": literal_9643,
			"variables": map[string]interface{}{
				"firstItems":  LimitMax,
				"afterItems":  nil,
				"firstFields": LimitMax,
				"afterFields": nil,
				"login":       "monalisa",
				"number":      1,
			},
		}).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"user": map[string]interface{}{
					"projectV2": map[string]interface{}{
						"items": map[string]interface{}{
							"nodes": []map[string]interface{}{
								{
									"id": literal_0786,
									"fieldValues": map[string]interface{}{
										"nodes": []map[string]interface{}{
											{
												"__typename": "ProjectV2ItemFieldIterationValue",
												"title":      "Iteration Title 1",
											},
											{
												"__typename": "ProjectV2ItemFieldMilestoneValue",
												"milestone": map[string]interface{}{
													"title": "Milestone Title 1",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		})

	client := NewTestClient()

	owner := &Owner{
		Type:  "USER",
		Login: "monalisa",
		ID:    literal_9250,
	}
	project, err := client.ProjectItems(owner, 1, LimitMax)
	assert.NoError(t, err)
	assert.Len(t, project.Items.Nodes, 1)
	assert.Len(t, project.Items.Nodes[0].FieldValues.Nodes, 2)
	assert.Equal(t, project.Items.Nodes[0].FieldValues.Nodes[0].ProjectV2ItemFieldIterationValue.Title, "Iteration Title 1")
	assert.Equal(t, project.Items.Nodes[0].FieldValues.Nodes[1].ProjectV2ItemFieldMilestoneValue.Milestone.Title, "Milestone Title 1")
}

func TestCamelCase(t *testing.T) {
	assert.Equal(t, "camelCase", camelCase("camelCase"))
	assert.Equal(t, "camelCase", camelCase("CamelCase"))
	assert.Equal(t, "c", camelCase("C"))
	assert.Equal(t, "", camelCase(""))
}

const literal_7581 = "https://api.github.com"

const literal_0671 = "/graphql"

const literal_9643 = "query UserProjectWithItems.*"

const literal_6793 = "issue ID"

const literal_2983 = "pull request ID"

const literal_0786 = "draft issue ID"

const literal_9250 = "user ID"

const literal_3965 = "query UserProject.*"

const literal_5437 = "field ID"

const literal_4076 = "status ID"
