package ebiten

import (
	"context"

	"github.com/davecgh/go-spew/spew"
)

func assert(ctx context.Context, condition bool, args ...any) {
	if condition {
		return
	}
	if len(args) == 0 {
		panic("assertion failed")
	}
	panic(spew.Sdump(args...))
}
