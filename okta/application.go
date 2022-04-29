package okta

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/okta/okta-sdk-golang/v2/okta"
	"github.com/okta/okta-sdk-golang/v2/okta/query"
)

// Application described an Okta application without exposing Okta types outside of this package.
type Application struct {
	ID           string
	ClientID     string
	ClientSecret string
}

func CreateApplicationGroupAssignment(app *Application, groupID string) error {
	ctx, client := getContextAndClient()

	assignments, _, err := client.Application.GetApplicationGroupAssignment(ctx, app.ID, groupID, nil)
	if err != nil {
		return fmt.Errorf("failed to get application group assignment for application %q and group %q: %w", app.ID, groupID, err)
	}

	if assignments == nil {
		_, _, err := client.Application.CreateApplicationGroupAssignment(ctx, app.ID, groupID, okta.ApplicationGroupAssignment{})
		if err != nil {
			return fmt.Errorf("failed to create application group assignment for application %q and group %q: %w", app.ID, groupID, err)
		}
	}

	return nil
}

func GetApplicationByLabel(label string) (*Application, error) {
	ctx, client := getContextAndClient()

	filter := query.NewQueryParams(query.WithQ(label), query.WithLimit(1))
	apps, resp, err := client.Application.ListApplications(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("error getting client ID for label %q; API error: %w", label, err)
	}

	// Not found.
	if len(apps) == 0 {
		return nil, nil
	}

	var oidcApps []*okta.OpenIdConnectApplication
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error getting client ID for label %q; error reading response body: %w", label, err)
	}

	err = json.Unmarshal(body, &oidcApps)
	if err != nil {
		return nil, fmt.Errorf("error getting client ID for label %q; error parsing response body: %w", label, err)
	}

	return &Application{
		ID:           oidcApps[0].Id,
		ClientID:     oidcApps[0].Credentials.OauthClient.ClientId,
		ClientSecret: oidcApps[0].Credentials.OauthClient.ClientSecret, // This is not returned by the Okta API!
	}, nil
}

// CreateApplication in Okta and return it.
func CreateApplication(label string) (*Application, error) {
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
		return nil, fmt.Errorf("error creating app: %w", err)
	}

	var oidcApp *okta.OpenIdConnectApplication
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error creating app %q; error reading response body: %w", label, err)
	}

	err = json.Unmarshal(body, &oidcApp)
	if err != nil {
		return nil, fmt.Errorf("error creating app %q; error parsing response body: %w", label, err)
	}

	return &Application{
		ID:           oidcApp.Id,
		ClientID:     oidcApp.Credentials.OauthClient.ClientId,
		ClientSecret: oidcApp.Credentials.OauthClient.ClientSecret,
	}, err
}

func DeleteApplication(app *Application) error {
	ctx, client := getContextAndClient()

	_, err := client.Application.DeactivateApplication(ctx, app.ID)
	if err != nil {
		return fmt.Errorf("error deactivating application %q: %w", app.ID, err)
	}

	_, err = client.Application.DeleteApplication(ctx, app.ID)
	if err != nil {
		return fmt.Errorf("error deleting application %q: %w", app.ID, err)
	}

	return nil
}
