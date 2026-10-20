package enum

import (
	"encoding/json"
	"errors"
	"sort"
)

var (
	ErrInvalidValue  = errors.New("无效的枚举值")
	ErrNotStringJSON = errors.New("枚举必须为字符串")
)

func enumValid[T comparable](names map[T]string, v T) bool {
	_, ok := names[v]
	return ok
}

func enumString[T comparable](names map[T]string, v T) string {
	if name, ok := names[v]; ok {
		return name
	}
	return "unknown"
}

func enumMarshalJSON[T comparable](names map[T]string, v T) ([]byte, error) {
	name, ok := names[v]
	if !ok {
		return nil, ErrInvalidValue
	}
	return json.Marshal(name)
}

func enumUnmarshalJSON[T comparable](values map[string]T, data []byte, dst *T) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return ErrNotStringJSON
	}
	v, ok := values[s]
	if !ok {
		return ErrInvalidValue
	}
	*dst = v
	return nil
}

func enumMarshalText[T comparable](names map[T]string, v T) ([]byte, error) {
	name, ok := names[v]
	if !ok {
		return nil, ErrInvalidValue
	}
	return []byte(name), nil
}

func enumUnmarshalText[T comparable](values map[string]T, data []byte, dst *T) error {
	v, ok := values[string(data)]
	if !ok {
		return ErrInvalidValue
	}
	*dst = v
	return nil
}

type Option struct {
	Value uint8  `json:"value"`
	Label string `json:"label"`
}

func enumOptions[T ~uint8](names map[T]string) []Option {
	opts := make([]Option, 0, len(names))
	for v, name := range names {
		opts = append(opts, Option{Value: uint8(v), Label: name})
	}
	sort.Slice(opts, func(i, j int) bool { return opts[i].Value < opts[j].Value })
	return opts
}
