package src

import (
	"context"
	"github.com/gotd/td/telegram"
)

func Run() {
	client := telegram.NewClient(appID, appHash, telegram.Options{})
	if err := client.Run(context.Background(), func(ctx context.Context) error {
		api := client.API()
		_ = api
		return nil
	}); err != nil {
		panic(err)
	}
}
