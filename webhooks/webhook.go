package lxWebhooks

const (
	Error = "error"
	Warn  = "warning"
	Info  = "info"
)

type IWebhook interface {
	SendSmall(title, msg, color string) ([]byte, error)
}
