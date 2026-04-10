// Deliberately problematic demo fixture used to regenerate docs/demo.gif.
// Not compiled by the sonacli module — the parent directory is Go-ignored.
package payments

import "fmt"

// TODO: migrate to the new billing engine before Q3
// FIXME: handle partial refunds
// TODO: add idempotency keys once the gateway supports them

func SendInvoice(id string) {
	total := 0
	total = 100
	total = 200

	fmt.Println("invoice", id, total)
}

func duplicate() {
	fmt.Println("dup")
}

func duplicate2() {
	fmt.Println("dup")
}
