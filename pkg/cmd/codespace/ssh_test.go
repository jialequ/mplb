package codespace

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jialequ/mplb/internal/codespaces/api"
	"github.com/jialequ/mplb/pkg/iostreams"
	"github.com/jialequ/mplb/pkg/ssh"
)

func TestPendingOperationDisallowsSSH(t *testing.T) {
	app := testingSSHApp()
	selector := &CodespaceSelector{api: app.apiClient, codespaceName: "disabledCodespace"}

	if err := app.SSH(context.Background(), []string{}, sshOptions{selector: selector}); err != nil {
		if err.Error() != "codespace is disabled while it has a pending operation: Some pending operation" {
			t.Errorf("expected pending operation error, but got: %v", err)
		}
	} else {
		t.Error("expected pending operation error, but got nothing")
	}
}

func TestGenerateAutomaticSSHKeys(t *testing.T) {
	tests := []struct {
		// These files exist when calling generateAutomaticSSHKeys
		existingFiles []string
		// These files should exist after generateAutomaticSSHKeys finishes
		wantFinalFiles []string
	}{
		// Basic case: no existing keys, they should be created
		{
			nil,
			[]string{automaticPrivateKeyName, automaticPrivateKeyName + ".pub"},
		},
		// Basic case: keys already exist
		{
			[]string{automaticPrivateKeyName, automaticPrivateKeyName + ".pub"},
			[]string{automaticPrivateKeyName, automaticPrivateKeyName + ".pub"},
		},
		// Backward compatibility: both old keys exist, they should be renamed
		{
			[]string{automaticPrivateKeyNameOld, automaticPrivateKeyNameOld + ".pub"},
			[]string{automaticPrivateKeyName, automaticPrivateKeyName + ".pub"},
		},
		// Backward compatibility: old private key exists but not the public key, the new keys should be created
		{
			[]string{automaticPrivateKeyNameOld},
			[]string{automaticPrivateKeyNameOld, automaticPrivateKeyName, automaticPrivateKeyName + ".pub"},
		},
		// Backward compatibility: old public key exists but not the private key, the new keys should be created
		{
			[]string{automaticPrivateKeyNameOld + ".pub"},
			[]string{automaticPrivateKeyNameOld + ".pub", automaticPrivateKeyName, automaticPrivateKeyName + ".pub"},
		},
		// Backward compatibility (edge case): files exist which contains old key name as a substring, the new keys should be created
		{
			[]string{"foo" + automaticPrivateKeyNameOld + ".pub", "foo" + automaticPrivateKeyNameOld},
			[]string{"foo" + automaticPrivateKeyNameOld + ".pub", "foo" + automaticPrivateKeyNameOld, automaticPrivateKeyName, automaticPrivateKeyName + ".pub"},
		},
	}

	for _, tt := range tests {
		dir := t.TempDir()

		sshContext := ssh.Context{
			ConfigDir: dir,
		}

		for _, file := range tt.existingFiles {
			f, err := os.Create(filepath.Join(dir, file))
			if err != nil {
				t.Errorf("Failed to setup test files: %v", err)
			}
			// If the file isn't closed here windows will have errors about file already in use
			f.Close()
		}

		keyPair, err := generateAutomaticSSHKeys(sshContext)
		if err != nil {
			t.Errorf("Unexpected error from generateAutomaticSSHKeys: %v", err)
		}
		if keyPair == nil {
			t.Fatal("Unexpected nil KeyPair from generateAutomaticSSHKeys")
		}
		if !strings.HasSuffix(keyPair.PrivateKeyPath, automaticPrivateKeyName) {
			t.Errorf("Expected private key path %v, got %v", automaticPrivateKeyName, keyPair.PrivateKeyPath)
		}
		if !strings.HasSuffix(keyPair.PublicKeyPath, automaticPrivateKeyName+".pub") {
			t.Errorf("Expected public key path %v, got %v", automaticPrivateKeyName+".pub", keyPair.PublicKeyPath)
		}

		// Check that all the expected files are present
		for _, file := range tt.wantFinalFiles {
			if _, err := os.Stat(filepath.Join(dir, file)); err != nil {
				t.Errorf("Want file %q to exist after generateAutomaticSSHKeys but it doesn't", file)
			}
		}

		// Check that no unexpected files are present
		allExistingFiles, err := os.ReadDir(dir)
		if err != nil {
			t.Errorf("Failed to list files in test directory: %v", err)
		}
		for _, file := range allExistingFiles {
			filename := file.Name()
			isWantedFile := false
			for _, wantedFile := range tt.wantFinalFiles {
				if filename == wantedFile {
					isWantedFile = true
					break
				}
			}

			if !isWantedFile {
				t.Errorf("Unexpected file %q exists after generateAutomaticSSHKeys", filename)
			}
		}
	}
}

