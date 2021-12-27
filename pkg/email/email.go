package email

import (
	"bytes"
	"context"
	"html/template"
	"os"

	"github.com/mailgun/mailgun-go/v4"
)

func BuildMessage(body string, data interface{}) (string, error) {
	if data == nil {
		return body, nil
	}

	buf := new(bytes.Buffer)
	tpl := template.Must(template.New("email body").Parse(body))

	err := tpl.Execute(buf, data)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

func SendMessage(ctx context.Context, subject, body string, data interface{}, to ...string) error {
	mg, err := mailgun.NewMailgunFromEnv()
	if err != nil {
		return err
	}

	html, err := BuildMessage(body, data)
	if err != nil {
		return err
	}

	msg := mg.NewMessage(os.Getenv("MG_FROM"), subject, "", to...)
	msg.SetHtml(html)

	_, _, err = mg.Send(ctx, msg)
	return err
}
