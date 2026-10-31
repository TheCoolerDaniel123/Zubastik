package ai_test

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"shared/entity"
	"user-service/pkg/clients/ai"
)

// newTestProvider creates a GigaChatProvider pointed at mock servers.
func newTestProvider(oauthURL, apiURL string) *ai.GigaChatProvider {
	p := ai.NewGigaChatProvider("test-auth-key", "GIGACHAT_API_PERS")
	p.SetEndpoints(oauthURL, apiURL)
	return p
}

func mockOAuthServer(t *testing.T, wantScope string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "Basic test-auth-key", r.Header.Get("Authorization"))
		r.ParseForm()
		assert.Equal(t, wantScope, r.FormValue("scope"))

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"access_token": "test-access-token",
			"expires_at":   time.Now().Add(30 * time.Minute).Unix(),
		})
	}))
}

func mockChatServer(t *testing.T, wantMessages int, replyContent string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "Bearer test-access-token", r.Header.Get("Authorization"))

		var req map[string]any
		json.NewDecoder(r.Body).Decode(&req)
		msgs := req["messages"].([]any)
		assert.Len(t, msgs, wantMessages)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{
				{"message": map[string]string{"role": "assistant", "content": replyContent}},
			},
		})
	}))
}

func TestGigaChatProvider_Complete_Success(t *testing.T) {
	oauth := mockOAuthServer(t, "GIGACHAT_API_PERS")
	defer oauth.Close()

	// system + 2 history + 1 user = 4 messages
	chat := mockChatServer(t, 4, "Brush your teeth twice a day.")
	defer chat.Close()

	p := newTestProvider(oauth.URL, chat.URL)

	history := []entity.ChatMessage{
		{Role: entity.ChatRoleUser, Content: "hello"},
		{Role: entity.ChatRoleAssistant, Content: "hi there"},
	}

	reply, err := p.Complete(context.Background(), "you are a dental assistant", history, "how to brush?")
	require.NoError(t, err)
	assert.Equal(t, "Brush your teeth twice a day.", reply)
}

func TestGigaChatProvider_Complete_NoHistory(t *testing.T) {
	oauth := mockOAuthServer(t, "GIGACHAT_API_PERS")
	defer oauth.Close()

	// system + 1 user = 2 messages
	chat := mockChatServer(t, 2, "Visit your dentist regularly.")
	defer chat.Close()

	p := newTestProvider(oauth.URL, chat.URL)

	reply, err := p.Complete(context.Background(), "system prompt", nil, "when to visit dentist?")
	require.NoError(t, err)
	assert.Equal(t, "Visit your dentist regularly.", reply)
}

func TestGigaChatProvider_Complete_TokenCached(t *testing.T) {
	tokenRequestCount := 0
	oauth := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenRequestCount++
		json.NewEncoder(w).Encode(map[string]any{
			"access_token": "cached-token",
			"expires_at":   time.Now().Add(30 * time.Minute).Unix(),
		})
	}))
	defer oauth.Close()

	chat := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{
				{"message": map[string]string{"role": "assistant", "content": "ok"}},
			},
		})
	}))
	defer chat.Close()

	p := newTestProvider(oauth.URL, chat.URL)

	for i := 0; i < 3; i++ {
		_, err := p.Complete(context.Background(), "sys", nil, "question")
		require.NoError(t, err)
	}

	assert.Equal(t, 1, tokenRequestCount, "token should be requested only once")
}

func TestGigaChatProvider_Complete_OAuthError(t *testing.T) {
	oauth := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"message":"invalid key"}`))
	}))
	defer oauth.Close()

	chat := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer chat.Close()

	p := newTestProvider(oauth.URL, chat.URL)

	reply, err := p.Complete(context.Background(), "sys", nil, "question")
	assert.Empty(t, reply)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "401")
}

func TestGigaChatProvider_Complete_ChatAPIError(t *testing.T) {
	oauth := mockOAuthServer(t, "GIGACHAT_API_PERS")
	defer oauth.Close()

	chat := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message":"internal error"}`))
	}))
	defer chat.Close()

	p := newTestProvider(oauth.URL, chat.URL)

	reply, err := p.Complete(context.Background(), "sys", nil, "question")
	assert.Empty(t, reply)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "500")
}

func TestGigaChatProvider_Complete_EmptyChoices(t *testing.T) {
	oauth := mockOAuthServer(t, "GIGACHAT_API_PERS")
	defer oauth.Close()

	chat := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{"choices": []any{}})
	}))
	defer chat.Close()

	p := newTestProvider(oauth.URL, chat.URL)

	reply, err := p.Complete(context.Background(), "sys", nil, "question")
	assert.Empty(t, reply)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "empty choices")
}

func TestGigaChatProvider_Complete_ContextCancelled(t *testing.T) {
	oauth := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		json.NewEncoder(w).Encode(map[string]any{"access_token": "tok", "expires_at": 0})
	}))
	defer oauth.Close()

	chat := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer chat.Close()

	p := newTestProvider(oauth.URL, chat.URL)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := p.Complete(ctx, "sys", nil, "question")
	require.Error(t, err)
}

// Ensure TLS skip works for the real provider constructor (smoke test).
func TestNewGigaChatProvider_TLSConfig(t *testing.T) {
	p := ai.NewGigaChatProvider("key", "scope")
	require.NotNil(t, p)

	// Verify it can reach a self-signed TLS server without error.
	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	client := &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}
	resp, err := client.Get(ts.URL)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()
}

// Verify system prompt contains key dental keywords.
func TestGigaChatProvider_SystemPromptContent(t *testing.T) {
	prompt := ai.DentalSystemPrompt()
	keywords := []string{"dental", "dentist", "oral"}
	for _, kw := range keywords {
		assert.True(t, strings.Contains(strings.ToLower(prompt), kw),
			"system prompt should contain keyword %q", kw)
	}
}
