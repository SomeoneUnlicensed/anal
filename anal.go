// Package anal implements ANAL, a compact, token-lean format for data
// flowing between agents and models.
//
// ANAL declares field names once in a header and then sends only values. It
// drops quotes and commas, and collapses booleans and null to single-token
// markers. It needs no model retraining: it wins by matching the tokenizer,
// not by fighting it.
//
// The supported shapes are flat objects and flat lists:
//
//	@table: {id,sku,qty}
//	1|P123|2
//	2|P124|1
//
//	@obj: {event,order_id}
//	order.created|8123
//
//	@list:
//	code
//	coffee
//
// Marshal and Unmarshal follow encoding/json conventions: structs and maps
// become objects, slices become tables or lists, and json struct tags are
// honored. Nested objects and arrays are not supported in ANAL v1 and
// return an error.
package anal

import (
	"encoding/json"
	"fmt"
	"reflect"
)

// Marshal returns the ANAL encoding of v.
//
// v must be a flat object or a flat list at the top level. Structs, maps,
// slices, and scalars are normalized with encoding/json, so json struct
// tags are honored. Nested values return an error.
func Marshal(v any) ([]byte, error) {
	normalized, err := normalize(v)
	if err != nil {
		return nil, err
	}
	switch value := normalized.(type) {
	case map[string]any:
		return encodeObject(value)
	case []any:
		return encodeList(value)
	default:
		return nil, fmt.Errorf("anal: top-level value must be an object or array, got %T", normalized)
	}
}

// Unmarshal parses ANAL-encoded data and stores the result in v.
func Unmarshal(data []byte, v any) error {
	parsed, err := parse(string(data))
	if err != nil {
		return err
	}
	if target, ok := v.(*any); ok {
		*target = parsed
		return nil
	}
	jsonBytes, err := json.Marshal(parsed)
	if err != nil {
		return err
	}
	return json.Unmarshal(jsonBytes, v)
}

func normalize(v any) (any, error) {
	jsonBytes, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	var out any
	if err := json.Unmarshal(jsonBytes, &out); err != nil {
		return nil, err
	}
	if out == nil && reflect.TypeOf(v) != nil && reflect.TypeOf(v).Kind() == reflect.Slice {
		out = []any{}
	}
	return out, nil
}
