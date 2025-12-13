package sonic

import (
	"encoding/json"
	"io"
)

// Config mirrors the upstream sonic configuration type but delegates to encoding/json.
type Config struct{}

var (
	ConfigStd     = Config{}
	ConfigDefault = Config{}
	ConfigFastest = Config{}
)

func (Config) Marshal(v any) ([]byte, error) {
	return json.Marshal(v)
}

func Marshal(v any) ([]byte, error) {
	return json.Marshal(v)
}

func (Config) Unmarshal(data []byte, v any) error {
	return json.Unmarshal(data, v)
}

func Unmarshal(data []byte, v any) error {
	return json.Unmarshal(data, v)
}

func (Config) MarshalIndent(v any, prefix, indent string) ([]byte, error) {
	return json.MarshalIndent(v, prefix, indent)
}

func (Config) NewDecoder(r io.Reader) *json.Decoder {
	return json.NewDecoder(r)
}

func NewDecoder(r io.Reader) *json.Decoder {
	return json.NewDecoder(r)
}

func (Config) NewEncoder(w io.Writer) *json.Encoder {
	return json.NewEncoder(w)
}

func NewEncoder(w io.Writer) *json.Encoder {
	return json.NewEncoder(w)
}

func (Config) MarshalString(v any) (string, error) {
	bz, err := json.Marshal(v)
	return string(bz), err
}

func MarshalString(v any) (string, error) {
	bz, err := json.Marshal(v)
	return string(bz), err
}

func (Config) UnmarshalString(data string, v any) error {
	return json.Unmarshal([]byte(data), v)
}

func UnmarshalString(data string, v any) error {
	return json.Unmarshal([]byte(data), v)
}
