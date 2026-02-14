package main

import (
	"context"

	"bt-bot/telegram"
)

func main() {
	ctx := context.Background()
	telegram.Login(ctx)
}
