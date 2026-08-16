package ai

import "context"

// Recognition 图片识别结果
type Recognition struct {
	Name        string   `json:"name"`
	Ingredients []string `json:"ingredients"`
	Tags        []string `json:"tags"`
}

// Provider AI 提供者接口（支持 OpenAI 兼容与 Ollama 两种实现）
type Provider interface {
	// Name 提供者名称（日志用）
	Name() string
	// CompleteTags 根据提示词返回标签列表
	// keyOverride 为前端通过 X-AI-Key 传入的覆盖密钥，为空则用默认配置
	CompleteTags(ctx context.Context, prompt, keyOverride string) ([]string, error)
	// RecognizeImage 识别图片，返回菜名/食材/建议标签
	RecognizeImage(ctx context.Context, imageURL, keyOverride string) (*Recognition, error)
}
