package browseruse_test

import (
	"context"
	"net/http"
	"os"

	browseruse "github.com/psyb0t/go-browser-use"
)

// This README example requires a real API key, so it only compiles in tests.
//
//nolint:testableexamples // Requires provider credentials; never run in tests.
func Example() {
	auth := browseruse.WithRequestEditorFn(
		func(_ context.Context, req *http.Request) error {
			req.Header.Set(
				"X-Browser-Use-API-Key",
				os.Getenv("BROWSER_USE_API_KEY"),
			)

			return nil
		},
	)

	client, err := browseruse.NewClientWithResponses(
		"https://api.browser-use.com/api/v4",
		auth,
	)
	if err != nil {
		panic(err)
	}

	res, err := client.ListRunsRunsGetWithResponse(
		context.Background(),
		&browseruse.ListRunsRunsGetParams{},
	)
	if err != nil {
		panic(err)
	}

	if res.JSON200 == nil {
		panic(res.Status())
	}

	_ = res.JSON200.Runs
}
