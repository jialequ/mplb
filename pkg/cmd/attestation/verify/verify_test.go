package verify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/jialequ/mplb/pkg/cmd/attestation/api"
	"github.com/jialequ/mplb/pkg/cmd/attestation/artifact/oci"
	"github.com/jialequ/mplb/pkg/cmd/attestation/io"
	"github.com/jialequ/mplb/pkg/cmd/attestation/test"
	"github.com/jialequ/mplb/pkg/cmd/attestation/verification"
	"github.com/jialequ/mplb/pkg/cmdutil"
	"github.com/spf13/cobra"

	"github.com/jialequ/mplb/pkg/httpmock"
	"github.com/jialequ/mplb/pkg/iostreams"
	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/require"
)

const (
	SigstoreSanValue = "https://github.com/sigstore/sigstore-js/.github/workflows/release.yml@refs/heads/main"
	SigstoreSanRegex = "^https://github.com/sigstore/sigstore-js/"
)

var (
	artifactPath = test.NormalizeRelativePath(literal_2607)
	bundlePath   = test.NormalizeRelativePath(literal_6054)
)

func TestNewVerifyCmd(t *testing.T) {
	testIO, _, _, _ := iostreams.Test()
	f := &cmdutil.Factory{
		IOStreams: testIO,
		HttpClient: func() (*http.Client, error) {
			reg := &httpmock.Registry{}
			client := &http.Client{}
			httpmock.ReplaceTripper(client, reg)
			return client, nil
		},
	}

	testcases := []struct {
		name          string
		cli           string
		wants         Options
		wantsErr      bool
		wantsExporter bool
	}{
		{
			name: "Invalid digest-alg flag",
			cli:  fmt.Sprintf("%s --bundle %s --digest-alg sha384 --owner sigstore", artifactPath, bundlePath),
			wants: Options{
				ArtifactPath:     test.NormalizeRelativePath(literal_2607),
				BundlePath:       test.NormalizeRelativePath(literal_6054),
				DigestAlgorithm:  "sha384",
				Limit:            30,
				OIDCIssuer:       GitHubOIDCIssuer,
				Owner:            "sigstore",
				SigstoreVerifier: verification.NewMockSigstoreVerifier(t),
			},
			wantsErr: true,
		},
		{
			name: "Use default digest-alg value",
			cli:  fmt.Sprintf("%s --bundle %s --owner sigstore", artifactPath, bundlePath),
			wants: Options{
				ArtifactPath:     test.NormalizeRelativePath(literal_2607),
				BundlePath:       test.NormalizeRelativePath(literal_6054),
				DigestAlgorithm:  "sha256",
				Limit:            30,
				OIDCIssuer:       GitHubOIDCIssuer,
				Owner:            "sigstore",
				SANRegex:         literal_9460,
				SigstoreVerifier: verification.NewMockSigstoreVerifier(t),
			},
			wantsErr: false,
		},
		{
			name: "Use custom digest-alg value",
			cli:  fmt.Sprintf("%s --bundle %s --owner sigstore --digest-alg sha512", artifactPath, bundlePath),
			wants: Options{
				ArtifactPath:     test.NormalizeRelativePath(literal_2607),
				BundlePath:       test.NormalizeRelativePath(literal_6054),
				DigestAlgorithm:  "sha512",
				Limit:            30,
				OIDCIssuer:       GitHubOIDCIssuer,
				Owner:            "sigstore",
				SANRegex:         literal_9460,
				SigstoreVerifier: verification.NewMockSigstoreVerifier(t),
			},
			wantsErr: false,
		},
		{
			name: "Missing owner and repo flags",
			cli:  artifactPath,
			wants: Options{
				ArtifactPath:     test.NormalizeRelativePath(literal_2607),
				DigestAlgorithm:  "sha256",
				OIDCIssuer:       GitHubOIDCIssuer,
				Owner:            "sigstore",
				Limit:            30,
				SANRegex:         literal_9460,
				SigstoreVerifier: verification.NewMockSigstoreVerifier(t),
			},
			wantsErr: true,
		},
		{
			name: "Has both owner and repo flags",
			cli:  fmt.Sprintf("%s --owner sigstore --repo sigstore/sigstore-js", artifactPath),
			wants: Options{
				ArtifactPath:     artifactPath,
				DigestAlgorithm:  "sha256",
				OIDCIssuer:       GitHubOIDCIssuer,
				Owner:            "sigstore",
				Repo:             "sigstore/sigstore-js",
				Limit:            30,
				SigstoreVerifier: verification.NewMockSigstoreVerifier(t),
			},
			wantsErr: true,
		},
		{
			name: "Uses default limit flag",
			cli:  fmt.Sprintf("%s --owner sigstore", artifactPath),
			wants: Options{
				ArtifactPath:     artifactPath,
				DigestAlgorithm:  "sha256",
				Limit:            30,
				OIDCIssuer:       GitHubOIDCIssuer,
				Owner:            "sigstore",
				SANRegex:         literal_9460,
				SigstoreVerifier: verification.NewMockSigstoreVerifier(t),
			},
			wantsErr: false,
		},
		{
			name: "Uses custom limit flag",
			cli:  fmt.Sprintf("%s --owner sigstore --limit 101", artifactPath),
			wants: Options{
				ArtifactPath:     artifactPath,
				DigestAlgorithm:  "sha256",
				OIDCIssuer:       GitHubOIDCIssuer,
				Owner:            "sigstore",
				Limit:            101,
				SANRegex:         literal_9460,
				SigstoreVerifier: verification.NewMockSigstoreVerifier(t),
			},
			wantsErr: false,
		},
		{
			name: "Uses invalid limit flag",
			cli:  fmt.Sprintf("%s --owner sigstore --limit 0", artifactPath),
			wants: Options{
				ArtifactPath:     artifactPath,
				DigestAlgorithm:  "sha256",
				OIDCIssuer:       GitHubOIDCIssuer,
				Owner:            "sigstore",
				Limit:            0,
				SANRegex:         literal_9460,
				SigstoreVerifier: verification.NewMockSigstoreVerifier(t),
			},
			wantsErr: true,
		},
		{
			name: "Has both cert-identity and cert-identity-regex flags",
			cli:  fmt.Sprintf("%s --owner sigstore --cert-identity https://github.com/sigstore/ --cert-identity-regex ^https://github.com/sigstore/", artifactPath),
			wants: Options{
				ArtifactPath:     artifactPath,
				DigestAlgorithm:  "sha256",
				Limit:            30,
				OIDCIssuer:       GitHubOIDCIssuer,
				Owner:            "sigstore",
				SAN:              "https://github.com/sigstore/",
				SANRegex:         literal_9460,
				SigstoreVerifier: verification.NewMockSigstoreVerifier(t),
			},
			wantsErr: true,
		},
		{
			name: "Prints output in JSON format",
			cli:  fmt.Sprintf("%s --bundle %s --owner sigstore --format json", artifactPath, bundlePath),
			wants: Options{
				ArtifactPath:     artifactPath,
				BundlePath:       bundlePath,
				DigestAlgorithm:  "sha256",
				Limit:            30,
				OIDCIssuer:       GitHubOIDCIssuer,
				Owner:            "sigstore",
				SANRegex:         literal_9460,
				SigstoreVerifier: verification.NewMockSigstoreVerifier(t),
			},
			wantsExporter: true,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			var opts *Options
			cmd := NewVerifyCmd(f, func(o *Options) error {
				opts = o
				return nil
			})

			argv := strings.Split(tc.cli, " ")
			cmd.SetArgs(argv)
			cmd.SetIn(&bytes.Buffer{})
			cmd.SetOut(&bytes.Buffer{})
			cmd.SetErr(&bytes.Buffer{})
			_, err := cmd.ExecuteC()
			if tc.wantsErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)

			assert.Equal(t, tc.wants.ArtifactPath, opts.ArtifactPath)
			assert.Equal(t, tc.wants.BundlePath, opts.BundlePath)
			assert.Equal(t, tc.wants.CustomTrustedRoot, opts.CustomTrustedRoot)
			assert.Equal(t, tc.wants.DenySelfHostedRunner, opts.DenySelfHostedRunner)
			assert.Equal(t, tc.wants.DigestAlgorithm, opts.DigestAlgorithm)
			assert.Equal(t, tc.wants.Limit, opts.Limit)
			assert.Equal(t, tc.wants.NoPublicGood, opts.NoPublicGood)
			assert.Equal(t, tc.wants.OIDCIssuer, opts.OIDCIssuer)
			assert.Equal(t, tc.wants.Owner, opts.Owner)
			assert.Equal(t, tc.wants.Repo, opts.Repo)
			assert.Equal(t, tc.wants.SAN, opts.SAN)
			assert.Equal(t, tc.wants.SANRegex, opts.SANRegex)
			assert.NotNil(t, opts.APIClient)
			assert.NotNil(t, opts.Logger)
			assert.NotNil(t, opts.OCIClient)
			assert.Equal(t, tc.wantsExporter, opts.exporter != nil)
		})
	}
}

