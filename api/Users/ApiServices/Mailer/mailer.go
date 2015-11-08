package mailer

import(
	"text/template"
	"bytes"
)

// Sends plaintext via mailgun.
//
// body must be plaintext, no html. Format as desired.
// to should be form NAME <EMAIL>
// subject should be succinct.
func (mailer *Mailer) Send(body, to, subject string) error {
	
	var err error

	m:= mailer.gun.NewMessage(mailer.source, subject, body)
	err = m.AddRecipient(to)
	if err!=nil {
		return err
	}

	_,_, err = mailer.gun.Send(m)
	
	return err

}

// Sends plaintext via mailgun.
//
// This allows easy access to the prepared templates
// associated with this mailer
func (mailer *Mailer) SendPrepared(templateId string, content interface{},
	to, subject string) error {

	return mailer.SendTemplated(mailer.Templates[templateId], content,
		to, subject)
}

// Sends plaintext via mailgun.
//
// This one supports efficient text templating for the body.
func (mailer *Mailer) SendTemplated(bodyTemplate *template.Template,
	content interface{},
	to, subject string) error {

	var bodyBuffer bytes.Buffer 
    err:= bodyTemplate.Execute(&bodyBuffer, content)
    if err!=nil {
     	return err
     } 
    body:= bodyBuffer.String()

    return mailer.Send(body, to, subject)

}