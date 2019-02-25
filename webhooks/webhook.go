package lxWebhooks

type IWebhook interface {
	SendSmall(title, msg, color string) ([]byte, error)
}
