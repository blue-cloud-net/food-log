package ai

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// MaxImageBytes 传给 AI 识别的图片大小上限（20MB）
const MaxImageBytes = 20 << 20

// Ollama 本地 Ollama 实现（/api/chat，支持 llava/qwen2.5-vl 等视觉模型）
type Ollama struct {
	baseURL    string // 如 http://localhost:11434
	model      string
	timeout    time.Duration
	httpClient *http.Client
}

// NewOllama 创建 Ollama Provider
func NewOllama(baseURL, model string, timeout time.Duration) *Ollama {
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}
	if model == "" {
		model = "llava"
	}
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	return &Ollama{
		baseURL:    strings.TrimRight(baseURL, "/"),
		model:      model,
		timeout:    timeout,
		httpClient: &http.Client{},
	}
}

// Name 提供者名称
func (p *Ollama) Name() string { return "ollama" }

type ollamaMessage struct {
	Role    string   `json:"role"`
	Content string   `json:"content"`
	Images  []string `json:"images,omitempty"` // base64 图片
}

type ollamaChatRequest struct {
	Model    string          `json:"model"`
	Messages []ollamaMessage `json:"messages"`
	Stream   bool            `json:"stream"`
}

type ollamaChatResponse struct {
	Message struct {
		Content string `json:"content"`
	} `json:"message"`
}

func (p *Ollama) doChat(ctx context.Context, messages []ollamaMessage) (string, error) {
	if p.model == "" {
		return "", errors.New("未配置 Ollama 模型")
	}
	body, err := json.Marshal(ollamaChatRequest{Model: p.model, Messages: messages, Stream: false})
	if err != nil {
		return "", err
	}

	reqCtx := ctx
	if p.timeout > 0 {
		var cancel context.CancelFunc
		reqCtx, cancel = context.WithTimeout(ctx, p.timeout)
		defer cancel()
	}

	req, err := http.NewRequestWithContext(reqCtx, http.MethodPost, p.baseURL+"/api/chat", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Ollama 接口错误(%d): %s", resp.StatusCode, truncate(string(data), 200))
	}

	var cr ollamaChatResponse
	if err := json.Unmarshal(data, &cr); err != nil {
		return "", err
	}
	if strings.TrimSpace(cr.Message.Content) == "" {
		return "", errors.New("Ollama 返回为空")
	}
	return cr.Message.Content, nil
}

// CompleteTags 根据提示词返回标签列表（keyOverride 对本地 Ollama 无意义）
func (p *Ollama) CompleteTags(ctx context.Context, prompt, _ string) ([]string, error) {
	raw, err := p.doChat(ctx, []ollamaMessage{{Role: "user", Content: prompt}})
	if err != nil {
		return nil, err
	}
	return parseTags(raw)
}

// RecognizeImage 识别图片（下载图片转 base64 后传给 Ollama 视觉模型）
func (p *Ollama) RecognizeImage(ctx context.Context, imageURL, _ string) (*Recognition, error) {
	b64, err := fetchImageBase64(ctx, imageURL, p.timeout)
	if err != nil {
		return nil, err
	}
	raw, err := p.doChat(ctx, []ollamaMessage{{Role: "user", Content: visionPrompt, Images: []string{b64}}})
	if err != nil {
		return nil, err
	}
	return parseRecognition(raw)
}

// fetchImageBase64 下载图片并转 base64
func fetchImageBase64(ctx context.Context, url string, timeout time.Duration) (string, error) {
	reqCtx := ctx
	if timeout > 0 {
		var cancel context.CancelFunc
		reqCtx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}
	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		return "", fmt.Errorf("下载图片失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("下载图片失败(%d)", resp.StatusCode)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, MaxImageBytes+1))
	if err != nil {
		return "", err
	}
	if len(data) > MaxImageBytes {
		return "", errors.New("图片超过 20MB 限制")
	}
	return base64.StdEncoding.EncodeToString(data), nil
}
