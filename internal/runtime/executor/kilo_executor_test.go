package executor

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/router-for-me/CLIProxyAPI/v6/internal/config"
	cliproxyauth "github.com/router-for-me/CLIProxyAPI/v6/sdk/cliproxy/auth"
)

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestFetchKiloModelsPreservesAllReturnedModels(t *testing.T) {
	t.Parallel()

	ctx := context.WithValue(context.Background(), "cliproxy.roundtripper", roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		if req.URL.String() != "https://api.kilo.ai/api/openrouter/models" {
			t.Fatalf("unexpected request URL %q", req.URL.String())
		}
		if got := req.Header.Get("Authorization"); got != "Bearer kilo-token" {
			t.Fatalf("authorization header = %q, want %q", got, "Bearer kilo-token")
		}
		if got := req.Header.Get("X-Kilocode-OrganizationID"); got != "org-123" {
			t.Fatalf("organization header = %q, want %q", got, "org-123")
		}

		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body: io.NopCloser(strings.NewReader(`{"data":[{"id":"anthropic/claude-3-7-sonnet","name":"Claude Sonnet 3.7","context_length":200000},{"id":"openai/gpt-4.1","name":"GPT-4.1","context_length":128000},{"id":"google/gemini-2.5-pro","name":"Gemini 2.5 Pro","context_length":1048576}]}`)),
		}, nil
	}))

	auth := &cliproxyauth.Auth{
		Provider: "kilo",
		Metadata: map[string]any{
			"kilocodeToken":  "kilo-token",
			"organization_id": "org-123",
		},
	}

	models := FetchKiloModels(ctx, auth, &config.Config{})

	wantIDs := []string{
		"kilo/auto",
		"anthropic/claude-3-7-sonnet",
		"openai/gpt-4.1",
		"google/gemini-2.5-pro",
	}

	if len(models) != len(wantIDs) {
		t.Fatalf("model count = %d, want %d (%#v)", len(models), len(wantIDs), models)
	}

	for i, wantID := range wantIDs {
		if got := models[i].ID; got != wantID {
			t.Fatalf("models[%d].ID = %q, want %q", i, got, wantID)
		}
	}
}
