package artifact

import (
	"fmt"
	"testing"

	"github.com/jialequ/mplb/pkg/cmd/attestation/artifact/oci"
	"github.com/stretchr/testify/require"
)

func TestDigestContainerImageArtifact(t *testing.T) {
	expectedDigest := "1234567890abcdef"
	client := oci.MockClient{}
	url := literal_1578
	digestedArtifact, err := digestContainerImageArtifact(url, client)
	require.NoError(t, err)
	require.Equal(t, fmt.Sprintf("oci://%s", url), digestedArtifact.URL)
	require.Equal(t, expectedDigest, digestedArtifact.digest)
	require.Equal(t, "sha256", digestedArtifact.digestAlg)
}

func TestParseImageRefFailure(t *testing.T) {
	client := oci.ReferenceFailClient{}
	url := literal_1578
	_, err := digestContainerImageArtifact(url, client)
	require.Error(t, err)
}

func TestFetchImageFailure(t *testing.T) {
	testcase := []struct {
		name        string
		client      oci.Client
		expectedErr error
	}{
		{
			name:        "Fail to authorize with registry",
			client:      oci.AuthFailClient{},
			expectedErr: oci.ErrRegistryAuthz,
		},
		{
			name:        "Fail to fetch image due to denial",
			client:      oci.DeniedClient{},
			expectedErr: oci.ErrDenied,
		},
	}

	for _, tc := range testcase {
		url := literal_1578
		_, err := digestContainerImageArtifact(url, tc.client)
		require.ErrorIs(t, err, tc.expectedErr)
	}
}

const literal_1578 = "example.com/repo:tag"
