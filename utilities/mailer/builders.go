package mailer

import(

	"github.com/mailgun/mailgun-go"

	"text/template"

	"encoding/json"
	"io/ioutil"
)

// A mailgun compatiable, send only client.
type Mailer struct{
	gun mailgun.Mailgun
	source string // The address this mailer sends from
	Templates map[string]*template.Template
}

// Creates a mailgun client that we can use.
//
// Priv and public are the keys handed out by mailgun.
// Domain is the domain assigned to the keypair.
func GetMailer(priv, pub, domain, sendingAddress string) *Mailer {
	gun:= mailgun.NewMailgun(domain, priv, pub)
	templateContainer:= make(map[string]*template.Template)
	return &Mailer{gun, sendingAddress, templateContainer}
}

// Acquires a mailer metadata from a file located on disk.
//
// Must be json encoded and match the format of MailGunMeta
func GetMailerFromFile(loc string) (*Mailer, error) {
	metaRaw, err:= ioutil.ReadFile(loc)
	if err!=nil {
		return nil, err
	}

	var meta MailGunMeta
	err = json.Unmarshal(metaRaw, &meta)
	if err!=nil {
		return nil, err	
	}

	mailer:= GetMailer(meta.PrivateKey, meta.PublicKey,
		meta.Domain, meta.SendingAddress)

	// Make sure we prepare all templates whose location are encoded
	// in the metadata.
	for id, loc:= range meta.Templates{
		err = mailer.Prepare(id, loc)
		if err!=nil {
			return nil, err	
		}
	}

	return mailer, nil
}