package main

import (
	"fmt"
	"testing"

	"github.com/pact-foundation/pact-go/dsl"
	"github.com/pact-foundation/pact-go/types"
)

func Test_pactUsecase(t *testing.T) {
	// Create Pact connecting to local Daemon
	pact := &dsl.Pact{
		Provider: "pact-sandbox-provider",
	}

	// Start provider API in the background
	go Run()

	// Verify the Provider using the locally saved Pact Files
	pact.VerifyProvider(t, types.VerifyRequest{
		ProviderBaseURL: fmt.Sprintf("http://localhost:%d", port),
		PactURLs: []string{
			"../consumer/pacts/pact-sandbox-consumer-pact-sandbox-provider.json",
		},
		StateHandlers: types.StateHandlers{
			// Setup any state required by the test
			// in this case, we ensure there is a "user" in the system
			// "User foo exists": func() error {
			// 	lastName = "crickets"
			// 	return nil
			// },
		},
	})
}
