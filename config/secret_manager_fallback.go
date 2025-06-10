//go:build nosecretmanager
// +build nosecretmanager

package config

import "fmt"

// getSecretFromGoogleSecretManager returns an error when Secret Manager is not available
func getSecretFromGoogleSecretManager(projectID, secretName string) (string, error) {
	return "", fmt.Errorf("Secret Manager not available - build with Secret Manager support or use basic mode")
}
