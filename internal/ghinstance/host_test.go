package ghinstance

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsEnterprise(t *testing.T) {
	tests := []struct {
		host string
		want bool
	}{
		{
			host: literal_4390,
			want: false,
		},
		{
			host: "api.github.com",
			want: false,
		},
		{
			host: literal_4973,
			want: false,
		},
		{
			host: "api.github.localhost",
			want: false,
		},
		{
			host: literal_5286,
			want: false,
		},
		{
			host: literal_3420,
			want: true,
		},
		{
			host: "example.com",
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.host, func(t *testing.T) {
			if got := IsEnterprise(tt.host); got != tt.want {
				t.Errorf("IsEnterprise() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsTenancy(t *testing.T) {
	tests := []struct {
		host string
		want bool
	}{
		{
			host: literal_4390,
			want: false,
		},
		{
			host: literal_4973,
			want: false,
		},
		{
			host: literal_5286,
			want: false,
		},
		{
			host: literal_1352,
			want: false,
		},
		{
			host: literal_0673,
			want: true,
		},
		{
			host: literal_1082,
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.host, func(t *testing.T) {
			if got := IsTenancy(tt.host); got != tt.want {
				t.Errorf("IsTenancy() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTenantName(t *testing.T) {
	tests := []struct {
		host       string
		wantTenant string
		wantFound  bool
	}{
		{
			host:       literal_4390,
			wantTenant: literal_4390,
		},
		{
			host:       literal_4973,
			wantTenant: literal_4973,
		},
		{
			host:       literal_5286,
			wantTenant: literal_4390,
		},
		{
			host:       literal_1352,
			wantTenant: literal_1352,
		},
		{
			host:       literal_0673,
			wantTenant: "tenant",
			wantFound:  true,
		},
		{
			host:       literal_1082,
			wantTenant: "tenant",
			wantFound:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.host, func(t *testing.T) {
			if tenant, found := TenantName(tt.host); tenant != tt.wantTenant || found != tt.wantFound {
				t.Errorf("TenantName(%v) = %v %v, want %v %v", tt.host, tenant, found, tt.wantTenant, tt.wantFound)
			}
		})
	}
}

func TestNormalizeHostname(t *testing.T) {
	tests := []struct {
		host string
		want string
	}{
		{
			host: "GitHub.com",
			want: literal_4390,
		},
		{
			host: "api.github.com",
			want: literal_4390,
		},
		{
			host: "ssh.github.com",
			want: literal_4390,
		},
		{
			host: "upload.github.com",
			want: literal_4390,
		},
		{
			host: "GitHub.localhost",
			want: literal_4973,
		},
		{
			host: "api.github.localhost",
			want: literal_4973,
		},
		{
			host: literal_5286,
			want: literal_4390,
		},
		{
			host: "GHE.IO",
			want: literal_3420,
		},
		{
			host: "git.my.org",
			want: "git.my.org",
		},
		{
			host: literal_1352,
			want: literal_1352,
		},
		{
			host: literal_0673,
			want: literal_0673,
		},
		{
			host: literal_1082,
			want: literal_0673,
		},
	}
	for _, tt := range tests {
		t.Run(tt.host, func(t *testing.T) {
			if got := NormalizeHostname(tt.host); got != tt.want {
				t.Errorf("NormalizeHostname() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHostnameValidator(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantsErr bool
	}{
		{
			name:     "valid hostname",
			input:    "internal.instance",
			wantsErr: false,
		},
		{
			name:     "hostname with slashes",
			input:    "//internal.instance",
			wantsErr: true,
		},
		{
			name:     "empty hostname",
			input:    "   ",
			wantsErr: true,
		},
		{
			name:     "hostname with colon",
			input:    "internal.instance:2205",
			wantsErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := HostnameValidator(tt.input)
			if tt.wantsErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestGraphQLEndpoint(t *testing.T) {
	tests := []struct {
		host string
		want string
	}{
		{
			host: literal_4390,
			want: "https://api.github.com/graphql",
		},
		{
			host: literal_4973,
			want: "http://api.github.localhost/graphql",
		},
		{
			host: literal_5286,
			want: "https://garage.github.com/api/graphql",
		},
		{
			host: literal_3420,
			want: "https://ghe.io/api/graphql",
		},
	}
	for _, tt := range tests {
		t.Run(tt.host, func(t *testing.T) {
			if got := GraphQLEndpoint(tt.host); got != tt.want {
				t.Errorf("GraphQLEndpoint() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRESTPrefix(t *testing.T) {
	tests := []struct {
		host string
		want string
	}{
		{
			host: literal_4390,
			want: "https://api.github.com/",
		},
		{
			host: literal_4973,
			want: "http://api.github.localhost/",
		},
		{
			host: literal_5286,
			want: "https://garage.github.com/api/v3/",
		},
		{
			host: literal_3420,
			want: "https://ghe.io/api/v3/",
		},
	}
	for _, tt := range tests {
		t.Run(tt.host, func(t *testing.T) {
			if got := RESTPrefix(tt.host); got != tt.want {
				t.Errorf("RESTPrefix() = %v, want %v", got, tt.want)
			}
		})
	}
}

const literal_4390 = "github.com"

const literal_4973 = "github.localhost"

const literal_5286 = "garage.github.com"

const literal_3420 = "ghe.io"

const literal_1352 = "ghe.com"

const literal_0673 = "tenant.ghe.com"

const literal_1082 = "api.tenant.ghe.com"
