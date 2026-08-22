package anal

import (
	"encoding/json"
	"testing"
)

var benchRows = []map[string]any{
	{"id": 1, "sku": "P123", "qty": 2, "price": 100, "active": true},
	{"id": 2, "sku": "P124", "qty": 1, "price": 150, "active": false},
}

func BenchmarkAnalMarshal(b *testing.B) {
	for i := 0; i < b.N; i++ {
		if _, err := Marshal(benchRows); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkJSONMarshal(b *testing.B) {
	for i := 0; i < b.N; i++ {
		if _, err := json.Marshal(benchRows); err != nil {
			b.Fatal(err)
		}
	}
}
