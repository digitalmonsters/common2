package ses

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/digitalmonsters/go-common/boilerplate"
)

type EmailSender struct {
	config    *boilerplate.SESConfig
	session   *session.Session
	sesClient *ses.SES
}

func NewEmailSender(cfg *boilerplate.SESConfig) *EmailSender {
	u := &EmailSender{
		config: cfg,
	}
	return u
}

func (s *EmailSender) Send(request *ses.SendEmailInput) (*ses.SendEmailOutput, error) {
	client, err := s.getClient()
	if err != nil {
		return nil, err
	}

	return client.SendEmail(request)
}

func (s *EmailSender) getClient() (*ses.SES, error) {
	if s.session == nil {
		if sess, err := session.NewSession(&aws.Config{Region: aws.String(s.config.Region)}); err != nil {
			return nil, err
		} else {
			s.session = sess
		}
	}

	if s.sesClient == nil {
		s.sesClient = ses.New(s.session)
	}
	return s.sesClient, nil
}
