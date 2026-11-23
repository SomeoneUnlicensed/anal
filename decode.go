package anal

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"
)

type block struct {
	kind   string
	fields []string
	rows   []string
}

func parse(text string) (any, error) {
	lines := strings.Split(text, "\n")
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	var current *block
	for _, line := range lines {
		line = strings.TrimSuffix(line, "\r")
		if strings.TrimSpace(line) == "" {
			if current == nil {
				continue
			}
			current.rows = append(current.rows, "")
			continue
		}
		if strings.HasPrefix(line, "@") {
			if current != nil {
				return nil, errors.New("anal: multiple blocks are not supported")
			}
			kind, fields, err := parseHeader(line)
			if err != nil {
				return nil, err
			}
			current = &block{kind: kind, fields: fields}
		} else {
			if current == nil {
				return nil, errors.New("anal: row before header")
			}
			current.rows = append(current.rows, line)
		}
	}

	if current == nil {
		return nil, nil
	}
	return build(current)
}

func parseHeader(line string) (string, []string, error) {
	left, right, found := strings.Cut(line[1:], ":")
	if !found {
		left, right = line[1:], ""
	}
	parts := strings.Fields(left)
	kind := ""
	if len(parts) > 0 {
		kind = parts[0]
	}
	fields, err := parseFields(right)
	if err != nil {
		return "", nil, err
	}
	return kind, fields, nil
}

func parseFields(text string) ([]string, error) {
	text = strings.TrimSpace(text)
	if text == "" || text == "{}" {
		return nil, nil
	}
	if strings.HasPrefix(text, "{") && strings.HasSuffix(text, "}") {
		parts := splitUnescaped(text[1:len(text)-1], ',')
		fields := make([]string, 0, len(parts))
		for _, part := range parts {
			if trimmed := strings.TrimSpace(part); trimmed != "" {
				fields = append(fields, trimmed)
			}
		}
		return fields, nil
	}
	return nil, errors.New("anal: bad field header")
}

func build(b *block) (any, error) {
	switch b.kind {
	case "list":
		out := make([]any, 0, len(b.rows))
		for _, row := range b.rows {
			out = append(out, decodeScalar(row))
		}
		return out, nil
	case "obj":
		if len(b.fields) == 0 {
			return map[string]any{}, nil
		}
		row := ""
		if len(b.rows) > 0 {
			row = b.rows[0]
		}
		return rowToObject(b.fields, row), nil
	case "table":
		if len(b.fields) == 0 {
			out := make([]any, len(b.rows))
			for i := range out {
				out[i] = map[string]any{}
			}
			return out, nil
		}
		out := make([]any, 0, len(b.rows))
		for _, row := range b.rows {
			out = append(out, rowToObject(b.fields, row))
		}
		return out, nil
	default:
		return nil, errors.New("anal: unknown block kind " + b.kind)
	}
}

func rowToObject(fields []string, row string) map[string]any {
	values := splitRow(row)
	out := make(map[string]any, len(fields))
	for i, field := range fields {
		var value any
		if i < len(values) {
			value = decodeScalar(values[i])
		}
		setPath(out, field, value)
	}
	return out
}

// setPath stores value at a dot-joined header path, rebuilding nested maps.
// Segments are unescaped so literal dots in key names survive the trip.
func setPath(root map[string]any, rawPath string, value any) {
	segments := splitUnescaped(rawPath, '.')
	current := root
	for _, segment := range segments[:len(segments)-1] {
		key := unescape(segment)
		next, ok := current[key].(map[string]any)
		if !ok {
			next = make(map[string]any)
			current[key] = next
		}
		current = next
	}
	current[unescape(segments[len(segments)-1])] = value
}

// splitUnescaped splits s on sep, treating an escaped sep (and any escaped
// char) as data. Backslashes are preserved for the caller to unescape.
func splitUnescaped(s string, sep byte) []string {
	var out []string
	var cur strings.Builder
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '\\':
			if i+1 < len(s) {
				cur.WriteByte(s[i])
				cur.WriteByte(s[i+1])
				i++
			} else {
				cur.WriteByte(s[i])
			}
		case sep:
			out = append(out, cur.String())
			cur.Reset()
		default:
			cur.WriteByte(s[i])
		}
	}
	out = append(out, cur.String())
	return out
}

func splitRow(row string) []string {
	return splitUnescaped(row, '|')
}

func decodeScalar(raw string) any {
	switch raw {
	case trueToken:
		return true
	case falseToken:
		return false
	case nullToken:
		return nil
	}
	if numberPattern.MatchString(raw) {
		if value, err := strconv.ParseFloat(raw, 64); err == nil {
			return value
		}
	}
	if strings.HasPrefix(raw, "`") && strings.HasSuffix(raw, "`") && len(raw) >= 2 {
		return unescape(raw[1 : len(raw)-1])
	}
	if strings.HasPrefix(raw, "&") {
		var value any
		if err := json.Unmarshal([]byte(unescape(raw[1:])), &value); err == nil {
			return value
		}
		return raw
	}
	return raw
}

func unescape(s string) string {
	var sb strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] != '\\' || i+1 >= len(s) {
			sb.WriteByte(s[i])
			continue
		}
		switch s[i+1] {
		case '\\':
			sb.WriteByte('\\')
		case '`':
			sb.WriteByte('`')
		case '|':
			sb.WriteByte('|')
		case 'n':
			sb.WriteByte('\n')
		case 't':
			sb.WriteByte('\t')
		case 'r':
			sb.WriteByte('\r')
		default:
			sb.WriteByte(s[i+1])
		}
		i++
	}
	return sb.String()
}
