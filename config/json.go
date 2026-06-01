package config

import (
	"encoding/json"
	"io"
	"os"
)

type jsonSource struct{ path string }

func (s *jsonSource) Load(target any) error {
	f, err := os.Open(s.path)
	if err != nil {
		return err
	}
	defer f.Close()
	return decodeJSON(f, target)
}

func decodeJSON(r io.Reader, target any) error {
	var raw any
	if err := json.NewDecoder(r).Decode(&raw); err != nil {
		return err
	}
	data, ok := raw.(map[string]any)
	if !ok {
		return nil
	}
	return setFromNestedMap(target, data)
}

func setFromNestedMap(target any, data map[string]any) error {
	fields, err := collectFields(target)
	if err != nil {
		return err
	}
	for _, f := range fields {
		val := lookupNested(data, f.path)
		if val == nil {
			continue
		}
		if err := setFieldFromAny(f.value, val); err != nil {
			return err
		}
	}
	return nil
}
