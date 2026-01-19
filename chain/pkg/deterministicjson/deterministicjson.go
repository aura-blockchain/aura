// Package deterministicjson provides canonical JSON marshaling with sorted map keys
// while preserving numeric precision via json.Number. It avoids the random map
// iteration order of encoding/json that can lead to consensus nondeterminism.
package deterministicjson

import (
	"bytes"
	"encoding/json"
	"sort"
)

// Marshal encodes v into JSON with deterministic key ordering.
// It first marshals using encoding/json to honor existing struct tags,
// then re-encodes with sorted map keys while preserving numbers via json.Number.
func Marshal(v interface{}) ([]byte, error) {
	raw, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()

	var val interface{}
	if err := dec.Decode(&val); err != nil {
		return nil, err
	}

	buf := &bytes.Buffer{}
	if err := encode(buf, val); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// Unmarshal decodes deterministic JSON produced by Marshal into v.
// It simply delegates to encoding/json since ordering is already canonical.
func Unmarshal(data []byte, v interface{}) error {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	return dec.Decode(v)
}

// encode writes a deterministic JSON representation of v into buf.
func encode(buf *bytes.Buffer, v interface{}) error {
	switch t := v.(type) {
	case nil:
		buf.WriteString("null")
		return nil
	case bool:
		if t {
			buf.WriteString("true")
		} else {
			buf.WriteString("false")
		}
		return nil
	case string:
		b, err := json.Marshal(t)
		if err != nil {
			return err
		}
		buf.Write(b)
		return nil
	case json.Number:
		buf.WriteString(t.String())
		return nil
	case []interface{}:
		buf.WriteByte('[')
		for i, elem := range t {
			if i > 0 {
				buf.WriteByte(',')
			}
			if err := encode(buf, elem); err != nil {
				return err
			}
		}
		buf.WriteByte(']')
		return nil
	case map[string]interface{}:
		keys := make([]string, 0, len(t))
		for k := range t {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		buf.WriteByte('{')
		for i, k := range keys {
			if i > 0 {
				buf.WriteByte(',')
			}
			keyBytes, err := json.Marshal(k)
			if err != nil {
				return err
			}
			buf.Write(keyBytes)
			buf.WriteByte(':')
			if err := encode(buf, t[k]); err != nil {
				return err
			}
		}
		buf.WriteByte('}')
		return nil
	default:
		b, err := json.Marshal(t)
		if err != nil {
			return err
		}
		buf.Write(b)
		return nil
	}
}
