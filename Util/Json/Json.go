package Json

import "github.com/bytedance/sonic"

func Unmarshal(data []byte, v any) error {
	return sonic.Unmarshal(data, v)
}

func Marshal(v any) ([]byte, error) {
	return sonic.Marshal(v)
}

func UnmarshalString(data string, v any) error {
	return sonic.UnmarshalString(data, v)
}

func MarshalString(v any) (string, error) {
	return sonic.MarshalString(v)
}
