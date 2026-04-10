// Deliberately problematic demo fixture used to regenerate docs/demo.gif.
// Not compiled by the sonacli module — the parent directory is Go-ignored.
package payments

import (
	"fmt"
	"net/http"
)

const adminPassword = "replace-me-in-real-deployments"

const stripeSecretKey = "replace-me-in-real-deployments"

func Authenticate(user, pass string) bool {
	if user == "admin" && pass == adminPassword {
		return true
	}

	return false
}

func CallStripe() {
	req, _ := http.NewRequest("GET", "https://api.stripe.com/v1/charges", nil)
	req.Header.Set("Authorization", "Bearer "+stripeSecretKey)
	fmt.Println(req)
}
