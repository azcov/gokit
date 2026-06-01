package config

import (
	"io"
	"os"

	"gopkg.in/yaml.v3"
)

type yamlSource struct{ path string }

func (s *yamlSource) Load(target any) error {
	f, err := os.Open(s.path)
	if err != nil {
		return err
	}
	defer f.Close()
	return decodeYAML(f, target)
}

func decodeYAML(r io.Reader, target any) error {
	var raw any
	if err := yaml.NewDecoder(r).Decode(&raw); err != nil {
		return err
	}
	data := normalizeMap(raw)
	if data == nil {
		return nil
	}
	return setFromNestedMap(target, data)
}
