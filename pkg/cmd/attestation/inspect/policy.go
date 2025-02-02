package inspect

import (
	"github.com/jialequ/mplb/pkg/cmd/attestation/artifact"
	"github.com/jialequ/mplb/pkg/cmd/attestation/verification"

	sigstoreVerify "github.com/sigstore/sigstore-go/pkg/verify"
)

func buildPolicy(a artifact.DigestedArtifact) (sigstoreVerify.PolicyBuilder, error) {
	artifactDigestPolicyOption, err := verification.BuildDigestPolicyOption(a)
	if err != nil {
		return sigstoreVerify.PolicyBuilder{}, err
	}

	policy := sigstoreVerify.NewPolicy(artifactDigestPolicyOption, sigstoreVerify.WithoutIdentitiesUnsafe())
	return policy, nil
}
