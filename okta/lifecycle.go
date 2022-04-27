package okta

import (
	"encoding/json"
	"fmt"
	"io"
)

// clientResponse actually contains more information, but we only need the new secret.
type clientResponse struct {
	ClientSecret string `json:"client_secret"`
}

func NewSecret(clientID string) (string, error) {
	ctx, client := getContextAndClient()

	url := fmt.Sprintf("/oauth2/v1/clients/%s/lifecycle/newSecret", clientID)
	req, err := client.CloneRequestExecutor().NewRequest("POST", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create a new secret rotation request for client ID %q: %w", clientID, err)
	}

	resp, err := client.CloneRequestExecutor().Do(ctx, req, nil)
	if err != nil {
		return "", fmt.Errorf("failed to rotate secret for client ID %q: %w", clientID, err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to rotate secret for client ID %q; error reading response body: %w", clientID, err)
	}

	var c clientResponse
	err = json.Unmarshal(body, &c)
	if err != nil {
		return "", fmt.Errorf("failed to rotate secret for client ID %q; error parsing response body: %w", clientID, err)
	}

	return c.ClientSecret, nil
}
