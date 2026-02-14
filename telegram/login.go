package telegram

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/gotd/contrib/bg"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
)

var globalClient *telegram.Client
var stopClient bg.StopFunc // 用于停止客户端连接的函数

func Login(ctx context.Context) {
	uuid := uuid.New().String()

	client := telegram.NewClient(AppID, AppHash, telegram.Options{
		Logger:         logger,
		SessionStorage: GetSessionStorage(uuid),
	})

	needLogin := !SessionExists(uuid)

	// 如果需要登录，先使用 client.Run 进行认证
	if needLogin {
		// 创建一个新的 context 用于登录
		loginCtx := context.Background()
		if err := client.Run(loginCtx, func(ctx context.Context) error {
			// 进行验证登陆
			phoneAuth := &terminalAuth{}
			flow := auth.NewFlow(phoneAuth, auth.SendCodeOptions{
				AllowFlashCall: false,
				CurrentNumber:  false,
				AllowAppHash:   false,
			})
			if err := client.Auth().IfNecessary(ctx, flow); err != nil {
				fmt.Printf("认证失败: %v\n", err)
				return err
			}
			fmt.Println("登陆成功")
			return nil
		}); err != nil {
			fmt.Printf("运行失败: %v\n", err)
		}
	}
}
