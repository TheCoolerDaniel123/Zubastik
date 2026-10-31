package ai

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	"shared/entity"
)

const (
	defaultOAuthURL = "https://ngw.devices.sberbank.ru:9443/api/v2/oauth"
	defaultAPIURL   = "https://gigachat.devices.sberbank.ru/api/v1/chat/completions"
	gigaChatModel   = "GigaChat"

	dentalSystemPrompt = `You are a dental health assistant built into the Zub dental health app.

Your exclusive purpose is to help users with topics related to dentistry and oral health, including:
- Oral hygiene tips and daily routines (brushing technique, flossing, mouthwash)
- Tooth care and prevention (diet, fluoride, sealants, sensitivity)
- Common dental conditions (cavities, gum disease, tooth decay, bad breath, etc.)
- Dental procedures and treatments (fillings, crowns, root canals, braces, implants, whitening, etc.)
- Post-treatment care and recovery
- Choosing dental products (toothbrush type, toothpaste, floss, mouthwash, etc.)
- Paediatric dentistry and dental development
- Knowing when to see a dentist

If a user asks about any topic not related to dentistry or oral health, politely decline and redirect them to dental topics.
Never provide a specific medical diagnosis. For any serious or urgent dental concern, always recommend visiting a licensed dentist.
Keep your responses clear, friendly, and concise.`
)

// DentalSystemPrompt returns the system prompt used for the dental AI assistant.
func DentalSystemPrompt() string { return dentalSystemPrompt }

type gigaChatTokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresAt   int64  `json:"expires_at"`
}

type gigaChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type gigaChatRequest struct {
	Model    string            `json:"model"`
	Messages []gigaChatMessage `json:"messages"`
}

type gigaChatChoice struct {
	Message gigaChatMessage `json:"message"`
}

type gigaChatResponse struct {
	Choices []gigaChatChoice `json:"choices"`
}

type GigaChatProvider struct {
	authorizationKey string
	scope            string
	httpClient       *http.Client
	oauthURL         string
	apiURL           string

	mu          sync.Mutex
	accessToken string
	expiresAt   time.Time
}

func NewGigaChatProvider(authorizationKey, scope string) *GigaChatProvider {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	return &GigaChatProvider{
		authorizationKey: authorizationKey,
		scope:            scope,
		httpClient:       &http.Client{Transport: transport, Timeout: 30 * time.Second},
		oauthURL:         defaultOAuthURL,
		apiURL:           defaultAPIURL,
	}
}

// SetEndpoints overrides the default GigaChat URLs — used in tests.
func (p *GigaChatProvider) SetEndpoints(oauthURL, apiURL string) {
	p.oauthURL = oauthURL
	p.apiURL = apiURL
}

func (p *GigaChatProvider) getAccessToken(ctx context.Context) (string, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.accessToken != "" && time.Now().Before(p.expiresAt) {
		return p.accessToken, nil
	}

	data := url.Values{}
	data.Set("scope", p.scope)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.oauthURL, strings.NewReader(data.Encode()))
	if err != nil {
		return "", fmt.Errorf("gigachat: build token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Basic "+p.authorizationKey)
	req.Header.Set("RqUID", newRqUID())

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("gigachat: token request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("gigachat: token response %d: %s", resp.StatusCode, body)
	}

	var tokenResp gigaChatTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", fmt.Errorf("gigachat: decode token response: %w", err)
	}

	p.accessToken = tokenResp.AccessToken
	p.expiresAt = time.Now().Add(25 * time.Minute)

	return p.accessToken, nil
}

func (p *GigaChatProvider) Complete(ctx context.Context, systemPrompt string, history []entity.ChatMessage, userMessage string) (string, error) {
	token, err := p.getAccessToken(ctx)
	if err != nil {
		return "", err
	}

	messages := []gigaChatMessage{
		{Role: "system", Content: systemPrompt},
	}
	for _, msg := range history {
		messages = append(messages, gigaChatMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}
	messages = append(messages, gigaChatMessage{Role: "user", Content: userMessage})

	reqBody := gigaChatRequest{
		Model:    gigaChatModel,
		Messages: messages,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("gigachat: marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.apiURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("gigachat: build chat request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("gigachat: chat request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("gigachat: chat response %d: %s", resp.StatusCode, body)
	}

	var chatResp gigaChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return "", fmt.Errorf("gigachat: decode chat response: %w", err)
	}

	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("gigachat: empty choices in response")
	}

	return chatResp.Choices[0].Message.Content, nil
}

func newRqUID() string {
	return uuid.New().String()
}
