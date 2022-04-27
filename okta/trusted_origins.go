package okta

import (
	"fmt"

	"github.com/okta/okta-sdk-golang/v2/okta"
	"github.com/okta/okta-sdk-golang/v2/okta/query"
)

func CreateTrustedOrigin(origin string) error {
	ctx, client := getContextAndClient()

	trustedOrigin := &okta.TrustedOrigin{
		Name:   origin,
		Origin: origin,
		Scopes: []*okta.Scope{{Type: "CORS"}, {Type: "REDIRECT"}},
	}

	_, _, err := client.TrustedOrigin.CreateOrigin(ctx, *trustedOrigin)
	if err != nil {
		return fmt.Errorf("failed to create trusted origin")
	}

	return nil
}

func DeleteTrustedOrigin(origin string) error {
	ctx, client := getContextAndClient()

	filter := query.NewQueryParams(query.WithFilter(fmt.Sprintf("origin eq %q", origin)), query.WithLimit(1))
	origins, _, err := client.TrustedOrigin.ListOrigins(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to get trusted origin %q for deletion: %w", origin, err)
	}

	_, err = client.TrustedOrigin.DeleteOrigin(ctx, origins[0].Id)
	if err != nil {
		return fmt.Errorf("failed to delete trusted origin %q: %w", origin, err)
	}

	return nil
}
