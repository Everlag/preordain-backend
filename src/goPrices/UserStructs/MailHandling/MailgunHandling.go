package MailHandling

import(

	"github.com/mailgun/mailgun-go"

)

const credentialsLoc string = "mailgunMeta.json"

type Mailer struct{

	sendingAddress string

	mailer mailgun.Mailgun

}

func (aMailer *Mailer) DispatchMail(name, email, subject, body string) error {
	
	message:= aMailer.mailer.NewMessage(
	    aMailer.sendingAddress, // From
	    subject, // Subject
	    body,  // Plain-text body
	)

	err:= message.AddRecipient(name + "<" + email + ">")
	if err!=nil {
		return err
	}

	_, _, err = aMailer.mailer.Send(message)
	if err!=nil {
		return err
	}

	return nil

}

func NewMailer() (*Mailer, error) {
	
	creds, err:= getMailgunCredentials()
	if err!=nil {
		return &Mailer{}, err
	}

	mailer:= mailgun.NewMailgun(creds.Domain,
		creds.PrivateKey,
		creds.PublicKey)

	return &Mailer{creds.SendingAddress, mailer}, nil

}