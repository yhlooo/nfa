package agents

import (
	"encoding/json"

	"github.com/firebase/genkit/go/ai"
)

const (
	MetaKeyModelName         = "modelName"
	MetaKeyAvailableModels   = "availableModels"
	MetaKeyDefaultModel      = "defaultModel"
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

// GetMetaStringSliceValue 从 _meta 中获取指定 key 的字符串切片值
func GetMetaStringSliceValue(meta any, key string) []string {
	v, ok := GetMetaValue(meta, key).([]any)
	if !ok {
		return nil
	}

	ret := make([]string, 0, len(v))
	for _, item := range v {
		vStr, ok := item.(string)
		if !ok {
			continue
		}
		ret = append(ret, vStr)
	}

	return ret
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
