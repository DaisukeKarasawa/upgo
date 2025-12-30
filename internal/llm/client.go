package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.uber.org/zap"
)

type Client struct {
	baseURL string
	model   string
	timeout int
	logger  *zap.Logger
	client  *http.Client
}

type OllamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

type OllamaResponse struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
}

func NewClient(baseURL string, model string, timeout int, logger *zap.Logger) *Client {
	return &Client{
		baseURL: baseURL,
		model:   model,
		timeout: timeout,
		logger:  logger,
		client: &http.Client{
			Timeout: time.Duration(timeout) * time.Second,
		},
	}
}

func (c *Client) CheckConnection(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/api/tags", c.baseURL), nil)
	if err != nil {
		return fmt.Errorf("リクエストの作成に失敗しました: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("Ollamaへの接続に失敗しました: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Ollama APIがエラーを返しました: status %d", resp.StatusCode)
	}

	// Check if the model exists
	var tagsResponse struct {
		Models []struct {
			Name string `json:"name"`
		} `json:"models"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tagsResponse); err != nil {
		return fmt.Errorf("レスポンスの解析に失敗しました: %w", err)
	}

	modelExists := false
	for _, model := range tagsResponse.Models {
		if model.Name == c.model {
			modelExists = true
			break
		}
	}

	if !modelExists {
		return fmt.Errorf("モデル '%s' が見つかりません。'ollama pull %s' を実行してください", c.model, c.model)
	}

	return nil
}

func (c *Client) Generate(ctx context.Context, prompt string) (string, error) {
	reqBody := OllamaRequest{
		Model:  c.model,
		Prompt:  prompt,
		Stream:  false,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("リクエストボディの作成に失敗しました: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/api/generate", c.baseURL), bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("リクエストの作成に失敗しました: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("リクエストの送信に失敗しました: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("Ollama APIがエラーを返しました: status %d, body: %s", resp.StatusCode, string(body))
	}

	var ollamaResp OllamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return "", fmt.Errorf("レスポンスの解析に失敗しました: %w", err)
	}

	return ollamaResp.Response, nil
}
