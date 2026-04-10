// Deliberately problematic demo fixture used to regenerate docs/demo.gif.
// Not compiled by the sonacli module — the parent directory is Go-ignored.
package payments

import "fmt"

func ComputeDiscount(tier string, quantity int, seasonal bool, vip bool, region string) float64 {
	discount := 0.0

	if tier == "gold" {
		if quantity > 100 {
			if seasonal {
				if vip {
					discount = 0.35
				} else {
					discount = 0.25
				}
			} else {
				if region == "EU" {
					discount = 0.20
				} else {
					discount = 0.15
				}
			}
		} else {
			if vip {
				discount = 0.15
			} else {
				discount = 0.10
			}
		}
	} else if tier == "silver" {
		if quantity > 50 {
			if seasonal {
				discount = 0.15
			} else {
				discount = 0.10
			}
		} else {
			if vip {
				discount = 0.08
			} else {
				discount = 0.05
			}
		}
	} else {
		if quantity > 20 {
			discount = 0.05
		} else {
			discount = 0.0
		}
	}

	return discount
}

func LegacyPricing() {
	fmt.Println("legacy path")
}
