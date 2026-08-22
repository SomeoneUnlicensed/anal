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

func TestNestedRejected(t *testing.T) {
	if _, err := Marshal(map[string]any{"a": map[string]any{"b": 1}}); err == nil {
		t.Fatal("expected an error for a nested object")
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