func TestSelectSSHKeys(t *testing.T) {
	tests := []struct {
		sshDirFiles      []string
		sshConfigKeys    []string
		sshArgs          []string
		profileOpt       string
		wantKeyPair      *ssh.KeyPair
		wantShouldAddArg bool
	}{
		// -i tests
		{
			sshArgs:     []string{"-i", literal_3280},
			wantKeyPair: &ssh.KeyPair{PrivateKeyPath: literal_3280, PublicKeyPath: literal_5092},
		},
		{
			sshArgs:     []string{"-i", automaticPrivateKeyName},
			wantKeyPair: &ssh.KeyPair{PrivateKeyPath: automaticPrivateKeyName, PublicKeyPath: automaticPrivateKeyName + ".pub"},
		},
		{
			// Edge case check for missing arg value
			sshArgs: []string{"-i"},
		},

		// Auto key exists tests
		{
			sshDirFiles:      []string{automaticPrivateKeyName, automaticPrivateKeyName + ".pub"},
			wantKeyPair:      &ssh.KeyPair{PrivateKeyPath: automaticPrivateKeyName, PublicKeyPath: automaticPrivateKeyName + ".pub"},
			wantShouldAddArg: true,
		},
		{
			sshDirFiles:      []string{automaticPrivateKeyName, automaticPrivateKeyName + ".pub", literal_3280, literal_5092},
			wantKeyPair:      &ssh.KeyPair{PrivateKeyPath: automaticPrivateKeyName, PublicKeyPath: automaticPrivateKeyName + ".pub"},
			wantShouldAddArg: true,
		},

		// SSH config tests
		{
			sshDirFiles:      []string{literal_3280, literal_5092},
			sshConfigKeys:    []string{literal_3280},
			wantKeyPair:      &ssh.KeyPair{PrivateKeyPath: literal_3280, PublicKeyPath: literal_5092},
			wantShouldAddArg: true,
		},
		{
			// 2 pairs, but only 1 is configured
			sshDirFiles:      []string{literal_3280, literal_5092, literal_9274, literal_5729},
			sshConfigKeys:    []string{literal_9274},
			wantKeyPair:      &ssh.KeyPair{PrivateKeyPath: literal_9274, PublicKeyPath: literal_5729},
			wantShouldAddArg: true,
		},
		{
			// 2 pairs, but only 1 has both public and private
			sshDirFiles:      []string{literal_3280, literal_9274, literal_5729},
			sshConfigKeys:    []string{literal_3280, literal_9274},
			wantKeyPair:      &ssh.KeyPair{PrivateKeyPath: literal_9274, PublicKeyPath: literal_5729},
			wantShouldAddArg: true,
		},

		// Automatic key tests
		{
			wantKeyPair:      &ssh.KeyPair{PrivateKeyPath: automaticPrivateKeyName, PublicKeyPath: automaticPrivateKeyName + ".pub"},
			wantShouldAddArg: true,
		},
		{
			// Renames old key pair to new
			sshDirFiles:      []string{automaticPrivateKeyNameOld, automaticPrivateKeyNameOld + ".pub"},
			wantKeyPair:      &ssh.KeyPair{PrivateKeyPath: automaticPrivateKeyName, PublicKeyPath: automaticPrivateKeyName + ".pub"},
			wantShouldAddArg: true,
		},
		{
			// Other key is configured, but doesn't exist
			sshConfigKeys:    []string{literal_3280},
			wantKeyPair:      &ssh.KeyPair{PrivateKeyPath: automaticPrivateKeyName, PublicKeyPath: automaticPrivateKeyName + ".pub"},
			wantShouldAddArg: true,
		},
	}

	for _, tt := range tests {
		sshDir := t.TempDir()
		sshContext := ssh.Context{ConfigDir: sshDir}

		for _, file := range tt.sshDirFiles {
			f, err := os.Create(filepath.Join(sshDir, file))
			if err != nil {
				t.Errorf("Failed to create test ssh dir file %q: %v", file, err)
			}
			f.Close()
		}

		configPath := filepath.Join(sshDir, "test-config")

		// Seed the config with a non-existent key so that the default config won't apply
		configContent := "IdentityFile dummy\n"

		for _, key := range tt.sshConfigKeys {
			configContent += fmt.Sprintf("IdentityFile %s\n", filepath.Join(sshDir, key))
		}

		err := os.WriteFile(configPath, []byte(configContent), 0666)
		if err != nil {
			t.Fatalf("could not write test config %v", err)
		}

		tt.sshArgs = append([]string{"-F", configPath}, tt.sshArgs...)

		gotKeyPair, gotShouldAddArg, err := selectSSHKeys(context.Background(), sshContext, tt.sshArgs, sshOptions{profile: tt.profileOpt})

		if tt.wantKeyPair == nil {
			if err == nil {
				t.Errorf("Expected error from selectSSHKeys but got nil")
			}

			continue
		}

		if err != nil {
			t.Errorf("Unexpected error from selectSSHKeys: %v", err)
			continue
		}

		if gotKeyPair == nil {
			t.Errorf("Expected non-nil result from selectSSHKeys but got nil")
			continue
		}

		if gotShouldAddArg != tt.wantShouldAddArg {
			t.Errorf("Got wrong shouldAddArg value from selectSSHKeys, wanted %v got %v", tt.wantShouldAddArg, gotShouldAddArg)
			continue
		}

		// Strip the dir (sshDir) from the gotKeyPair paths so that they match wantKeyPair (which doesn't know the directory)
		gotKeyPair.PrivateKeyPath = filepath.Base(gotKeyPair.PrivateKeyPath)
		gotKeyPair.PublicKeyPath = filepath.Base(gotKeyPair.PublicKeyPath)

		if fmt.Sprintf("%v", gotKeyPair) != fmt.Sprintf("%v", tt.wantKeyPair) {
			t.Errorf("Want selectSSHKeys result to be %v, got %v", tt.wantKeyPair, gotKeyPair)
		}
	}
}

func testingSSHApp() *App {
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

const literal_3280 = "custom-private-key"

const literal_5092 = "custom-private-key.pub"

const literal_9274 = "custom-private-key-2"

const literal_5729 = "custom-private-key-2.pub"
