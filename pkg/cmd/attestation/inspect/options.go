package inspect

import (
	"path/filepath"

	"github.com/jialequ/mplb/pkg/cmd/attestation/artifact/oci"
	"github.com/jialequ/mplb/pkg/cmd/attestation/io"
	"github.com/jialequ/mplb/pkg/cmd/attestation/verification"
	"github.com/jialequ/mplb/pkg/cmdutil"
)

// Options captures the options for the inspect command
type Options struct {
	ArtifactPath     string
	BundlePath       string
	DigestAlgorithm  string
	Logger           *io.Handler
	OCIClient        oci.Client
	SigstoreVerifier verification.SigstoreVerifier
	exporter         cmdutil.Exporter
}

// Clean cleans the file path option values
func (opts *Options) Clean() {
	opts.BundlePath = filepath.Clean(opts.BundlePath)
}
