package config

func lookupNested(data map[string]any, path []string) any {
	current := data
	for i, segment := range path {
		if i == len(path)-1 {
			return current[segment]
		}
		next, ok := current[segment]
		if !ok {
			return nil
		}
		switch m := next.(type) {
		case map[string]any:
			current = m
		default:
			return nil
		}
	}
	return nil
}

func normalizeMap(val any) map[string]any {
	switch m := val.(type) {
	case map[string]any:
		for k, v := range m {
			m[k] = normalizeValue(v)
		}
		return m
	case map[any]any:
		result := make(map[string]any, len(m))
		for k, v := range m {
			result[toStringKey(k)] = normalizeValue(v)
		}
		return result
	}
	return nil
}

func normalizeValue(val any) any {
	switch v := val.(type) {
	case map[string]any:
		return normalizeMap(v)
	case map[any]any:
		return normalizeMap(v)
	case []any:
		for i, e := range v {
			v[i] = normalizeValue(e)
		}
		return v
	default:
		return val
	}
}

func toStringKey(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}