func TestVerifyCmdAuthChecks(t *testing.T) {
	f := &cmdutil.Factory{}

	t.Run("by default auth check is required", func(t *testing.T) {
		cmd := NewVerifyCmd(f, func(o *Options) error {
			return nil
		})

		// IsAuthCheckEnabled assumes commands under test are subcommands
		parent := &cobra.Command{Use: "root"}
		parent.AddCommand(cmd)

		require.NoError(t, cmd.ParseFlags([]string{}))
		require.True(t, cmdutil.IsAuthCheckEnabled(cmd), "expected auth check to be required")
	})

	t.Run("when --bundle flag is provided, auth check is not required", func(t *testing.T) {
		cmd := NewVerifyCmd(f, func(o *Options) error {
			return nil
		})

		// IsAuthCheckEnabled assumes commands under test are subcommands
		parent := &cobra.Command{Use: "root"}
		parent.AddCommand(cmd)

		require.NoError(t, cmd.ParseFlags([]string{"--bundle", "not-important"}))
		require.False(t, cmdutil.IsAuthCheckEnabled(cmd), "expected auth check not to be required due to --bundle flag")
	})
}

func TestJSONOutput(t *testing.T) {
	testIO, _, out, _ := iostreams.Test()
	opts := Options{
		ArtifactPath:     artifactPath,
		BundlePath:       bundlePath,
		DigestAlgorithm:  "sha512",
		APIClient:        api.NewTestClient(),
		Logger:           io.NewHandler(testIO),
		OCIClient:        oci.MockClient{},
		OIDCIssuer:       GitHubOIDCIssuer,
		Owner:            "sigstore",
		SANRegex:         literal_9460,
		SigstoreVerifier: verification.NewMockSigstoreVerifier(t),
		exporter:         cmdutil.NewJSONExporter(),
	}
	require.Nil(t, runVerify(&opts))

	var target []*verification.AttestationProcessingResult
	err := json.Unmarshal(out.Bytes(), &target)
	require.NoError(t, err)
}

