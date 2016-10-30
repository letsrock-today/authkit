package confirmer

import (
	"bytes"
	"fmt"
	"html/template"
	"net/smtp"

	"github.com/letsrock-today/hydra-sample/sample/authkit/backend/config"
)

const (
	emailTemplate = `From: {{.From}}
To: {{.To}}
Subject: {{.Subject}}

{{.Body}}
 
--
Best,
{{.From}}`
)

func Send(to, subj, text string) error {
	var err error
	t := template.New("emailTemplate")
	if t, err = t.Parse(emailTemplate); err != nil {
		return err
	}

	c := config.Get().EmailConfig

	var b bytes.Buffer
	if err = t.Execute(&b, struct {
		From    string
		To      string
		Subject string
		Body    string
	}{
		From:    c.Sender,
		To:      to,
		Subject: subj,
		Body:    text,
	}); err != nil {
		return err
	}

	auth := smtp.PlainAuth("", c.Sender, c.SenderPass, c.MailServer)
	err = smtp.SendMail(
		fmt.Sprintf("%s:%s", c.MailServer, c.MailPort),
		auth,
		c.Sender,
		[]string{to},
		b.Bytes())
	return err
}
