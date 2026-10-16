package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

// EmbeddingProvider defines how to generate text embeddings
type EmbeddingProvider interface {
	Embed(ctx context.Context, texts []string) ([][]float32, error)
}

// CompletionProvider defines how to generate LLM chat completions
type CompletionProvider interface {
	Complete(ctx context.Context, systemPrompt, userMessage string) (string, int, error)
}

// ---------------------------------------------------------------------------
// OpenAI Provider
// ---------------------------------------------------------------------------

type OpenAIProvider struct {
	apiKey         string
	model          string
	embeddingModel string
	httpClient     *http.Client
}

func NewOpenAIProvider() *OpenAIProvider {
	return &OpenAIProvider{
		apiKey:         getEnvOrPlaceholder("OPENAI_API_KEY", "sk-placeholder-set-your-key-in-env"),
		model:          getEnvOrPlaceholder("OPENAI_MODEL", "gpt-4o-mini"),
		embeddingModel: getEnvOrPlaceholder("OPENAI_EMBEDDING_MODEL", "text-embedding-3-small"),
		httpClient:     &http.Client{},
	}
}

func (p *OpenAIProvider) Embed(ctx context.Context, texts []string) ([][]float32, error) {
	if isPlaceholder(p.apiKey) {
		return generateDummyEmbeddings(texts, 384), nil
	}

	reqBody, _ := json.Marshal(map[string]interface{}{
		"model": p.embeddingModel,
		"input": texts,
	})

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/embeddings", bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+p.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("openai embedding request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("openai embedding API error %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Data []struct {
			Embedding []float32 `json:"embedding"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse openai embedding response: %w", err)
	}

	embeddings := make([][]float32, len(result.Data))
	for i, d := range result.Data {
		embeddings[i] = d.Embedding
	}
	return embeddings, nil
}

func (p *OpenAIProvider) Complete(ctx context.Context, systemPrompt, userMessage string) (string, int, error) {
	if isPlaceholder(p.apiKey) {
		return generateDummyAnswer(userMessage), 0, nil
	}

	reqBody, _ := json.Marshal(map[string]interface{}{
		"model": p.model,
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": userMessage},
		},
		"max_tokens":  1024,
		"temperature": 0.3,
	})

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/chat/completions", bytes.NewReader(reqBody))
	if err != nil {
		return "", 0, err
	}
	req.Header.Set("Authorization", "Bearer "+p.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return "", 0, fmt.Errorf("openai completion request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", 0, fmt.Errorf("openai API error %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Usage struct {
			TotalTokens int `json:"total_tokens"`
		} `json:"usage"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", 0, fmt.Errorf("failed to parse openai completion response: %w", err)
	}

	if len(result.Choices) == 0 {
		return "", 0, fmt.Errorf("no completion choices returned")
	}

	return result.Choices[0].Message.Content, result.Usage.TotalTokens, nil
}

// ---------------------------------------------------------------------------
// Gemini Provider
// ---------------------------------------------------------------------------

type GeminiProvider struct {
	apiKey         string
	model          string
	embeddingModel string
	httpClient     *http.Client
}

func NewGeminiProvider() *GeminiProvider {
	return &GeminiProvider{
		apiKey:         getEnvOrPlaceholder("GEMINI_API_KEY", "AIza-placeholder-set-your-key-in-env"),
		model:          getEnvOrPlaceholder("GEMINI_MODEL", "gemini-3.5-flash-lite"),
		embeddingModel: getEnvOrPlaceholder("GEMINI_EMBEDDING_MODEL", "gemini-embedding-001"),
		httpClient:     &http.Client{},
	}
}

func (p *GeminiProvider) Embed(ctx context.Context, texts []string) ([][]float32, error) {
	if isPlaceholder(p.apiKey) {
		return generateDummyEmbeddings(texts, 768), nil
	}

	var embeddings [][]float32
	for _, text := range texts {
		url := fmt.Sprintf(
			"https://generativelanguage.googleapis.com/v1beta/models/%s:embedContent?key=%s",
			p.embeddingModel, p.apiKey,
		)
		reqBody, _ := json.Marshal(map[string]interface{}{
			"model": "models/" + p.embeddingModel,
			"content": map[string]interface{}{
				"parts": []map[string]string{{"text": text}},
			},
		})

		req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(reqBody))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := p.httpClient.Do(req)
		if err != nil {
			return nil, err
		}
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("gemini embedding API error %d: %s", resp.StatusCode, string(body))
		}

		var result struct {
			Embedding struct {
				Values []float32 `json:"values"`
			} `json:"embedding"`
		}
		if err := json.Unmarshal(body, &result); err != nil {
			return nil, err
		}
		embeddings = append(embeddings, result.Embedding.Values)
	}
	return embeddings, nil
}

