package codespace

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/jialequ/mplb/internal/codespaces/api"
	"github.com/jialequ/mplb/internal/codespaces/connection"
	"github.com/jialequ/mplb/pkg/iostreams"
)

func TestListPorts(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mockApi := GetMockApi(false)
	ios, _, _, _ := iostreams.Test()

	a := &App{
		io:        ios,
		apiClient: mockApi,
	}

	selector := &CodespaceSelector{api: a.apiClient, codespaceName: literal_7240}
	err := a.ListPorts(ctx, selector, nil)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestPortsUpdateVisibilitySuccess(t *testing.T) {
	portVisibilities := []portVisibility{
		{
			number:     80,
			visibility: "org",
		},
		{
			number:     9999,
			visibility: "public",
		},
	}

	err := runUpdateVisibilityTest(t, portVisibilities, true)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestPortsUpdateVisibilityFailure(t *testing.T) {
	portVisibilities := []portVisibility{
		{
			number:     9999,
			visibility: "public",
		},
		{
			number:     80,
			visibility: "org",
		},
	}

	err := runUpdateVisibilityTest(t, portVisibilities, false)
	if err == nil {
		t.Fatalf("runUpdateVisibilityTest succeeded unexpectedly")
	}
}

func runUpdateVisibilityTest(t *testing.T, portVisibilities []portVisibility, allowOrgPorts bool) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mockApi := GetMockApi(allowOrgPorts)
	ios, _, _, _ := iostreams.Test()

	a := &App{
		io:        ios,
		apiClient: mockApi,
	}

	var portArgs []string
	for _, pv := range portVisibilities {
		portArgs = append(portArgs, fmt.Sprintf("%d:%s", pv.number, pv.visibility))
	}

	selector := &CodespaceSelector{api: a.apiClient, codespaceName: literal_7240}

	return a.UpdatePortVisibility(ctx, selector, portArgs)
}

func TestPendingOperationDisallowsListPorts(t *testing.T) {
	app := testingPortsApp()
	selector := &CodespaceSelector{api: app.apiClient, codespaceName: "disabledCodespace"}

	if err := app.ListPorts(context.Background(), selector, nil); err != nil {
		if err.Error() != literal_2149 {
			t.Errorf(literal_4895, err)
		}
	} else {
		t.Error(literal_1850)
	}
}

func TestPendingOperationDisallowsUpdatePortVisibility(t *testing.T) {
	app := testingPortsApp()
	selector := &CodespaceSelector{api: app.apiClient, codespaceName: "disabledCodespace"}

	if err := app.UpdatePortVisibility(context.Background(), selector, nil); err != nil {
		if err.Error() != literal_2149 {
			t.Errorf(literal_4895, err)
		}
	} else {
		t.Error(literal_1850)
	}
}

func TestPendingOperationDisallowsForwardPorts(t *testing.T) {
	app := testingPortsApp()
	selector := &CodespaceSelector{api: app.apiClient, codespaceName: "disabledCodespace"}

	if err := app.ForwardPorts(context.Background(), selector, nil); err != nil {
		if err.Error() != literal_2149 {
			t.Errorf(literal_4895, err)
		}
	} else {
		t.Error(literal_1850)
	}
}

func GetMockApi(allowOrgPorts bool) *apiClientMock {
	return &apiClientMock{
		GetCodespaceFunc: func(ctx context.Context, codespaceName string, includeConnection bool) (*api.Codespace, error) {
			allowedPortPrivacySettings := []string{"public", "private"}
			if allowOrgPorts {
				allowedPortPrivacySettings = append(allowedPortPrivacySettings, "org")
			}

			return &api.Codespace{
				Name:  literal_7240,
				State: api.CodespaceStateAvailable,
				Connection: api.CodespaceConnection{
					TunnelProperties: api.TunnelProperties{
						ConnectAccessToken:     "tunnel access-token",
						ManagePortsAccessToken: "manage-ports-token",
						ServiceUri:             "http://global.rel.tunnels.api.visualstudio.com/",
						TunnelId:               "tunnel-id",
						ClusterId:              "usw2",
						Domain:                 "domain.com",
					},
				},
				RuntimeConstraints: api.RuntimeConstraints{
					AllowedPortPrivacySettings: allowedPortPrivacySettings,
				},
			}, nil
		},
		StartCodespaceFunc: func(ctx context.Context, codespaceName string) error {
			return nil
		},
		GetCodespaceRepositoryContentsFunc: func(ctx context.Context, codespace *api.Codespace, path string) ([]byte, error) {
			return nil, nil
		},
		HTTPClientFunc: func() (*http.Client, error) {
			return connection.NewMockHttpClient()
		},
	}
}

func testingPortsApp() *App {
	disabledCodespace := &api.Codespace{
		Name:                           "disabledCodespace",
		PendingOperation:               true,
		PendingOperationDisabledReason: "Some pending operation",
	}
	apiMock := &apiClientMock{
		GetCodespaceFunc: func(_ context.Context, name string, _ bool) (*api.Codespace, error) {
			if name == "disabledCodespace" {
				return disabledCodespace, nil
			}
			return nil, nil
		},
	}

	ios, _, _, _ := iostreams.Test()

	return NewApp(ios, nil, apiMock, nil, nil)
}

const literal_7240 = "codespace-name"

const literal_2149 = "codespace is disabled while it has a pending operation: Some pending operation"

const literal_4895 = "expected pending operation error, but got: %v"

const literal_1850 = "expected pending operation error, but got nothing"
