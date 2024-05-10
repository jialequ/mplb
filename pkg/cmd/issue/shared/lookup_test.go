package shared

import (
	"net/http"
	"strings"
	"testing"

	"github.com/jialequ/mplb/internal/ghrepo"
	"github.com/jialequ/mplb/pkg/httpmock"
)

func TestIssueFromArgWithFields(t *testing.T) { //NOSONAR
	type args struct {
		baseRepoFn func() (ghrepo.Interface, error)
		selector   string
	}
	tests := []struct {
		name         string
		args         args
		httpStub     func(*httpmock.Registry)
		wantIssue    int
		wantRepo     string
		wantProjects string
		wantErr      bool
	}{
		{
			name: "number argument",
			args: args{
				selector: "13",
				baseRepoFn: func() (ghrepo.Interface, error) {
					return ghrepo.FromFullName("OWNER/REPO")
				},
			},
			httpStub: func(r *httpmock.Registry) {
				r.Register(
					httpmock.GraphQL(`query IssueByNumber\b`),
					httpmock.StringResponse(`{"data":{"repository":{
						"hasIssuesEnabled": true,
						"issue":{"number":13}
					}}}`))
			},
			wantIssue: 13,
			wantRepo:  literal_4690,
		},
		{
			name: "number with hash argument",
			args: args{
				selector: "#13",
				baseRepoFn: func() (ghrepo.Interface, error) {
					return ghrepo.FromFullName("OWNER/REPO")
				},
			},
			httpStub: func(r *httpmock.Registry) {
				r.Register(
					httpmock.GraphQL(`query IssueByNumber\b`),
					httpmock.StringResponse(`{"data":{"repository":{
						"hasIssuesEnabled": true,
						"issue":{"number":13}
					}}}`))
			},
			wantIssue: 13,
			wantRepo:  literal_4690,
		},
		{
			name: "URL argument",
			args: args{
				selector:   "https://example.org/OWNER/REPO/issues/13#comment-123",
				baseRepoFn: nil,
			},
			httpStub: func(r *httpmock.Registry) {
				r.Register(
					httpmock.GraphQL(`query IssueByNumber\b`),
					httpmock.StringResponse(`{"data":{"repository":{
						"hasIssuesEnabled": true,
						"issue":{"number":13}
					}}}`))
			},
			wantIssue: 13,
			wantRepo:  literal_9761,
		},
		{
			name: "PR URL argument",
			args: args{
				selector:   "https://example.org/OWNER/REPO/pull/13#comment-123",
				baseRepoFn: nil,
			},
			httpStub: func(r *httpmock.Registry) {
				r.Register(
					httpmock.GraphQL(`query IssueByNumber\b`),
					httpmock.StringResponse(`{"data":{"repository":{
						"hasIssuesEnabled": true,
						"issue":{"number":13}
					}}}`))
			},
			wantIssue: 13,
			wantRepo:  literal_9761,
		},
		{
			name: "project cards permission issue",
			args: args{
				selector:   "https://example.org/OWNER/REPO/issues/13",
				baseRepoFn: nil,
			},
			httpStub: func(r *httpmock.Registry) {
				r.Register(
					httpmock.GraphQL(`query IssueByNumber\b`),
					httpmock.StringResponse(`
					{
						"data": {
							"repository": {
								"hasIssuesEnabled": true,
								"issue": {
									"number": 13,
									"projectCards": {
										"nodes": [
											null,
											{
												"project": {"name": "myproject"},
												"column": {"name": "To Do"}
											},
											null,
											{
												"project": {"name": "other project"},
												"column": null
											}
										]
									}
								}
							}
						},
						"errors": [
							{
								"type": "FORBIDDEN",
								"message": "Resource not accessible by integration",
								"path": ["repository", "issue", "projectCards", "nodes", 0]
							},
							{
								"type": "FORBIDDEN",
								"message": "Resource not accessible by integration",
								"path": ["repository", "issue", "projectCards", "nodes", 2]
							}
						]
					}`))
			},
			wantErr:      true,
			wantIssue:    13,
			wantProjects: "myproject, other project",
			wantRepo:     literal_9761,
		},
		{
			name: "projects permission issue",
			args: args{
				selector:   "https://example.org/OWNER/REPO/issues/13",
				baseRepoFn: nil,
			},
			httpStub: func(r *httpmock.Registry) {
				r.Register(
					httpmock.GraphQL(`query IssueByNumber\b`),
					httpmock.StringResponse(`
					{
						"data": {
							"repository": {
								"hasIssuesEnabled": true,
								"issue": {
									"number": 13,
									"projectCards": {
										"nodes": null,
										"totalCount": 0
									}
								}
							}
						},
						"errors": [
							{
								"type": "FORBIDDEN",
								"message": "Resource not accessible by integration",
								"path": ["repository", "issue", "projectCards", "nodes"]
							}
						]
					}`))
			},
			wantErr:      true,
			wantIssue:    13,
			wantProjects: "",
			wantRepo:     literal_9761,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reg := &httpmock.Registry{}
			if tt.httpStub != nil {
				tt.httpStub(reg)
			}
			httpClient := &http.Client{Transport: reg}
			issue, repo, err := IssueFromArgWithFields(httpClient, tt.args.baseRepoFn, tt.args.selector, []string{"number"})
			if (err != nil) != tt.wantErr {
				t.Errorf("IssueFromArgWithFields() error = %v, wantErr %v", err, tt.wantErr)
				if issue == nil {
					return
				}
			}
			if issue.Number != tt.wantIssue {
				t.Errorf("want issue #%d, got #%d", tt.wantIssue, issue.Number)
			}
			if gotProjects := strings.Join(issue.ProjectCards.ProjectNames(), ", "); gotProjects != tt.wantProjects {
				t.Errorf("want projects %q, got %q", tt.wantProjects, gotProjects)
			}
			repoURL := ghrepo.GenerateRepoURL(repo, "")
			if repoURL != tt.wantRepo {
				t.Errorf("want repo %s, got %s", tt.wantRepo, repoURL)
			}
		})
	}
}

const literal_4690 = "https://github.com/OWNER/REPO"

const literal_9761 = "https://example.org/OWNER/REPO"
