package main

import (
	"fmt"

	"example.com/sonacli/sample-basic/internal/mathutil"
)

func main() {
	// TODO: replace this sample calculation with a real workflow.
	values := []int{3, 4, 5}
	fmt.Printf("sample total: %d\n", mathutil.Sum(values))
}
