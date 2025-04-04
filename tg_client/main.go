package tg_client

import (
	"context"
	"fmt"
	"time"

	"github.com/gotd/td/session"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/tg"

	"tg-dl/config"
	"tg-dl/constants"
)

func InitClientWithConfig(cfg *config.Config) *telegram.Client {
	client := telegram.NewClient(cfg.ApiId, cfg.ApiHash, telegram.Options{
		SessionStorage: &session.FileStorage{
			Path: constants.SessionPath,
		},
		RetryInterval: 5 * time.Second,
	})

	return client
}

func MaybeAuth(ctx context.Context, client *telegram.Client, cfg *config.Config) error {
	status, err := client.Auth().Status(ctx)
	if err != nil {
		return fmt.Errorf("failed to get auth status: %w", err)
	}

	// If not authorized, do auth flow
	if !status.Authorized {
		flow := auth.NewFlow(
			auth.Constant(cfg.Phone, "", auth.CodeAuthenticatorFunc(func(ctx context.Context, sentCode *tg.AuthSentCode) (string, error) {
				fmt.Print("Enter code from Telegram: ")
				var code string
				fmt.Scanln(&code)
				return code, nil
			})),
			auth.SendCodeOptions{},
		)

		if err := flow.Run(ctx, client.Auth()); err != nil {
			return fmt.Errorf("auth flow failed: %w", err)
		}

		fmt.Println("Successfully authenticated!")
	}

	return nil
}
