// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types

// StringList represents a list of strings for KV store persistence
type StringList struct {
	Values []string `protobuf:"bytes,1,rep,name=values,proto3" json:"values,omitempty"`
}

// Marshal implements proto.Message for StringList
func (m *StringList) Marshal() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	// Simple length-prefixed encoding
	var totalLen int
	for _, v := range m.Values {
		totalLen += len(v) + 4 // 4 bytes for length prefix
	}

	result := make([]byte, 0, totalLen+4)
	// Write count
	count := uint32(len(m.Values))
	result = append(result, byte(count>>24), byte(count>>16), byte(count>>8), byte(count))

	// Write each string
	for _, v := range m.Values {
		vlen := uint32(len(v))
		result = append(result, byte(vlen>>24), byte(vlen>>16), byte(vlen>>8), byte(vlen))
		result = append(result, []byte(v)...)
	}

	return result, nil
}

// Unmarshal implements proto.Message for StringList
func (m *StringList) Unmarshal(data []byte) error {
	if len(data) < 4 {
		m.Values = []string{}
		return nil
	}

	// Read count
	count := uint32(data[0])<<24 | uint32(data[1])<<16 | uint32(data[2])<<8 | uint32(data[3])
	offset := 4

	m.Values = make([]string, count)
	for i := uint32(0); i < count; i++ {
		if offset+4 > len(data) {
			return ErrInvalidAmount
		}

		// Read string length
		vlen := uint32(data[offset])<<24 | uint32(data[offset+1])<<16 | uint32(data[offset+2])<<8 | uint32(data[offset+3])
		offset += 4

		if offset+int(vlen) > len(data) {
			return ErrInvalidAmount
		}

		// Read string
		m.Values[i] = string(data[offset : offset+int(vlen)])
		offset += int(vlen)
	}

	return nil
}

// Reset implements proto.Message for StringList
func (m *StringList) Reset() {
	*m = StringList{}
}

// String implements proto.Message for StringList
func (m *StringList) String() string {
	if m == nil {
		return "StringList{}"
	}
	return "StringList{Values: " + string(rune(len(m.Values))) + "}"
}

// ProtoMessage implements proto.Message for StringList
func (m *StringList) ProtoMessage() {}
