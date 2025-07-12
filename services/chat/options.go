package chat

import (
	"fmt"
	"time"

	"github.com/hewenyu/newapi-go/types"
)

// ChatOption 聊天选项函数类型
type ChatOption func(*ChatConfig)

// ChatConfig 聊天配置结构
type ChatConfig struct {
	Model            string                    `json:"model"`
	MaxTokens        int                       `json:"max_tokens"`
	Temperature      float64                   `json:"temperature"`
	TopP             float64                   `json:"top_p"`
	N                int                       `json:"n"`
	Stream           bool                      `json:"stream"`
	Stop             interface{}               `json:"stop"`
	PresencePenalty  float64                   `json:"presence_penalty"`
	FrequencyPenalty float64                   `json:"frequency_penalty"`
	LogitBias        map[string]float64        `json:"logit_bias"`
	User             string                    `json:"user"`
	Functions        []types.ChatFunction      `json:"functions"`
	FunctionCall     interface{}               `json:"function_call"`
	Tools            []types.Tool              `json:"tools"`
	ToolChoice       interface{}               `json:"tool_choice"`
	ResponseFormat   *types.ChatResponseFormat `json:"response_format"`
	Seed             int                       `json:"seed"`
	LogProbs         bool                      `json:"logprobs"`
	TopLogProbs      int                       `json:"top_logprobs"`
	Timeout          time.Duration             `json:"timeout"`
	ExtraBody        map[string]interface{}    `json:"extra_body"`
}

// DefaultChatConfig 返回默认的聊天配置
func DefaultChatConfig() *ChatConfig {
	return &ChatConfig{
		Model:            "gpt-3.5-turbo",
		MaxTokens:        1000,
		Temperature:      1.0,
		TopP:             1.0,
		N:                1,
		Stream:           false,
		PresencePenalty:  0.0,
		FrequencyPenalty: 0.0,
		LogitBias:        make(map[string]float64),
		LogProbs:         false,
		TopLogProbs:      0,
		Timeout:          30 * time.Second,
		ExtraBody:        make(map[string]interface{}),
	}
}

// WithModel 设置聊天模型
func WithModel(model string) ChatOption {
	return func(config *ChatConfig) {
		config.Model = model
	}
}

// WithMaxTokens 设置最大Token数量
func WithMaxTokens(maxTokens int) ChatOption {
	return func(config *ChatConfig) {
		config.MaxTokens = maxTokens
	}
}

// WithTemperature 设置温度参数
func WithTemperature(temperature float64) ChatOption {
	return func(config *ChatConfig) {
		config.Temperature = temperature
	}
}

// WithTopP 设置TopP参数
func WithTopP(topP float64) ChatOption {
	return func(config *ChatConfig) {
		config.TopP = topP
	}
}

// WithN 设置生成选择数量
func WithN(n int) ChatOption {
	return func(config *ChatConfig) {
		config.N = n
	}
}

// WithStream 设置是否使用流式响应
func WithStream(stream bool) ChatOption {
	return func(config *ChatConfig) {
		config.Stream = stream
	}
}

// WithStop 设置停止序列
func WithStop(stop interface{}) ChatOption {
	return func(config *ChatConfig) {
		config.Stop = stop
	}
}

// WithPresencePenalty 设置存在惩罚
func WithPresencePenalty(penalty float64) ChatOption {
	return func(config *ChatConfig) {
		config.PresencePenalty = penalty
	}
}

// WithFrequencyPenalty 设置频率惩罚
func WithFrequencyPenalty(penalty float64) ChatOption {
	return func(config *ChatConfig) {
		config.FrequencyPenalty = penalty
	}
}

// WithLogitBias 设置Logit偏差
func WithLogitBias(bias map[string]float64) ChatOption {
	return func(config *ChatConfig) {
		config.LogitBias = bias
	}
}

// WithUser 设置用户ID
func WithUser(user string) ChatOption {
	return func(config *ChatConfig) {
		config.User = user
	}
}

// WithFunctions 设置函数列表
func WithFunctions(functions []types.ChatFunction) ChatOption {
	return func(config *ChatConfig) {
		config.Functions = functions
	}
}

// WithFunctionCall 设置函数调用
func WithFunctionCall(functionCall interface{}) ChatOption {
	return func(config *ChatConfig) {
		config.FunctionCall = functionCall
	}
}

// WithTools 设置工具列表
func WithTools(tools []types.Tool) ChatOption {
	return func(config *ChatConfig) {
		config.Tools = tools
	}
}

