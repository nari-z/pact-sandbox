package consumer

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"testing"

	"github.com/pact-foundation/pact-go/dsl"
)

type TODO struct {
	ID    string `json:"id" pact:"example=1"`
	Title string `json:"title" pact:"example=task"`
	Done  bool   `json:"done" pact:"example=false"`
}

func Test_todoBacicUsecase(t *testing.T) {
	// Create Pact connecting to local Daemon
	pact := &dsl.Pact{
		Consumer: "pact-sandbox-consumer",
		Provider: "pact-sandbox-provider",
		Host:     "localhost",
	}
	defer pact.Teardown()

	// Pass in test case
	var test = func() error {
		todo, err := addTODO(pact, "歯を磨く")
		if err != nil {
			t.Fatalf(`failed to add todo: %v`, err)
		}

		if err := finishTODO(pact, todo.ID); err != nil {
			t.Fatalf(`failed to finish todo: %v`, err)
		}

		return nil
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
			Status:  http.StatusCreated,
			Headers: dsl.MapMatcher{"Content-Type": dsl.String("application/json")},
			Body: dsl.Match(&TODO{
				ID:    "1",
				Title: "歯を磨く",
				Done:  false,
			}),
		})
	pact.
		AddInteraction().
		Given("exists todo").
		UponReceiving("A finish to todo").
		WithRequest(dsl.Request{
			Method:  http.MethodPatch,
			Path:    dsl.String("/todo/1"),
			Headers: dsl.MapMatcher{"Content-Type": dsl.String("application/json"), "Authorization": dsl.String("Bearer test-user-token")},
			Body:    nil,
		}).
		WillRespondWith(dsl.Response{
			Status:  http.StatusOK,
			Headers: dsl.MapMatcher{"Content-Type": dsl.String("application/json")},
			Body:    nil,
		})

	// Verify
	if err := pact.Verify(test); err != nil {
		log.Fatalf("Error on Verify: %v", err)
	}
}

func addTODO(pact *dsl.Pact, title string) (TODO, error) {
	u := fmt.Sprintf("http://localhost:%d/todo", pact.Server.Port)
	req, err := http.NewRequest(http.MethodPost, u, strings.NewReader(`{"title":"歯を磨く"}`))
	if err != nil {
		return TODO{}, err
	}

	// NOTE: by default, request bodies are expected to be sent with a Content-Type
	// of application/json. If you don't explicitly set the content-type, you
	// will get a mismatch during Verification.
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-user-token")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return TODO{}, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return TODO{}, err
	}

	todo := TODO{}
	if err := json.Unmarshal(body, &todo); err != nil {
		return TODO{}, err
	}

	return todo, nil
}

func finishTODO(pact *dsl.Pact, id string) error {
	u := fmt.Sprintf("http://localhost:%d/todo/%s", pact.Server.Port, id)
	req, err := http.NewRequest(http.MethodPatch, u, nil)
	if err != nil {
		return err
	}

	// NOTE: by default, request bodies are expected to be sent with a Content-Type
	// of application/json. If you don't explicitly set the content-type, you
	// will get a mismatch during Verification.
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-user-token")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	return nil
}
