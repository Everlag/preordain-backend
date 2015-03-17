package UserStructs

import "strings"

const PasswordResetTemplate string =
"name, if you requested a password reset token, here it is: PasswordResetToken" +
".\n\rIf you didn't make the request, please ignore this; if you are receiving these" +
" unsolicited at regular intervals, please contact me.\n\rThis token will expire in 24" +
" hours from the arrival of this email - only one token can be active at a given time."

const PasswordResetSubjectTemplate string =
"Password Reset Request for name"

// Attempts to send a message using the mailing credentials the manager was initialized
// with.
func (aManager *UserManager) sendPasswordResetMail(name, email, token string) error {
	
	fullMessage:= strings.Replace(PasswordResetTemplate, "name", name, -1)
	fullMessage = strings.Replace(fullMessage, "PasswordResetToken", token, -1)

	fullSubject:= strings.Replace(PasswordResetSubjectTemplate, "name", name, -1)

	return aManager.mailer.DispatchMail(name, email, fullSubject, fullMessage)

}