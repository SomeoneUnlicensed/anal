package anal_test

import (
	"fmt"

	"github.com/SomeoneUnlicensed/anal"
)

func Example() {
	rows := []map[string]any{
		{"id": 1, "sku": "P123"},
		{"id": 2, "sku": "P124"},
	}
	data, _ := anal.Marshal(rows)
	fmt.Print(string(data))
	// Output:
	// @table: {id,sku}
	// 1|P123
	// 2|P124
}
