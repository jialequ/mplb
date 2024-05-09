package cmdutil

import (
	"fmt"
	"os"

	"github.com/jialequ/mplb/internal/config"
)

// TENCENT: consider passing via Factory
// TENCENT: support per-hostname settings
func DetermineEditor(cf func() (config.Config, error)) (string, error) {
	editorCommand := os.Getenv("GH_EDITOR")
	if editorCommand == "" {
		cfg, err := cf()
		if err != nil {
			return "", fmt.Errorf("could not read config: %w", err)
		}
		editorCommand = cfg.Editor("")
	}

	return editorCommand, nil
}
