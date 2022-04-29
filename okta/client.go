package okta

import (
	"context"
	"fmt"
	"sync"

	"github.com/okta/okta-sdk-golang/v2/okta"
)

var (
	initClient sync.Once

	// Do not use these two directly! Use getContextAndClient instead!
	ctx    context.Context
	client *okta.Client
)

// getContextAndClient to use the Okta API.
func getContextAndClient() (context.Context, *okta.Client) {
	initClient.Do(func() {
		var err error
		ctx, client, err = okta.NewClient(
			context.Background(),
			okta.WithCache(false),
		)
		if err != nil {
			panic(fmt.Errorf("error while initializing Okta client: %w", err))
		}
	})

	return ctx, client
}

// _true returns a pointer to a boolean with the value true.
func _true() *bool {
	b := true
	return &b
}
