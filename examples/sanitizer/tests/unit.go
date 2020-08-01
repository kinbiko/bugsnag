package unit

import (
	"context"
	"fmt"

	"github.com/kinbiko/bugsnag"
)

type Unit struct {
	snag *bugsnag.Notifier
}

func NewUnit(n *bugsnag.Notifier) *Unit {
	return &Unit{snag: n}
}

func (u *Unit) DoDangerousOperation(ctx context.Context, shouldFail bool) {
	if shouldFail {
		u.snag.Notify(ctx, u.snag.Wrap(ctx, fmt.Errorf("something happened in the dangerous operation")))
	}
}
