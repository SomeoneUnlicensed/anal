package anal

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

const (
	trueToken  = "T"
	falseToken = "F"
	nullToken  = "~"
)

var numberPattern = regexp.MustCompile(`^-?[0-9]+(\.[0-9]+)?([eE][+-]?[0-9]+)?$`)

func encodeObject(m map[string]any) ([]byte, error) {
	fields := make([]string, 0, len(m))
	for key := range m {
		fields = append(fields, key)
	}
	sort.Strings(fields)

	var sb strings.Builder
	sb.WriteString("@obj: {")
	for i, field := range fields {
		if i > 0 {
			sb.WriteByte(',')
		}
		sb.WriteString(field)
	}
	sb.WriteString("}\n")

	for i, field := range fields {
		if i > 0 {
			sb.WriteByte('|')
		}
		encoded, err := encodeScalar(m[field])
		if err != nil {
			return nil, err
		}
		sb.WriteString(encoded)
	}
	if len(fields) > 0 {
		sb.WriteByte('\n')
	}
	return []byte(sb.String()), nil
}

func encodeList(list []any) ([]byte, error) {
	if len(list) == 0 {
		return []byte("@list:\n"), nil
	}

	allObjects := true
	for _, item := range list {
		if _, ok := item.(map[string]any); !ok {
			allObjects = false
			break
		}
	}
	if allObjects {
		return encodeTable(list)
	}

	var sb strings.Builder
	sb.WriteString("@list:\n")
	for _, item := range list {
		encoded, err := encodeScalar(item)
		if err != nil {
			return nil, err
		}
		sb.WriteString(encoded)
		sb.WriteByte('\n')
	}
	return []byte(sb.String()), nil
}

func encodeTable(list []any) ([]byte, error) {
	fields := unionFields(list)

	var sb strings.Builder
	sb.WriteString("@table: {")
	for i, field := range fields {
		if i > 0 {
			sb.WriteByte(',')
		}
		sb.WriteString(field)
	}
	sb.WriteString("}\n")

	if len(fields) == 0 {
		for range list {
			sb.WriteByte('\n')
		}
		return []byte(sb.String()), nil
	}

	for _, item := range list {
		row := item.(map[string]any)
		for i, field := range fields {
			if i > 0 {
				sb.WriteByte('|')
			}
			encoded, err := encodeScalar(row[field])
			if err != nil {
				return nil, err
			}
			sb.WriteString(encoded)
		}
		sb.WriteByte('\n')
	}
	return []byte(sb.String()), nil
}

func unionFields(list []any) []string {
	seen := make(map[string]bool)
	var fields []string
	for _, item := range list {
		for key := range item.(map[string]any) {
			if !seen[key] {
				seen[key] = true
				fields = append(fields, key)
			}
		}
	}
	sort.Strings(fields)
	return fields
}

// encodeScalar encodes a value in a top-level slot (a table cell, an object
// field, or a list item), where only "|" and "\n" separate values.
func encodeScalar(value any) (string, error) {
	return encodeValue(value, false)
}

// encodeValue encodes value. nested is true when the result will be embedded
// inside an object ({}) or array ([]) literal, where "," ":" "{" "}" "[" "]"
// also need to be kept unambiguous.
func encodeValue(value any, nested bool) (string, error) {
	switch typed := value.(type) {
	case nil:
		return nullToken, nil
	case bool:
		if typed {
			return trueToken, nil
		}
		return falseToken, nil
	case string:
		if needsQuote(typed, nested) {
			return "`" + escape(typed) + "`", nil
		}
		return typed, nil
	case float64:
		return strconv.FormatFloat(typed, 'g', -1, 64), nil
	case map[string]any:
		return encodeNestedObject(typed)
	case []any:
		return encodeNestedArray(typed)
	default:
		return "", fmt.Errorf("anal: unsupported value type %T", value)
	}
}

func encodeNestedObject(m map[string]any) (string, error) {
	fields := make([]string, 0, len(m))
	for key := range m {
		fields = append(fields, key)
	}
	sort.Strings(fields)

	var sb strings.Builder
	sb.WriteByte('{')
	for i, field := range fields {
		if i > 0 {
			sb.WriteByte(',')
		}
		sb.WriteString(encodeKey(field))
		sb.WriteByte(':')
		encoded, err := encodeValue(m[field], true)
		if err != nil {
			return "", err
		}
		sb.WriteString(encoded)
	}
	sb.WriteByte('}')
	return sb.String(), nil
}

func encodeNestedArray(list []any) (string, error) {
	var sb strings.Builder
	sb.WriteByte('[')
	for i, item := range list {
		if i > 0 {
			sb.WriteByte(',')
		}
		encoded, err := encodeValue(item, true)
		if err != nil {
			return "", err
		}
		sb.WriteString(encoded)
	}
	sb.WriteByte(']')
	return sb.String(), nil
}

func encodeKey(k string) string {
	if k == "" || strings.HasPrefix(k, "`") || strings.ContainsAny(k, ":`\\\n\t\r") {
		return "`" + escape(k) + "`"
	}
	return k
}

func needsQuote(s string, nested bool) bool {
	if s == "" {
		return true
	}
	if numberPattern.MatchString(s) {
		return true
	}
	if s == trueToken || s == falseToken || s == nullToken {
		return true
	}
	// A bare value starting with "{" or "[" would be read back as a nested
	// object or array, so it must be quoted regardless of context.
	if strings.HasPrefix(s, "{") || strings.HasPrefix(s, "[") {
		return true
	}
	if strings.ContainsAny(s, "`\\|\n\t\r") {
		return true
	}
	if nested && strings.ContainsAny(s, ",:{}[]") {
		return true
	}
	return false
}

func escape(s string) string {
	var sb strings.Builder
	for _, r := range s {
		switch r {
		case '\\':
			sb.WriteString("\\\\")
		case '`':
			sb.WriteString("\\`")
		case '|':
			sb.WriteString("\\|")
		case '\n':
			sb.WriteString("\\n")
		case '\t':
			sb.WriteString("\\t")
		case '\r':
			sb.WriteString("\\r")
		default:
			sb.WriteRune(r)
		}
	}
	return sb.String()
}