// WithToolChoice 设置工具选择
func WithToolChoice(toolChoice interface{}) ChatOption {
	return func(config *ChatConfig) {
		config.ToolChoice = toolChoice
	}
}

// WithResponseFormat 设置响应格式
func WithResponseFormat(format *types.ChatResponseFormat) ChatOption {
	return func(config *ChatConfig) {
		config.ResponseFormat = format
	}
}

// WithSeed 设置随机种子
func WithSeed(seed int) ChatOption {
	return func(config *ChatConfig) {
		config.Seed = seed
	}
}

// WithLogProbs 设置是否返回日志概率
func WithLogProbs(logProbs bool) ChatOption {
	return func(config *ChatConfig) {
		config.LogProbs = logProbs
	}
}

// WithTopLogProbs 设置顶级日志概率数量
func WithTopLogProbs(topLogProbs int) ChatOption {
	return func(config *ChatConfig) {
		config.TopLogProbs = topLogProbs
	}
}

// WithTimeout 设置超时时间
func WithTimeout(timeout time.Duration) ChatOption {
	return func(config *ChatConfig) {
		config.Timeout = timeout
	}
}

// WithExtraBody 设置额外的请求体参数
func WithExtraBody(extraBody map[string]interface{}) ChatOption {
	return func(config *ChatConfig) {
		config.ExtraBody = extraBody
	}
}

// ToRequest 将配置转换为请求结构
func (c *ChatConfig) ToRequest(messages []types.ChatMessage) *types.ChatCompletionRequest {
	req := &types.ChatCompletionRequest{
		Model:            c.Model,
		Messages:         messages,
		MaxTokens:        c.MaxTokens,
		Temperature:      c.Temperature,
		TopP:             c.TopP,
		N:                c.N,
		Stream:           c.Stream,
		Stop:             c.Stop,
		PresencePenalty:  c.PresencePenalty,
		FrequencyPenalty: c.FrequencyPenalty,
		LogitBias:        c.LogitBias,
		User:             c.User,
		Functions:        c.Functions,
		FunctionCall:     c.FunctionCall,
		Tools:            c.Tools,
		ToolChoice:       c.ToolChoice,
		ResponseFormat:   c.ResponseFormat,
		Seed:             c.Seed,
		LogProbs:         c.LogProbs,
		TopLogProbs:      c.TopLogProbs,
		ExtraBody:        c.ExtraBody,
	}

	// 设置默认值
	req.SetDefaults()

	return req
}

// Clone 克隆配置
func (c *ChatConfig) Clone() *ChatConfig {
	clone := *c

	// 深拷贝map
	if c.LogitBias != nil {
		clone.LogitBias = make(map[string]float64)
		for k, v := range c.LogitBias {
			clone.LogitBias[k] = v
		}
	}

	if c.ExtraBody != nil {
		clone.ExtraBody = make(map[string]interface{})
		for k, v := range c.ExtraBody {
			clone.ExtraBody[k] = v
		}
	}

	// 深拷贝切片
	if c.Functions != nil {
		clone.Functions = make([]types.ChatFunction, len(c.Functions))
		copy(clone.Functions, c.Functions)
	}

	if c.Tools != nil {
		clone.Tools = make([]types.Tool, len(c.Tools))
		copy(clone.Tools, c.Tools)
	}

	return &clone
}

// Validate 验证配置
func (c *ChatConfig) Validate() error {
	if c.Model == "" {
		return fmt.Errorf("model cannot be empty")
	}

	if c.MaxTokens < 0 {
		return fmt.Errorf("max_tokens must be non-negative")
	}

	if c.Temperature < 0.0 || c.Temperature > 2.0 {
		return fmt.Errorf("temperature must be between 0.0 and 2.0")
	}

	if c.TopP < 0.0 || c.TopP > 1.0 {
		return fmt.Errorf("top_p must be between 0.0 and 1.0")
	}

	if c.N < 1 {
		return fmt.Errorf("n must be at least 1")
	}

	if c.PresencePenalty < -2.0 || c.PresencePenalty > 2.0 {
		return fmt.Errorf("presence_penalty must be between -2.0 and 2.0")
	}

	if c.FrequencyPenalty < -2.0 || c.FrequencyPenalty > 2.0 {
		return fmt.Errorf("frequency_penalty must be between -2.0 and 2.0")
	}

	if c.TopLogProbs < 0 || c.TopLogProbs > 5 {
		return fmt.Errorf("top_logprobs must be between 0 and 5")
	}

	if c.Timeout <= 0 {
		return fmt.Errorf("timeout must be positive")
	}

	return nil
}
