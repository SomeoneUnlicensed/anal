package anal

import (
	"errors"
	"fmt"
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
		parts := strings.Split(text[1:len(text)-1], ",")
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
			value, err := decodeValue(row)
			if err != nil {
				return nil, err
			}
			out = append(out, value)
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
		return rowToObject(b.fields, row)
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
			obj, err := rowToObject(b.fields, row)
			if err != nil {
				return nil, err
			}
			out = append(out, obj)
		}
		return out, nil
	default:
		return nil, errors.New("anal: unknown block kind " + b.kind)
	}
}

func rowToObject(fields []string, row string) (map[string]any, error) {
	values, err := parseRow(row)
	if err != nil {
		return nil, err
	}
	out := make(map[string]any, len(fields))
	for i, field := range fields {
		if i < len(values) {
			out[field] = values[i]
		} else {
			out[field] = nil
		}
	}
	return out, nil
}

// parseRow splits a "|"-delimited row into decoded values, treating nested
// {} and [] literals and backtick-quoted strings as opaque so that a
// literal "|" inside them never gets mistaken for a field separator.
func parseRow(row string) ([]any, error) {
	var out []any
	pos := 0
	for {
		value, next, err := parseValue(row, pos, "|")
		if err != nil {
			return nil, err
		}
		out = append(out, value)
		pos = next
		if pos >= len(row) {
			return out, nil
		}
		if row[pos] != '|' {
			return nil, fmt.Errorf("anal: expected '|' at offset %d in row %q", pos, row)
		}
		pos++
	}
}

// decodeValue parses s as a single, self-contained value (a list item, or a
// value nested inside an already-isolated field).
func decodeValue(s string) (any, error) {
	value, pos, err := parseValue(s, 0, "")
	if err != nil {
		return nil, err
	}
	if pos != len(s) {
		return nil, fmt.Errorf("anal: trailing data at offset %d in value %q", pos, s)
	}
	return value, nil
}

// parseValue parses one value starting at pos. stops names the bytes (beyond
// the closing "}" / "]" of a nested literal) that terminate a bare scalar in
// the current context: "|" between row fields, "" when the value spans the
// rest of s, or "," between object/array elements.
func parseValue(s string, pos int, stops string) (any, int, error) {
	if pos >= len(s) {
		return "", pos, nil
	}
	switch s[pos] {
	case '`':
		return parseQuoted(s, pos)
	case '{':
		return parseObject(s, pos)
	case '[':
		return parseArray(s, pos)
	default:
		end := pos
		for end < len(s) && !strings.ContainsRune(stops, rune(s[end])) {
			end++
		}
		return decodeScalar(s[pos:end]), end, nil
	}
}

func parseObject(s string, pos int) (any, int, error) {
	pos++ // skip '{'
	obj := map[string]any{}
	if pos < len(s) && s[pos] == '}' {
		return obj, pos + 1, nil
	}
	for {
		var key string
		if pos < len(s) && s[pos] == '`' {
			decoded, next, err := parseQuoted(s, pos)
			if err != nil {
				return nil, pos, err
			}
			key, pos = decoded.(string), next
		} else {
			start := pos
			for pos < len(s) && s[pos] != ':' {
				pos++
			}
			if pos >= len(s) {
				return nil, pos, fmt.Errorf("anal: expected ':' in object %q", s)
			}
			key = s[start:pos]
		}
		if pos >= len(s) || s[pos] != ':' {
			return nil, pos, fmt.Errorf("anal: expected ':' in object %q", s)
		}
		pos++ // skip ':'

		value, next, err := parseValue(s, pos, ",}")
		if err != nil {
			return nil, pos, err
		}
		obj[key] = value
		pos = next

		if pos >= len(s) {
			return nil, pos, fmt.Errorf("anal: unterminated object %q", s)
		}
		switch s[pos] {
		case ',':
			pos++
		case '}':
			return obj, pos + 1, nil
		default:
			return nil, pos, fmt.Errorf("anal: expected ',' or '}' in object %q", s)
		}
	}
}

func parseArray(s string, pos int) (any, int, error) {
	pos++ // skip '['
	arr := []any{}
	if pos < len(s) && s[pos] == ']' {
		return arr, pos + 1, nil
	}
	for {
		value, next, err := parseValue(s, pos, ",]")
		if err != nil {
			return nil, pos, err
		}
		arr = append(arr, value)
		pos = next

		if pos >= len(s) {
			return nil, pos, fmt.Errorf("anal: unterminated array %q", s)
		}
		switch s[pos] {
		case ',':
			pos++
		case ']':
			return arr, pos + 1, nil
		default:
			return nil, pos, fmt.Errorf("anal: expected ',' or ']' in array %q", s)
		}
	}
}

func parseQuoted(s string, pos int) (any, int, error) {
	pos++ // skip opening backtick
	start := pos
	for pos < len(s) {
		if s[pos] == '\\' && pos+1 < len(s) {
			pos += 2
			continue
		}
		if s[pos] == '`' {
			return unescape(s[start:pos]), pos + 1, nil
		}
		pos++
	}
	return nil, pos, fmt.Errorf("anal: unterminated quoted string %q", s)
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
