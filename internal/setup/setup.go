package setup

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/lgrees/resy-cli/internal/utils/paths"
	"github.com/spf13/viper"
)

// SurveyConfig reads the repo-local internal/setup/secret.yml and persists the
// credentials to the application's config path (GetAppPaths().ConfigFilePath)
// so other commands can find them.
func SurveyConfig() error {
	// 1) read source file (internal/setup/secret.yml)
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("unable to resolve working dir: %w", err)
	}
	srcPath := filepath.Join(cwd, "internal", "setup", "secret.yml")

	src := viper.New()
	src.SetConfigFile(srcPath)
	src.SetConfigType("yaml")
	if err := src.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to read %s: %w", srcPath, err)
	}

	apiKey := src.GetString("resy_api_key")
	authToken := src.GetString("resy_auth_token")
	if apiKey == "" || authToken == "" {
		return fmt.Errorf("%s is missing required keys: resy_api_key and/or resy_auth_token", srcPath)
	}

	// 2) determine destination config path from GetAppPaths()
	p, err := paths.GetAppPaths()
	if err != nil {
		return fmt.Errorf("unable to determine app paths: %w", err)
	}
	destDir := filepath.Dir(p.ConfigFilePath)
	if destDir == "" {
		return fmt.Errorf("invalid destination directory for config: %s", p.ConfigFilePath)
	}

	// ensure directory exists with secure perms
	if err := os.MkdirAll(destDir, 0o700); err != nil {
		return fmt.Errorf("failed to create config dir %s: %w", destDir, err)
	}

	// 3) write atomically to dest (temp -> rename)
	tmpFile := p.ConfigFilePath + ".tmp.yml"

	dst := viper.New()
	dst.Set("resy_api_key", apiKey)
	dst.Set("resy_auth_token", authToken)
	dst.SetConfigFile(tmpFile)
	dst.SetConfigType("yaml")

	// WriteConfigAs will create/overwrite tmpFile with YAML content
	if err := dst.WriteConfigAs(tmpFile); err != nil {
		// fallback: try writing manually using viper.WriteConfig (less likely needed)
		_ = os.Remove(tmpFile)
		return fmt.Errorf("failed to write temp config %s: %w", tmpFile, err)
	}

	// set strict file perms (override umask if necessary)
	if err := os.Chmod(tmpFile, 0o600); err != nil {
		// non-fatal, but warn
		fmt.Fprintf(os.Stderr, "warning: failed to chmod %s: %v\n", tmpFile, err)
	}

	// move into place
	if err := os.Rename(tmpFile, p.ConfigFilePath); err != nil {
		_ = os.Remove(tmpFile)
		return fmt.Errorf("failed to install config %s -> %s: %w", tmpFile, p.ConfigFilePath, err)
	}

	// 4) ensure final file perms
	if err := os.Chmod(p.ConfigFilePath, 0o600); err != nil {
		// non-fatal; warn
		fmt.Fprintf(os.Stderr, "warning: failed to set perms on %s: %v\n", p.ConfigFilePath, err)
	}

	// 5) populate global viper runtime so current process sees the values immediately
	viper.Set("resy_api_key", apiKey)
	viper.Set("resy_auth_token", authToken)

	fmt.Printf("Loaded Resy credentials from %s and persisted to %s\n", srcPath, p.ConfigFilePath)
	return nil
}
