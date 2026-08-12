package config

// ActionTexts returns label texts from Action.Texts or Extra["texts"] (JSON or Go slice).
func ActionTexts(a Action) []string {
	if len(a.Texts) > 0 {
		return a.Texts
	}
	if a.Extra == nil {
		return nil
	}
	switch v := a.Extra["texts"].(type) {
	case []string:
		return v
	case []interface{}:
		var texts []string
		for _, t := range v {
			if s, ok := t.(string); ok && s != "" {
				texts = append(texts, s)
			}
		}
		return texts
	}
	return nil
}
