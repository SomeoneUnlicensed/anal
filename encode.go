package anal

import (
	"encoding/json"
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
	flat := flattenMap(m, "", nil)
	fields := make([]string, 0, len(flat))
	for key := range flat {
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
		encoded, err := encodeScalar(flat[field])
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
	rows := make([]map[string]any, len(list))
	seen := make(map[string]bool)
	var fields []string
	for i, item := range list {
		rows[i] = flattenMap(item.(map[string]any), "", nil)
		for key := range rows[i] {
			if !seen[key] {
				seen[key] = true
				fields = append(fields, key)
			}
		}
	}
	sort.Strings(fields)

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

	for _, row := range rows {
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

// flattenMap collapses nested objects into dot-joined header paths, e.g.
// {"user": {"name": "x"}} becomes {"user.name": "x"}. Empty maps, arrays and
// scalars stay as leaf values; path segments are escaped for the header.
func flattenMap(m map[string]any, prefix string, out map[string]any) map[string]any {
	if out == nil {
		out = make(map[string]any, len(m))
	}
	for key, value := range m {
		path := escapeField(key)
		if prefix != "" {
			path = prefix + "." + path
		}
		if child, ok := value.(map[string]any); ok && len(child) > 0 {
			flattenMap(child, path, out)
			continue
		}
		out[path] = value
	}
	return out
}

// escapeField escapes header metacharacters in one field path segment.
func escapeField(s string) string {
	var sb strings.Builder
	for _, r := range s {
		switch r {
		case '\\', '.', ',', '{', '}':
			sb.WriteByte('\\')
		}
		sb.WriteRune(r)
	}
	return sb.String()
}

func encodeScalar(value any) (string, error) {
	switch typed := value.(type) {
	case nil:
		return nullToken, nil
	case bool:
		if typed {
			return trueToken, nil
		}
		return falseToken, nil
	case string:
		if needsQuote(typed) {
			return "`" + escape(typed) + "`", nil
		}
		return typed, nil
	case float64:
		return strconv.FormatFloat(typed, 'g', -1, 64), nil
	case map[string]any, []any:
		raw, err := json.Marshal(typed)
		if err != nil {
			return "", err
		}
		return "&" + escape(string(raw)), nil
	default:
		return "", fmt.Errorf("anal: unsupported value type %T", value)
	}
}

func needsQuote(s string) bool {
	if s == "" {
		return true
	}
	if numberPattern.MatchString(s) {
		return true
	}
	if s == trueToken || s == falseToken || s == nullToken {
		return true
	}
	if strings.HasPrefix(s, "&") {
		return true
	}
	return strings.ContainsAny(s, "`\\|\n\t\r")
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
