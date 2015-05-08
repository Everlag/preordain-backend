package mailer

import(
	"text/template"
)


// Takes a name and email and sets it to form name <email>
func FormatAddress(name, email string) string {
	return name + "<" + email + ">"
}

type MailGunMeta struct{
	PrivateKey, PublicKey string
	Domain, SendingAddress string
	Templates map[string]string
}

func FetchTemplate(loc string) (*template.Template, error) {
	return template.ParseFiles(loc)
}

// Prepares a template given a location on disk.
//
//It may be access by mailer.Templates[id]
func (mailer *Mailer) Prepare(id, loc string) error {
	template, err:= FetchTemplate(loc)
	if err!=nil {
		return err
	}

	mailer.Templates[id] = template

	return nil
}