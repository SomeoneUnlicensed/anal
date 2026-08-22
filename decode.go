package anal

import (
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
		if i < len(values) {
			out[field] = decodeScalar(values[i])
		} else {
			out[field] = nil
		}
	}
	return out
}

func splitRow(row string) []string {
	var out []string
	var cur strings.Builder
	for i := 0; i < len(row); i++ {
		switch row[i] {
		case '\\':
			if i+1 < len(row) {
				cur.WriteByte(row[i])
				cur.WriteByte(row[i+1])
				i++
			} else {
				cur.WriteByte(row[i])
			}
		case '|':
			out = append(out, cur.String())
			cur.Reset()
		default:
			cur.WriteByte(row[i])
		}
	}
	out = append(out, cur.String())
	return out
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
