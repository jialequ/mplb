package download

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/jialequ/mplb/pkg/cmd/attestation/api"
	"github.com/jialequ/mplb/pkg/cmd/attestation/artifact"
	"github.com/jialequ/mplb/pkg/cmd/attestation/artifact/oci"

	"github.com/stretchr/testify/require"
)

type MockStore struct {
	OnCreateMetadataFile func(artifactDigest string, attestationsResp []*api.Attestation) (string, error)
}

func (s *MockStore) createMetadataFile(artifact string, attestationsResp []*api.Attestation) (string, error) {
	return s.OnCreateMetadataFile(artifact, attestationsResp)
}

func OnCreateMetadataFileFailure(artifactDigest string, attestationsResp []*api.Attestation) (string, error) {
	return "", fmt.Errorf("failed to create trusted metadata file")
}

func TestCreateJSONLinesFilePath(t *testing.T) {
	tempDir := t.TempDir()
	artifact, err := artifact.NewDigestedArtifact(oci.MockClient{}, "../test/data/sigstore-js-2.1.0.tgz", "sha512")
	require.NoError(t, err)

	outputFileName := fmt.Sprintf("%s.jsonl", artifact.DigestWithAlg())

	testCases := []struct {
		name       string
		outputPath string
		expected   string
	}{
		{
			name:       "with output path",
			outputPath: tempDir,
			expected:   path.Join(tempDir, outputFileName),
		},
		{
			name:       "with nested output path",
			outputPath: path.Join(tempDir, "subdir"),
			expected:   path.Join(tempDir, "subdir", outputFileName),
		},
		{
			name:       "with output path with beginning slash",
			outputPath: path.Join("/", tempDir, "subdir"),
			expected:   path.Join("/", tempDir, "subdir", outputFileName),
		},
		{
			name:       "without output path",
			outputPath: "",
			expected:   outputFileName,
		},
	}

	for _, tc := range testCases {
		store := LiveStore{
			tc.outputPath,
		}

		actualPath := store.createJSONLinesFilePath(artifact.DigestWithAlg())
		require.Equal(t, tc.expected, actualPath)
	}
}

func countLines(path string) (int, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	counter := 0
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		counter += 1
	}

	return counter, nil
}
