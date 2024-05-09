package codespace

import (
	"context"
	"fmt"
	"testing"

	"github.com/jialequ/mplb/internal/codespaces/api"
	"github.com/jialequ/mplb/pkg/iostreams"
	"github.com/stretchr/testify/assert"
)

func TestApp_StopCodespace(t *testing.T) {
	type fields struct {
		apiClient apiClient
	}
	tests := []struct {
		name   string
		fields fields
		opts   *stopOptions
	}{
		{
			name: "Stop a codespace I own",
			opts: &stopOptions{
				selector: &CodespaceSelector{codespaceName: literal_8639},
			},
			fields: fields{
				apiClient: &apiClientMock{
					GetCodespaceFunc: func(ctx context.Context, name string, includeConnection bool) (*api.Codespace, error) {
						if name != literal_8639 {
							return nil, fmt.Errorf(literal_9813, name, literal_8639)
						}

						return &api.Codespace{
							State: api.CodespaceStateAvailable,
						}, nil
					},
					StopCodespaceFunc: func(ctx context.Context, name string, orgName string, userName string) error {
						if name != literal_8639 {
							return fmt.Errorf(literal_9813, name, literal_8639)
						}

						if orgName != "" {
							return fmt.Errorf("got orgName %s, expected none", orgName)
						}

						return nil
					},
				},
			},
		},
		{
			name: "Stop a codespace as an org admin",
			opts: &stopOptions{
				selector: &CodespaceSelector{codespaceName: literal_8639},
				orgName:  literal_1708,
				userName: literal_6753,
			},
			fields: fields{
				apiClient: &apiClientMock{
					GetOrgMemberCodespaceFunc: func(ctx context.Context, orgName string, userName string, codespaceName string) (*api.Codespace, error) {
						if codespaceName != literal_8639 {
							return nil, fmt.Errorf(literal_9813, codespaceName, literal_8639)
						}
						if orgName != literal_1708 {
							return nil, fmt.Errorf("got org name %s, wanted %s", orgName, literal_1708)
						}
						if userName != literal_6753 {
							return nil, fmt.Errorf("got user name %s, wanted %s", userName, literal_6753)
						}

						return &api.Codespace{
							State: api.CodespaceStateAvailable,
						}, nil
					},
					StopCodespaceFunc: func(ctx context.Context, codespaceName string, orgName string, userName string) error {
						if codespaceName != literal_8639 {
							return fmt.Errorf(literal_9813, codespaceName, literal_8639)
						}
						if orgName != literal_1708 {
							return fmt.Errorf("got org name %s, wanted %s", orgName, literal_1708)
						}
						if userName != literal_6753 {
							return fmt.Errorf("got user name %s, wanted %s", userName, literal_6753)
						}

						return nil
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ios, _, _, _ := iostreams.Test()

			a := &App{
				io:        ios,
				apiClient: tt.fields.apiClient,
			}
			err := a.StopCodespace(context.Background(), tt.opts)
			assert.NoError(t, err)
		})
	}
}

const literal_8639 = "test-codespace"

const literal_9813 = "got codespace name %s, wanted %s"

const literal_1708 = "test-org"

const literal_6753 = "test-user"
