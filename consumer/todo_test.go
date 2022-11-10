package consumer

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"testing"

	"github.com/pact-foundation/pact-go/dsl"
)

func TestConsumer(t *testing.T) {
	type TODO struct {
		ID    string `json:"id" pact:"example=1"`
		Title string `json:"title" pact:"example=task"`
		Done  bool   `json:"done" pact:"example=false"`
	}

	// Create Pact connecting to local Daemon
	pact := &dsl.Pact{
		Consumer: "pact-sandbox-consumer",
		Provider: "pact-sandbox-provider",
		Host:     "localhost",
	}
	defer pact.Teardown()

	// Pass in test case
	var test = func() error {
		u := fmt.Sprintf("http://localhost:%d/todo", pact.Server.Port)
		req, err := http.NewRequest(http.MethodPost, u, strings.NewReader(`{"title":"歯を磨く"}`))
		if err != nil {
			return err
		}

		// NOTE: by default, request bodies are expected to be sent with a Content-Type
		// of application/json. If you don't explicitly set the content-type, you
		// will get a mismatch during Verification.
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer test-user-token")

		if _, err = http.DefaultClient.Do(req); err != nil {
			return err
		}

		return err
	}

	// Set up our expected interactions.
	pact.
		AddInteraction().
		Given("exists test-user").
		UponReceiving("A create to todo").
		WithRequest(dsl.Request{
			Method:  http.MethodPost,
			Path:    dsl.String("/todo"),
			Headers: dsl.MapMatcher{"Content-Type": dsl.String("application/json"), "Authorization": dsl.String("Bearer test-user-token")},
			Body: map[string]string{
				"title": "歯を磨く",
			},
		}).
		WillRespondWith(dsl.Response{
			Status:  201,
			Headers: dsl.MapMatcher{"Content-Type": dsl.String("application/json")},
			Body: dsl.Match(&TODO{
				ID:    "1",
				Title: "歯を磨く",
				Done:  false,
			}),
		})

	// Verify
	if err := pact.Verify(test); err != nil {
		log.Fatalf("Error on Verify: %v", err)
	}
}
