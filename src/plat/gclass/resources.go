package gclass

import (
	"codeberg.org/kvo/std/errors"

	"main/plat"
)

// Get a list of resources from Google Classroom for a user.
func ListRes(creds User, r chan []plat.Resource, e chan []errors.Error) {
	r <- nil
	e <- nil
	return
}
