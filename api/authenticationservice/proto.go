package authenticationservice

import (
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"net/url"
)

const (
	protoWireVarint  = 0
	protoWireFixed64 = 1
	protoWireBytes   = 2
)

type protoField struct {
	Number uint64
	Wire   uint64
	Value  []byte
}

func buildProtoForm(encoded string) []byte {
	values := url.Values{}
	values.Set("input_protobuf_encoded", encoded)
	return []byte(values.Encode())
}

func encodeProtoBase64(message []byte) string {
	return base64.StdEncoding.EncodeToString(message)
}

func appendProtoString(dst []byte, number uint64, value string) []byte {
	if value == "" {
		return dst
	}
	dst = appendProtoKey(dst, number, protoWireBytes)
	dst = appendProtoVarint(dst, uint64(len(value)))
	return append(dst, value...)
}

func appendProtoBytes(dst []byte, number uint64, value []byte) []byte {
	if len(value) == 0 {
		return dst
	}
	dst = appendProtoKey(dst, number, protoWireBytes)
	dst = appendProtoVarint(dst, uint64(len(value)))
	return append(dst, value...)
}

func appendProtoMessage(dst []byte, number uint64, message []byte) []byte {
	return appendProtoBytes(dst, number, message)
}

func appendProtoBool(dst []byte, number uint64, value bool) []byte {
	if value {
		return appendProtoUint64(dst, number, 1)
	}
	return appendProtoUint64(dst, number, 0)
}

func appendProtoUint64(dst []byte, number, value uint64) []byte {
	dst = appendProtoKey(dst, number, protoWireVarint)
	return appendProtoVarint(dst, value)
}

func appendProtoFixed64(dst []byte, number, value uint64) []byte {
	dst = appendProtoKey(dst, number, protoWireFixed64)
	var buf [8]byte
	binary.LittleEndian.PutUint64(buf[:], value)
	return append(dst, buf[:]...)
}

func appendProtoKey(dst []byte, number, wire uint64) []byte {
	return appendProtoVarint(dst, number<<3|wire)
}

func appendProtoVarint(dst []byte, value uint64) []byte {
	for value >= 0x80 {
		dst = append(dst, byte(value)|0x80)
		value >>= 7
	}
	return append(dst, byte(value))
}

func readProtoFields(message []byte) ([]protoField, error) {
	fields := make([]protoField, 0)
	for offset := 0; offset < len(message); {
		key, next, err := readProtoVarint(message, offset)
		if err != nil {
			return nil, err
		}
		offset = next
		field := protoField{
			Number: key >> 3,
			Wire:   key & 0x7,
		}

		switch field.Wire {
		case protoWireVarint:
			value, next, err := readProtoVarint(message, offset)
			if err != nil {
				return nil, err
			}
			field.Value = appendProtoVarint(nil, value)
			offset = next
		case protoWireFixed64:
			if offset+8 > len(message) {
				return nil, fmt.Errorf("truncated fixed64 field %d", field.Number)
			}
			field.Value = append([]byte(nil), message[offset:offset+8]...)
			offset += 8
		case protoWireBytes:
			size, next, err := readProtoVarint(message, offset)
			if err != nil {
				return nil, err
			}
			offset = next
			if size > uint64(len(message)-offset) {
				return nil, fmt.Errorf("truncated bytes field %d", field.Number)
			}
			field.Value = append([]byte(nil), message[offset:offset+int(size)]...)
			offset += int(size)
		default:
			return nil, fmt.Errorf("unsupported wire type %d for field %d", field.Wire, field.Number)
		}

		fields = append(fields, field)
	}
	return fields, nil
}

func readProtoVarint(data []byte, offset int) (uint64, int, error) {
	var value uint64
	for shift := uint(0); shift < 64; shift += 7 {
		if offset >= len(data) {
			return 0, offset, fmt.Errorf("truncated varint")
		}
		b := data[offset]
		offset++
		value |= uint64(b&0x7f) << shift
		if b < 0x80 {
			return value, offset, nil
		}
	}
	return 0, offset, fmt.Errorf("varint overflows uint64")
}
