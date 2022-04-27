package okta

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/okta/okta-sdk-golang/v2/okta"
	"github.com/okta/okta-sdk-golang/v2/okta/query"
)

func GetClientIDByLabel(label string) (string, error) {
	ctx, client := getContextAndClient()

	filter := query.NewQueryParams(query.WithQ(label), query.WithLimit(1))
	apps, resp, err := client.Application.ListApplications(ctx, filter)
	if err != nil {
		return "", fmt.Errorf("error getting client ID for label %q; API error: %w", label, err)
	}

	// Not found.
	if len(apps) == 0 {
		return "", nil
	}

	var oidcApps []*okta.OpenIdConnectApplication
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error getting client ID for label %q; error reading response body: %w", label, err)
	}

	err = json.Unmarshal(body, &oidcApps)
	if err != nil {
		return "", fmt.Errorf("error getting client ID for label %q; error parsing response body: %w", label, err)
	}

	return oidcApps[0].Credentials.OauthClient.ClientId, nil
}

// CreateApp in Okta and return the client ID and the client secret.
func CreateApp(label string) (string, string, error) {
	ctx, client := getContextAndClient()

	app := okta.NewOpenIdConnectApplication()
	app.Label = label
	app.Credentials = &okta.OAuthApplicationCredentials{
		OauthClient: &okta.ApplicationCredentialsOAuthClient{
			AutoKeyRotation:         _true(),
			TokenEndpointAuthMethod: "client_secret_post",
		},
	}

	responseType := okta.OAuthResponseType("code")
	grantTypeRefreshToken := okta.OAuthGrantType("refresh_token")
	grantTypeAuthorizationCode := okta.OAuthGrantType("authorization_code")
	app.Settings = &okta.OpenIdConnectApplicationSettings{
		OauthClient: &okta.OpenIdConnectApplicationSettingsClient{
			ClientUri: fmt.Sprintf("https://%s.zageno.com", label),
			LogoUri:   "",
			RedirectUris: []string{
				fmt.Sprintf("https://%s.zageno.com/oauth2/callback", label),
				fmt.Sprintf("https://%s-admin.zageno.com/oauth2/callback", label),
			},
			PostLogoutRedirectUris: []string{
				fmt.Sprintf("https://%s.zageno.com", label),
				fmt.Sprintf("https://%s-admin.zageno.com", label),
			},
			ResponseTypes:   []*okta.OAuthResponseType{&responseType},
			GrantTypes:      []*okta.OAuthGrantType{&grantTypeRefreshToken, &grantTypeAuthorizationCode},
			ConsentMethod:   "REQUIRED",
			ApplicationType: "web",
		},
	}

	_, resp, err := client.Application.CreateApplication(ctx, app, nil)
	if err != nil {
		return "", "", fmt.Errorf("error creating app: %w", err)
	}

	var oidcApp *okta.OpenIdConnectApplication
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", fmt.Errorf("error creating app %q; error reading response body: %w", label, err)
	}

	err = json.Unmarshal(body, &oidcApp)
	if err != nil {
		return "", "", fmt.Errorf("error creating app %q; error parsing response body: %w", label, err)
	}

	return oidcApp.Credentials.OauthClient.ClientId, oidcApp.Credentials.OauthClient.ClientSecret, err
}

func DeleteApp(label string) error {
	ctx, client := getContextAndClient()

	filter := query.NewQueryParams(query.WithQ(label), query.WithLimit(1))
	apps, _, err := client.Application.ListApplications(ctx, filter)
	if err != nil {
		return fmt.Errorf("error getting application for label %q for deletion: %w", label, err)
	}

	// Not found.
	if len(apps) == 0 {
		return nil
	}

	_, err = client.Application.DeactivateApplication(ctx, apps[0].(*okta.Application).Id)
	if err != nil {
		return fmt.Errorf("error deactivating application for label %q for deletion: %w", label, err)
	}

	_, err = client.Application.DeleteApplication(ctx, apps[0].(*okta.Application).Id)
	if err != nil {
		return fmt.Errorf("error deleting application for label %q: %w", label, err)
	}

	return nil
}
