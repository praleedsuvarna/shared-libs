//go:build !nosecretmanager
// +build !nosecretmanager

package config

import (
	"context"
	"fmt"
	"time"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
)

// getSecretFromGoogleSecretManager retrieves a secret from Google Cloud Secret Manager
func getSecretFromGoogleSecretManager(projectID, secretName string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create Secret Manager client
	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to create secret manager client: %v", err)
	}
	defer client.Close()

	// Build the request using the correct type
	req := &secretmanagerpb.AccessSecretVersionRequest{
		Name: fmt.Sprintf("projects/%s/secrets/%s/versions/latest", projectID, secretName),
	}

	// Call the API
	result, err := client.AccessSecretVersion(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to access secret %s: %v", secretName, err)
	}

	// Extract the secret data
	return string(result.Payload.Data), nil
}
