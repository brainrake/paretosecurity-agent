package checks

import (
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/crypto/ssh"
)

type SSHKeys struct {
	passed     bool
	failedKeys []string
}

// Name returns the name of the check
func (f *SSHKeys) Name() string {
	return "SSH keys have password protection"
}

// checks if private key has password protection
func (f *SSHKeys) hasPassword(privateKeyPath string) bool {
	keyBytes, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return true // assume secure if can't read
	}

	_, err = ssh.ParsePrivateKey(keyBytes)
	return err != nil // if error occurs, key likely has password or it's FIDO2 managed key
}

// Run executes the check
func (f *SSHKeys) Run() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	sshDir := filepath.Join(home, ".ssh")

	files, err := os.ReadDir(sshDir)
	if err != nil {
		f.passed = true
		return nil
	}

	f.passed = true
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".pub") {
			privateKeyPath := filepath.Join(sshDir, strings.TrimSuffix(file.Name(), ".pub"))
			if _, err := os.Stat(privateKeyPath); err == nil {
				if !f.hasPassword(privateKeyPath) {
					f.passed = false
					f.failedKeys = append(f.failedKeys, file.Name())
				}
			}
		}
	}

	return nil
}

// Passed returns the status of the check
func (f *SSHKeys) Passed() bool {
	return f.passed
}

// CanRun returns whether the check can run
func (f *SSHKeys) IsRunnable() bool {
	home, err := os.UserHomeDir()
	if err != nil {
		return false
	}

	sshPath := filepath.Join(home, ".ssh")
	if _, err := os.Stat(sshPath); os.IsNotExist(err) {
		return false
	}

	//check if there are any private keys in the .ssh directory
	files, err := os.ReadDir(sshPath)
	if err != nil {
		return false
	}

	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".pub") {
			privateKeyPath := filepath.Join(sshPath, strings.TrimSuffix(file.Name(), ".pub"))
			if _, err := os.Stat(privateKeyPath); err == nil {
				return true
			}
		}
	}
	return false

}

// UUID returns the UUID of the check
func (f *SSHKeys) UUID() string {
	return "ef69f752-0e89-46e2-a644-310429ae5f45"
}

// ReportIfDisabled returns whether the check should report if it is disabled
func (f *SSHKeys) ReportIfDisabled() bool {
	return false
}

// PassedMessage returns the message to return if the check passed
func (f *SSHKeys) PassedMessage() string {
	return "SSH keys are password protected"
}

// FailedMessage returns the message to return if the check failed
func (f *SSHKeys) FailedMessage() string {
	return "SSH keys are not using password"
}

// RequiresRoot returns whether the check requires root access
func (f *SSHKeys) RequiresRoot() bool {
	return false
}

// Status returns the status of the check
func (f *SSHKeys) Status() string {
	if f.Passed() {
		return f.PassedMessage()
	}
	return "Found unprotected SSH key(s): " + strings.Join(f.failedKeys, ", ")
}
