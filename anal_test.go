package anal

import (
	"reflect"
	"testing"
)

type order struct {
	Event   string  `json:"event"`
	OrderID int     `json:"order_id"`
	Status  string  `json:"status"`
	Total   float64 `json:"total"`
}

type item struct {
	ID     int     `json:"id"`
	SKU    string  `json:"sku"`
	Qty    int     `json:"qty"`
	Price  float64 `json:"price"`
	Active bool    `json:"active"`
}

func TestStructRoundTrip(t *testing.T) {
	want := order{Event: "order.created", OrderID: 8123, Status: "paid", Total: 199.9}
	data, err := Marshal(want)
	if err != nil {
		t.Fatal(err)
	}
	var got order
	if err := Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("round-trip mismatch:\n got: %#v\nwant: %#v", got, want)
	}
}

func TestTableRoundTrip(t *testing.T) {
	want := []item{
		{ID: 1, SKU: "P123", Qty: 2, Price: 100, Active: true},
		{ID: 2, SKU: "P124", Qty: 1, Price: 150, Active: false},
	}
	data, err := Marshal(want)
	if err != nil {
		t.Fatal(err)
	}
	var got []item
	if err := Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("round-trip mismatch:\n got: %#v\nwant: %#v", got, want)
	}
}

func TestListRoundTrip(t *testing.T) {
	want := []string{"code", "coffee"}
	data, err := Marshal(want)
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	if err := Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("round-trip mismatch:\n got: %#v\nwant: %#v", got, want)
	}
}

func TestAmbiguousStrings(t *testing.T) {
	input := map[string]any{
		"number_like": "2",
		"bool_like":   "T",
		"empty":       "",
		"pipe":        "a|b",
		"backtick":    "`x`",
		"real_number": 42,
	}
	data, err := Marshal(input)
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]any
	if err := Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	want := map[string]any{
		"number_like": "2",
		"bool_like":   "T",
		"empty":       "",
		"pipe":        "a|b",
		"backtick":    "`x`",
		"real_number": float64(42),
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("round-trip mismatch:\n got: %#v\nwant: %#v", got, want)
	}
}

func TestEmpty(t *testing.T) {
	data, err := Marshal([]string{})
	if err != nil {
		t.Fatal(err)
	}
	var gotList []string
	if err := Unmarshal(data, &gotList); err != nil {
		t.Fatal(err)
	}
	if len(gotList) != 0 {
		t.Fatalf("expected empty list, got %#v", gotList)
	}

	data, err = Marshal(map[string]int{})
	if err != nil {
		t.Fatal(err)
	}
	var gotObject map[string]any
	if err := Unmarshal(data, &gotObject); err != nil {
		t.Fatal(err)
	}
	if len(gotObject) != 0 {
		t.Fatalf("expected empty object, got %#v", gotObject)
	}
}

func TestEmptyObjectsTable(t *testing.T) {
	input := []map[string]any{{}, {}}
	data, err := Marshal(input)
	if err != nil {
		t.Fatal(err)
	}
	var got []map[string]any
	if err := Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 empty objects, got %#v", got)
	}
}

func TestNestedRoundTrip(t *testing.T) {
	want := map[string]any{
		"event": "order.created",
		"meta": map[string]any{
			"tags": []any{"vip", "rush"},
			"address": map[string]any{
				"city": "NYC",
				"zip":  float64(10001),
			},
		},
		"empty_obj": map[string]any{},
		"empty_arr": []any{},
	}
	data, err := Marshal(want)
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]any
	if err := Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal %q: %v", data, err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("round-trip mismatch:\n got: %#v\nwant: %#v", got, want)
	}
}

func TestNestedTableColumn(t *testing.T) {
	want := []any{
		map[string]any{"id": float64(1), "tags": []any{"a", "b|c", "d,e"}},
		map[string]any{"id": float64(2), "tags": []any{}},
	}
	data, err := Marshal(want)
	if err != nil {
		t.Fatal(err)
	}
	var got []any
	if err := Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal %q: %v", data, err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("round-trip mismatch:\n got: %#v\nwant: %#v", got, want)
	}
}

func TestMixedListRoundTrip(t *testing.T) {
	want := []any{
		float64(1),
		"two",
		map[string]any{"three": float64(3)},
		[]any{float64(4), float64(5)},
		nil,
	}
	data, err := Marshal(want)
	if err != nil {
		t.Fatal(err)
	}
	var got []any
	if err := Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal %q: %v", data, err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("round-trip mismatch:\n got: %#v\nwant: %#v", got, want)
	}
}

func TestNestedAmbiguousKey(t *testing.T) {
	want := map[string]any{
		"obj": map[string]any{
			"a:b": "colon key",
			"":    "empty key",
		},
	}
	data, err := Marshal(want)
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]any
	if err := Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal %q: %v", data, err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("round-trip mismatch:\n got: %#v\nwant: %#v", got, want)
	}
}

func TestBareBracketStringQuoted(t *testing.T) {
	want := map[string]any{"note": "[not an array]"}
	data, err := Marshal(want)
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]any
	if err := Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal %q: %v", data, err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("round-trip mismatch:\n got: %#v\nwant: %#v", got, want)
	}
}

func TestGoldenOutput(t *testing.T) {
	tests := []struct {
		name  string
		value any
		want  string
	}{
		{
			name:  "table",
			value: []map[string]any{{"id": 1, "sku": "P123"}, {"id": 2, "sku": "P124"}},
			want:  "@table: {id,sku}\n1|P123\n2|P124\n",
		},
		{
			name:  "object",
			value: map[string]any{"event": "order.created", "id": 8123},
			want:  "@obj: {event,id}\norder.created|8123\n",
		},
		{
			name:  "list",
			value: []string{"code", "coffee"},
			want:  "@list:\ncode\ncoffee\n",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			data, err := Marshal(test.value)
			if err != nil {
				t.Fatal(err)
			}
			if string(data) != test.want {
				t.Fatalf("got %q, want %q", data, test.want)
			}
		})
	}
}
