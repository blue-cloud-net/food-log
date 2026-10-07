package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

// ErrNoAPIKey 未配置 API Key
var ErrNoAPIKey = errors.New("未配置 AI API Key")

// OpenAICompat OpenAI 兼容协议实现（覆盖 OpenAI/DeepSeek/通义/智谱 等）
type OpenAICompat struct {
	baseURL    string // 如 https://api.openai.com/v1
	apiKey     string
	model      string
	timeout    time.Duration
	httpClient *http.Client
}

// NewOpenAICompat 创建 OpenAI 兼容 Provider
func NewOpenAICompat(baseURL, apiKey, model string, timeout time.Duration) *OpenAICompat {
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}
	if timeout <= 0 {
		timeout = 3 * time.Second
	}
	return &OpenAICompat{
		baseURL:    strings.TrimRight(baseURL, "/"),
		apiKey:     apiKey,
		model:      model,
		timeout:    timeout,
		httpClient: &http.Client{},
	}
}

// Name 提供者名称
func (p *OpenAICompat) Name() string { return "openai-compat" }

type chatMessage struct {
	Role    string `json:"role"`
	Content any    `json:"content"`
}

type chatCompletionsRequest struct {
	Model    string        `json:"model"`
	Messages []chatMessage `json:"messages"`
	Stream   bool          `json:"stream"`
}

type chatCompletionsResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func (p *OpenAICompat) endpoint() string {
	return p.baseURL + "/chat/completions"
}

func (p *OpenAICompat) doChat(ctx context.Context, messages []chatMessage, keyOverride string) (string, error) {
	apiKey := p.apiKey
	if keyOverride != "" {
		apiKey = keyOverride
	}
	if apiKey == "" {
		return "", ErrNoAPIKey
	}
	if p.model == "" {
		return "", errors.New("未配置 AI 模型")
	}

	body, err := json.Marshal(chatCompletionsRequest{Model: p.model, Messages: messages, Stream: false})
	if err != nil {
		return "", err
	}

	reqCtx := ctx
	if p.timeout > 0 {
		var cancel context.CancelFunc
		reqCtx, cancel = context.WithTimeout(ctx, p.timeout)
		defer cancel()
	}

	req, err := http.NewRequestWithContext(reqCtx, http.MethodPost, p.endpoint(), bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

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
		return "", fmt.Errorf("AI 接口错误(%d): %s", resp.StatusCode, truncate(string(data), 200))
	}

	var cr chatCompletionsResponse
	if err := json.Unmarshal(data, &cr); err != nil {
		return "", err
	}
	if len(cr.Choices) == 0 || strings.TrimSpace(cr.Choices[0].Message.Content) == "" {
		return "", errors.New("AI 返回为空")
	}
	return cr.Choices[0].Message.Content, nil
}

// CompleteTags 根据提示词返回标签列表
func (p *OpenAICompat) CompleteTags(ctx context.Context, prompt, keyOverride string) ([]string, error) {
	raw, err := p.doChat(ctx, []chatMessage{{Role: "user", Content: prompt}}, keyOverride)
	if err != nil {
		return nil, err
	}
	return parseTags(raw)
}

const visionPrompt = `识别这张美食图片，返回 JSON 对象（不要输出其它内容）：
{"name":"菜名","ingredients":["主要食材"],"tags":["1~3个标签"]}`

// RecognizeImage 识别图片，返回菜名/食材/建议标签
func (p *OpenAICompat) RecognizeImage(ctx context.Context, imageURL, keyOverride string) (*Recognition, error) {
	content := []any{
		map[string]any{"type": "text", "text": visionPrompt},
		map[string]any{"type": "image_url", "image_url": map[string]any{"url": imageURL}},
	}
	raw, err := p.doChat(ctx, []chatMessage{{Role: "user", Content: content}}, keyOverride)
	if err != nil {
		return nil, err
	}
	return parseRecognition(raw)
}

var (
	jsonArrayRe  = regexp.MustCompile(`\[[\s\S]*\]`)
	jsonObjectRe = regexp.MustCompile(`\{[\s\S]*\}`)
)

// parseTags 从模型输出中提取 JSON 字符串数组
func parseTags(content string) ([]string, error) {
	m := jsonArrayRe.FindString(content)
	if m == "" {
		return nil, errors.New("AI 响应中未找到 JSON 数组")
	}
	var tags []string
	if err := json.Unmarshal([]byte(m), &tags); err != nil {
		return nil, fmt.Errorf("解析 AI 标签失败: %w", err)
	}
	// 过滤空串
	out := tags[:0]
	for _, t := range tags {
		if t = strings.TrimSpace(t); t != "" {
			out = append(out, t)
		}
	}
	return out, nil
}

// parseRecognition 从模型输出中提取识别结果 JSON
func parseRecognition(content string) (*Recognition, error) {
	m := jsonObjectRe.FindString(content)
	if m == "" {
		return nil, errors.New("AI 响应中未找到 JSON 对象")
	}
	var rec Recognition
	if err := json.Unmarshal([]byte(m), &rec); err != nil {
		return nil, fmt.Errorf("解析 AI 识别结果失败: %w", err)
	}
	rec.Name = strings.TrimSpace(rec.Name)
	return &rec, nil
}

func truncate(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n]) + "..."
}
