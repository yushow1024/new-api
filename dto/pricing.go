package dto

import (
	"encoding/json"

	"github.com/QuantumNous/new-api/constant"
)

// 这里不好动就不动了，本来想独立出来的（
type OpenAIModels struct {
	Id                     string                  `json:"id"`
	Object                 string                  `json:"object"`
	Created                int                     `json:"created"`
	OwnedBy                string                  `json:"owned_by"`
	Ext                    json.RawMessage         `json:"ext,omitempty"`
	SupportedEndpointTypes []constant.EndpointType `json:"supported_endpoint_types"`
	Price                  *OpenAIModelPrice       `json:"price,omitempty"`
}

// OpenAIModelPrice contains the pricing configuration exposed by /v1/models.
// Pointer fields preserve explicit zero values while allowing an empty price object.
type OpenAIModelPrice struct {
	QuotaType            *int     `json:"quota_type,omitempty"`
	ModelRatio           *float64 `json:"model_ratio,omitempty"`
	ModelPrice           *float64 `json:"model_price,omitempty"`
	CompletionRatio      *float64 `json:"completion_ratio,omitempty"`
	CacheRatio           *float64 `json:"cache_ratio,omitempty"`
	CreateCacheRatio     *float64 `json:"create_cache_ratio,omitempty"`
	ImageRatio           *float64 `json:"image_ratio,omitempty"`
	AudioRatio           *float64 `json:"audio_ratio,omitempty"`
	AudioCompletionRatio *float64 `json:"audio_completion_ratio,omitempty"`
	BillingMode          string   `json:"billing_mode,omitempty"`
	BillingExpr          string   `json:"billing_expr,omitempty"`
	PerSecondPrice       *float64 `json:"billing_per_second_price,omitempty"`
	PerSecondRules       string   `json:"billing_per_second_rules,omitempty"`
}

type AnthropicModel struct {
	ID          string `json:"id"`
	CreatedAt   string `json:"created_at"`
	DisplayName string `json:"display_name"`
	Type        string `json:"type"`
}

type GeminiModel struct {
	Name                       interface{}   `json:"name"`
	BaseModelId                interface{}   `json:"baseModelId"`
	Version                    interface{}   `json:"version"`
	DisplayName                interface{}   `json:"displayName"`
	Description                interface{}   `json:"description"`
	InputTokenLimit            interface{}   `json:"inputTokenLimit"`
	OutputTokenLimit           interface{}   `json:"outputTokenLimit"`
	SupportedGenerationMethods []interface{} `json:"supportedGenerationMethods"`
	Thinking                   interface{}   `json:"thinking"`
	Temperature                interface{}   `json:"temperature"`
	MaxTemperature             interface{}   `json:"maxTemperature"`
	TopP                       interface{}   `json:"topP"`
	TopK                       interface{}   `json:"topK"`
}
