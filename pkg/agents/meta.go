package agents

import (
	"encoding/json"

	"github.com/firebase/genkit/go/ai"

	"github.com/yhlooo/nfa/pkg/models"
)

const (
	MetaKeyCurrentModels     = "currentModels"
	MetaKeyAvailableModels   = "availableModels"
	MetaKeyCurrentModelUsage = "currentModelUsage"
)

// GetMetaValue 从 _meta 中获取指定 key 的值
func GetMetaValue(meta any, key string) any {
	mapMeta, ok := meta.(map[string]any)
	if !ok {
		return nil
	}
	v, ok := mapMeta[key]
	if !ok {
		return nil
	}
	return v
}

// GetMetaStringValue 从 _meta 中获取指定 key 的字符串值
func GetMetaStringValue(meta any, key string) string {
	v, ok := GetMetaValue(meta, key).(string)
	if !ok {
		return ""
	}
	return v
}

// SetMetaCurrentModelUsage 往 _meta 设置当前模型用量
func SetMetaCurrentModelUsage(meta any, usage ai.GenerationUsage) {
	mapMeta, ok := meta.(map[string]any)
	if !ok {
		return
	}
	if mapMeta == nil {
		return
	}
	raw, _ := json.Marshal(usage)
	mapMeta[MetaKeyCurrentModelUsage] = string(raw)
}

// GetMetaCurrentModelUsageValue 从 _meta 中获取当前模型用量
func GetMetaCurrentModelUsageValue(meta any) ai.GenerationUsage {
	v := GetMetaStringValue(meta, MetaKeyCurrentModelUsage)
	if v == "" {
		return ai.GenerationUsage{}
	}

	usage := ai.GenerationUsage{}
	if err := json.Unmarshal([]byte(v), &usage); err != nil {
		return ai.GenerationUsage{}
	}

	return usage
}

// GetMetaAvailableModelsValue 从 _meta 中获取可用模型列表
func GetMetaAvailableModelsValue(meta any) []models.ModelConfig {
	v, ok := GetMetaValue(meta, MetaKeyAvailableModels).([]any)
	if !ok {
		return nil
	}

	ret := make([]models.ModelConfig, 0, len(v))
	for _, item := range v {
		itemMap, ok := item.(map[string]any)
		if !ok {
			continue
		}
		raw, _ := json.Marshal(itemMap)
		var cfg models.ModelConfig
		if err := json.Unmarshal(raw, &cfg); err != nil {
			continue
		}
		ret = append(ret, cfg)
	}

	return ret
}

// GetMetaCurrentModelsValue 从 _meta 中获取当前使用的模型
func GetMetaCurrentModelsValue(meta any) models.Models {
	v, ok := GetMetaValue(meta, MetaKeyCurrentModels).(map[string]any)
	if !ok {
		return models.Models{}
	}

	ret := models.Models{}

	if name, ok := v["primary"].(string); ok {
		ret.Primary = name
	}
	if name, ok := v["light"].(string); ok {
		ret.Light = name
	}
	if name, ok := v["vision"].(string); ok {
		ret.Vision = name
	}

	return ret
}