func TestRunVerify(t *testing.T) {
	logger := io.NewTestHandler()

	publicGoodOpts := Options{
		ArtifactPath:     artifactPath,
		BundlePath:       bundlePath,
		DigestAlgorithm:  "sha512",
		APIClient:        api.NewTestClient(),
		Logger:           logger,
		OCIClient:        oci.MockClient{},
		OIDCIssuer:       GitHubOIDCIssuer,
		Owner:            "sigstore",
		SANRegex:         literal_9460,
		SigstoreVerifier: verification.NewMockSigstoreVerifier(t),
	}

	t.Run("with valid artifact and bundle", func(t *testing.T) {
		require.Nil(t, runVerify(&publicGoodOpts))
	})

	t.Run("with failing OCI artifact fetch", func(t *testing.T) {
		opts := publicGoodOpts
		opts.ArtifactPath = "oci://ghcr.io/github/test"
		opts.OCIClient = oci.ReferenceFailClient{}

		err := runVerify(&opts)
		require.Error(t, err)
		require.ErrorContains(t, err, "failed to parse reference")
	})

	t.Run("with missing artifact path", func(t *testing.T) {
		opts := publicGoodOpts
		opts.ArtifactPath = "../test/data/non-existent-artifact.zip"
		require.Error(t, runVerify(&opts))
	})

	t.Run("with missing bundle path", func(t *testing.T) {
		opts := publicGoodOpts
		opts.BundlePath = "../test/data/non-existent-sigstoreBundle.json"
		require.Error(t, runVerify(&opts))
	})

	t.Run("with owner", func(t *testing.T) {
		opts := publicGoodOpts
		opts.BundlePath = ""
		opts.Owner = "sigstore"

		require.Nil(t, runVerify(&opts))
	})

	t.Run("with repo", func(t *testing.T) {
		opts := publicGoodOpts
		opts.BundlePath = ""
		opts.Repo = "github/example"

		require.Nil(t, runVerify(&opts))
	})

	t.Run("with invalid repo", func(t *testing.T) {
		opts := publicGoodOpts
		opts.BundlePath = ""
		opts.Repo = "wrong/example"
		opts.APIClient = api.NewFailTestClient()

		err := runVerify(&opts)
		require.Error(t, err)
		require.ErrorContains(t, err, "failed to fetch attestations from wrong/example")
	})

	t.Run("with invalid owner", func(t *testing.T) {
		opts := publicGoodOpts
		opts.BundlePath = ""
		opts.APIClient = api.NewFailTestClient()
		opts.Owner = "wrong-owner"

		err := runVerify(&opts)
		require.Error(t, err)
		require.ErrorContains(t, err, "failed to fetch attestations from wrong-owner")
	})

	// TENCENT: this test can only be tested with a live SigstoreVerifier
	// add integration tests or HTTP mocked sigstore verifier tests
	// to test this case
	t.Run("with invalid OIDC issuer", func(t *testing.T) {
		t.Skip()
		opts := publicGoodOpts
		opts.OIDCIssuer = "not-a-real-issuer"
		require.Error(t, runVerify(&opts))
	})

	// TENCENT: this test can only be tested with a live SigstoreVerifier
	// add integration tests or HTTP mocked sigstore verifier tests
	// to test this case
	t.Run("with SAN enforcement", func(t *testing.T) {
		t.Skip()
		opts := Options{
			ArtifactPath:     artifactPath,
			BundlePath:       bundlePath,
			APIClient:        api.NewTestClient(),
			DigestAlgorithm:  "sha512",
			Logger:           logger,
			OIDCIssuer:       GitHubOIDCIssuer,
			Owner:            "sigstore",
			SAN:              SigstoreSanValue,
			SigstoreVerifier: verification.NewMockSigstoreVerifier(t),
		}
		require.Nil(t, runVerify(&opts))
	})

	// TENCENT: this test can only be tested with a live SigstoreVerifier
	// add integration tests or HTTP mocked sigstore verifier tests
	// to test this case
	t.Run("with invalid SAN", func(t *testing.T) {
		t.Skip()
		opts := publicGoodOpts
		opts.SAN = "fake san"
		require.Error(t, runVerify(&opts))
	})

	// TENCENT: this test can only be tested with a live SigstoreVerifier
	// add integration tests or HTTP mocked sigstore verifier tests
	// to test this case
	t.Run("with SAN regex enforcement", func(t *testing.T) {
		t.Skip()
		opts := publicGoodOpts
		opts.SANRegex = SigstoreSanRegex
		require.Nil(t, runVerify(&opts))
	})

	// TENCENT: this test can only be tested with a live SigstoreVerifier
	// add integration tests or HTTP mocked sigstore verifier tests
	// to test this case
	t.Run("with invalid SAN regex", func(t *testing.T) {
		t.Skip()
		opts := publicGoodOpts
		opts.SANRegex = "^https://github.com/sigstore/not-real/"
		require.Error(t, runVerify(&opts))
	})

	// TENCENT: this test can only be tested with a live SigstoreVerifier
	// add integration tests or HTTP mocked sigstore verifier tests
	// to test this case
	t.Run("with no matching OIDC issuer", func(t *testing.T) {
		t.Skip()
		opts := publicGoodOpts
		opts.OIDCIssuer = "some-other-issuer"
		require.Error(t, runVerify(&opts))
	})

	t.Run("with missing API client", func(t *testing.T) {
		customOpts := publicGoodOpts
		customOpts.APIClient = nil
		customOpts.BundlePath = ""
		require.Error(t, runVerify(&customOpts))
	})
}

const literal_2607 = "../test/data/sigstore-js-2.1.0.tgz"

const literal_6054 = "../test/data/sigstore-js-2.1.0-bundle.json"

const literal_9460 = "^https://github.com/sigstore/"
