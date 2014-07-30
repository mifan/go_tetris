package utils

import (
	"fmt"

	"github.com/astaxie/beego/utils"
)

const emailConfTpl = `{"identity":"%s","username":"%s","password":"%s","host":"%s","port":%d,"from":"%s"}`

var emailConf string
var email = utils.NewEMail(emailConf)

// have to set it, otherwise you can not send email
func SetEmailConf(identity, username, password, host, from string, port int) {
	emailConf = fmt.Sprintf(emailConfTpl, identity, username, password, host, port, from)
}

func SendMail(text, subject string, to ...string) error {
	email.To = to
	email.Subject = subject
	email.Text = text
	return email.Send()
}
