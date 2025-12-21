package agents

const (
	MetaKeyModelName       = "modelName"
	MetaKeyAvailableModels = "availableModels"
	MetaKeyDefaultModel    = "defaultModel"
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
