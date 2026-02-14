package telegram

import (
	"context"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
)

type InvokeMiddleware struct {
	before func()
	after  func()
}

func NewInvokeMiddleware(before func(), after func()) *InvokeMiddleware {
	return &InvokeMiddleware{
		before: before,
		after:  after,
	}
}

func (m *InvokeMiddleware) Handle(next tg.Invoker) telegram.InvokeFunc {
	return func(ctx context.Context, input bin.Encoder, output bin.Decoder) error {
		m.before()
		defer m.after()
		return next.Invoke(ctx, input, output)
	}
}