func (p *GeminiProvider) Complete(ctx context.Context, systemPrompt, userMessage string) (string, int, error) {
	if isPlaceholder(p.apiKey) {
		return generateDummyAnswer(userMessage), 0, nil
	}

	url := fmt.Sprintf(
		"https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s",
		p.model, p.apiKey,
	)

	combined := systemPrompt + "\n\nUser Question:\n" + userMessage
	reqBody, _ := json.Marshal(map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"role":  "user",
				"parts": []map[string]string{{"text": combined}},
			},
		},
		"generationConfig": map[string]interface{}{
			"maxOutputTokens": 1024,
			"temperature":     0.3,
		},
	})

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(reqBody))
	if err != nil {
		return "", 0, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return "", 0, fmt.Errorf("gemini request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", 0, fmt.Errorf("gemini API error %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", 0, err
	}
	if len(result.Candidates) == 0 || len(result.Candidates[0].Content.Parts) == 0 {
		return "", 0, fmt.Errorf("no content returned from gemini")
	}
	return result.Candidates[0].Content.Parts[0].Text, 0, nil
}

// ---------------------------------------------------------------------------
// Mistral Provider
// ---------------------------------------------------------------------------

type MistralProvider struct {
	apiKey         string
	model          string
	embeddingModel string
	httpClient     *http.Client
}

func NewMistralProvider() *MistralProvider {
	return &MistralProvider{
		apiKey:         getEnvOrPlaceholder("MISTRAL_API_KEY", "placeholder-set-your-key-in-env"),
		model:          getEnvOrPlaceholder("MISTRAL_MODEL", "mistral-large-latest"),
		embeddingModel: getEnvOrPlaceholder("MISTRAL_EMBEDDING_MODEL", "mistral-embed"),
		httpClient:     &http.Client{},
	}
}

func (p *MistralProvider) Embed(ctx context.Context, texts []string) ([][]float32, error) {
	if isPlaceholder(p.apiKey) {
		return generateDummyEmbeddings(texts, 1024), nil
	}

	reqBody, _ := json.Marshal(map[string]interface{}{
		"model": p.embeddingModel,
		"input": texts,
	})

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.mistral.ai/v1/embeddings", bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+p.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("mistral embedding request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("mistral embedding API error %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Data []struct {
			Embedding []float32 `json:"embedding"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse mistral embedding response: %w", err)
	}

	embeddings := make([][]float32, len(result.Data))
	for i, d := range result.Data {
		embeddings[i] = d.Embedding
	}
	return embeddings, nil
}

func (p *MistralProvider) Complete(ctx context.Context, systemPrompt, userMessage string) (string, int, error) {
	if isPlaceholder(p.apiKey) {
		return generateDummyAnswer(userMessage), 0, nil
	}

	reqBody, _ := json.Marshal(map[string]interface{}{
		"model": p.model,
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": userMessage},
		},
		"max_tokens":  1024,
		"temperature": 0.3,
	})

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.mistral.ai/v1/chat/completions", bytes.NewReader(reqBody))
	if err != nil {
		return "", 0, err
	}
	req.Header.Set("Authorization", "Bearer "+p.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return "", 0, fmt.Errorf("mistral completion request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", 0, fmt.Errorf("mistral API error %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Usage struct {
			TotalTokens int `json:"total_tokens"`
		} `json:"usage"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", 0, fmt.Errorf("failed to parse mistral completion response: %w", err)
	}

	if len(result.Choices) == 0 {
		return "", 0, fmt.Errorf("no completion choices returned from mistral")
	}

	return result.Choices[0].Message.Content, result.Usage.TotalTokens, nil
}

// ---------------------------------------------------------------------------
// Provider factory
// ---------------------------------------------------------------------------

// NewEmbeddingProvider returns an embedding provider based on AI_PROVIDER env var
func NewEmbeddingProvider() EmbeddingProvider {
	switch strings.ToLower(os.Getenv("AI_PROVIDER")) {
	case "gemini":
		return NewGeminiProvider()
	case "mistral":
		return NewMistralProvider()
	default:
		return NewOpenAIProvider()
	}
}

// NewCompletionProvider returns a completion provider based on AI_PROVIDER env var
func NewCompletionProvider() CompletionProvider {
	switch strings.ToLower(os.Getenv("AI_PROVIDER")) {
	case "gemini":
		return NewGeminiProvider()
	case "mistral":
		return NewMistralProvider()
	default:
		return NewOpenAIProvider()
	}
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func getEnvOrPlaceholder(key, placeholder string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return placeholder
}

func isPlaceholder(key string) bool {
	return strings.Contains(key, "placeholder") || key == ""
}

// generateDummyEmbeddings returns zero-value embeddings for use before API keys are configured
func generateDummyEmbeddings(texts []string, dims int) [][]float32 {
	out := make([][]float32, len(texts))
	for i := range texts {
		out[i] = make([]float32, dims)
	}
	return out
}

// generateDummyAnswer produces a placeholder AI response when no API key is set
func generateDummyAnswer(question string) string {
	return fmt.Sprintf(
		"⚠️ AI is not yet configured. Please set AI_PROVIDER and the corresponding API key in your .env file.\n\nYour question was: \"%s\"",
		question,
	)
}
