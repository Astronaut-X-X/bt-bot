package telegram

import (
	"context"
	"fmt"

	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/tg"
)

// 终端登录器
type terminalAuth struct {
	phone string
}

func (a *terminalAuth) Phone(_ context.Context) (string, error) {
	fmt.Print("请输入手机号（带国家码，如+8613800138000）: ")
	var phone string
	fmt.Scanln(&phone)
	return phone, nil
}

func (a *terminalAuth) Password(_ context.Context) (string, error) {
	fmt.Print("请输入密码: ")
	var password string
	fmt.Scanln(&password)
	return password, nil
}

func (a *terminalAuth) Code(_ context.Context, _ *tg.AuthSentCode) (string, error) {
	fmt.Print("请输入验证码: ")
	var code string
	fmt.Scanln(&code)
	return code, nil
}

func (a *terminalAuth) AcceptTermsOfService(_ context.Context, tos tg.HelpTermsOfService) error {
	fmt.Println("需要接受服务条款: ", tos.Text)
	fmt.Println("（按回车继续）")
	fmt.Scanln()
	return nil
}

func (a *terminalAuth) SignUp(_ context.Context) (auth.UserInfo, error) {
	return auth.UserInfo{}, fmt.Errorf("注册暂不支持")
}

func (a *terminalAuth) AcceptLoginToken(_ context.Context, _ tg.AuthLoginToken) error {
	return nil
}
