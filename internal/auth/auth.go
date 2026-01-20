package auth

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

// Authenticator handles authentication via htpasswd files
type Authenticator struct {
	htpasswdPath string
}

// NewAuthenticator creates a new authenticator with the given htpasswd file path
func NewAuthenticator(htpasswdPath string) *Authenticator {
	return &Authenticator{
		htpasswdPath: htpasswdPath,
	}
}

// Authenticate checks if the username and password match an entry in the htpasswd file
func (a *Authenticator) Authenticate(username, password string) (bool, error) {
	file, err := os.Open(a.htpasswdPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("failed to open htpasswd file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}

		fileUsername := parts[0]
		filePassword := parts[1]

		if fileUsername == username {
			// Check if password matches (bcrypt hash)
			if err := bcrypt.CompareHashAndPassword([]byte(filePassword), []byte(password)); err == nil {
				return true, nil
			}
			return false, nil
		}
	}

	if err := scanner.Err(); err != nil {
		return false, fmt.Errorf("error reading htpasswd file: %w", err)
	}

	return false, nil
}

// Exists checks if the htpasswd file exists
func (a *Authenticator) Exists() bool {
	_, err := os.Stat(a.htpasswdPath)
	return err == nil
}

// CreateWithUser creates a new htpasswd file with the given user
func (a *Authenticator) CreateWithUser(username, password string) error {
	// Hash the password with bcrypt
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Create the htpasswd file
	file, err := os.Create(a.htpasswdPath)
	if err != nil {
		return fmt.Errorf("failed to create htpasswd file: %w", err)
	}
	defer file.Close()

	// Write the user entry
	_, err = fmt.Fprintf(file, "%s:%s\n", username, hashedPassword)
	if err != nil {
		return fmt.Errorf("failed to write htpasswd entry: %w", err)
	}

	return nil
}
