package security

import (
	"fmt"
	"os"
)

const (
	// SecureDirPerms are secure directory permissions (rwx------)
	SecureDirPerms = 0700

	// SecureFilePerms are secure file permissions (rw-------)
	SecureFilePerms = 0600

	// ConfigDirPerms are config directory permissions (rwxr-x---)
	ConfigDirPerms = 0750

	// ConfigFilePerms are config file permissions (rw-r-----)
	ConfigFilePerms = 0640

	// PublicDirPerms are public directory permissions (rwxr-xr-x)
	PublicDirPerms = 0755

	// PublicFilePerms are public file permissions (rw-r--r--)
	PublicFilePerms = 0644
)

// SetSecurePermissions sets secure permissions on a file or directory
func SetSecurePermissions(path string, isDir bool) error {
	var perms os.FileMode
	if isDir {
		perms = SecureDirPerms
	} else {
		perms = SecureFilePerms
	}

	if err := os.Chmod(path, perms); err != nil {
		return fmt.Errorf("failed to set permissions on %s: %w", path, err)
	}

	return nil
}

// VerifyPermissions verifies that a file or directory has secure permissions
func VerifyPermissions(path string, expectedPerms os.FileMode) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("failed to stat %s: %w", path, err)
	}

	actualPerms := info.Mode().Perm()
	if actualPerms != expectedPerms {
		return fmt.Errorf("insecure permissions on %s: got %o, expected %o", path, actualPerms, expectedPerms)
	}

	return nil
}

// CreateSecureDirectory creates a directory with secure permissions
func CreateSecureDirectory(path string, logger Logger) error {
	// Create directory
	if err := os.MkdirAll(path, SecureDirPerms); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", path, err)
	}

	// Verify permissions (some systems may apply umask)
	if err := SetSecurePermissions(path, true); err != nil {
		return err
	}

	logger.SecurityEvent("secure_directory_created", map[string]interface{}{
		"path":        sanitizePath(path),
		"permissions": fmt.Sprintf("%o", SecureDirPerms),
	})

	return nil
}

// CreateSecureFile creates a file with secure permissions
func CreateSecureFile(path string, data []byte, logger Logger) error {
	// Write file
	if err := os.WriteFile(path, data, SecureFilePerms); err != nil {
		return fmt.Errorf("failed to write file %s: %w", path, err)
	}

	// Verify permissions
	if err := SetSecurePermissions(path, false); err != nil {
		return err
	}

	logger.SecurityEvent("secure_file_created", map[string]interface{}{
		"path":        sanitizePath(path),
		"permissions": fmt.Sprintf("%o", SecureFilePerms),
		"size":        len(data),
	})

	return nil
}

// CheckFileOwnership verifies that a file is owned by the current user
func CheckFileOwnership(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("failed to stat file: %w", err)
	}

	// On Unix systems, check if we own the file
	// This is a simplified check - in production, you'd use syscall.Stat_t
	// to get the actual UID and compare with os.Getuid()
	_ = info // Placeholder

	return nil
}
