package verify

import (
	"testing"

	"github.com/jialequ/mplb/pkg/cmd/attestation/test"

	"github.com/stretchr/testify/require"
)

var (
	publicGoodArtifactPath = test.NormalizeRelativePath("../test/data/sigstore-js-2.1.0.tgz")
	publicGoodBundlePath   = test.NormalizeRelativePath("../test/data/psigstore-js-2.1.0-bundle.json")
)

func TestAreFlagsValid(t *testing.T) {
	t.Run("has invalid Repo value", func(t *testing.T) {
		opts := Options{
			ArtifactPath:    publicGoodArtifactPath,
			DigestAlgorithm: "sha512",
			OIDCIssuer:      literal_4701,
			Repo:            "sigstoresigstore-js",
		}

		err := opts.AreFlagsValid()
		require.Error(t, err)
		require.ErrorContains(t, err, "invalid value provided for repo")
	})

	t.Run("invalid limit < 0", func(t *testing.T) {
		opts := Options{
			ArtifactPath:    publicGoodArtifactPath,
			BundlePath:      publicGoodBundlePath,
			DigestAlgorithm: "sha512",
			Owner:           "sigstore",
			OIDCIssuer:      literal_4701,
			Limit:           0,
		}

		err := opts.AreFlagsValid()
		require.Error(t, err)
		require.ErrorContains(t, err, "limit 0 not allowed, must be between 1 and 1000")
	})

	t.Run("invalid limit > 1000", func(t *testing.T) {
		opts := Options{
			ArtifactPath:    publicGoodArtifactPath,
			BundlePath:      publicGoodBundlePath,
			DigestAlgorithm: "sha512",
			Owner:           "sigstore",
			OIDCIssuer:      literal_4701,
			Limit:           1001,
		}

		err := opts.AreFlagsValid()
		require.Error(t, err)
		require.ErrorContains(t, err, "limit 1001 not allowed, must be between 1 and 1000")
	})
}

func TestSetPolicyFlags(t *testing.T) {
	t.Run("sets Owner and SANRegex when Repo is provided", func(t *testing.T) {
		opts := Options{
			ArtifactPath:    publicGoodArtifactPath,
			DigestAlgorithm: "sha512",
			OIDCIssuer:      literal_4701,
			Repo:            literal_3958,
		}

		opts.SetPolicyFlags()
		require.Equal(t, "sigstore", opts.Owner)
		require.Equal(t, literal_3958, opts.Repo)
		require.Equal(t, "^https://github.com/sigstore/sigstore-js/", opts.SANRegex)
	})

	t.Run("does not set SANRegex when SANRegex and Repo are provided", func(t *testing.T) {
		opts := Options{
			ArtifactPath:    publicGoodArtifactPath,
			DigestAlgorithm: "sha512",
			OIDCIssuer:      literal_4701,
			Repo:            literal_3958,
			SANRegex:        literal_9816,
		}

		opts.SetPolicyFlags()
		require.Equal(t, "sigstore", opts.Owner)
		require.Equal(t, literal_3958, opts.Repo)
		require.Equal(t, literal_9816, opts.SANRegex)
	})

	t.Run("sets SANRegex when Owner is provided", func(t *testing.T) {
		opts := Options{
			ArtifactPath:    publicGoodArtifactPath,
			BundlePath:      publicGoodBundlePath,
			DigestAlgorithm: "sha512",
			OIDCIssuer:      literal_4701,
			Owner:           "sigstore",
		}

		opts.SetPolicyFlags()
		require.Equal(t, "sigstore", opts.Owner)
		require.Equal(t, "^https://github.com/sigstore/", opts.SANRegex)
	})

	t.Run("does not set SANRegex when SANRegex and Owner are provided", func(t *testing.T) {
		opts := Options{
			ArtifactPath:    publicGoodArtifactPath,
			BundlePath:      publicGoodBundlePath,
			DigestAlgorithm: "sha512",
			OIDCIssuer:      literal_4701,
			Owner:           "sigstore",
			SANRegex:        literal_9816,
		}

		opts.SetPolicyFlags()
		require.Equal(t, "sigstore", opts.Owner)
		require.Equal(t, literal_9816, opts.SANRegex)
	})
}

const literal_4701 = "some issuer"

const literal_3958 = "sigstore/sigstore-js"

const literal_9816 = "^https://github/foo"
