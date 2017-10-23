package mandrillmail

import (
	"errors"
	"html/template"
	"time"
)

// NOTE : The public types in this file are intended to be generic types, not specific to Mandrill, that
// are suitable for use in a general purpose Mail interface (such as what we've implemented for Globio).
// That is why there is overlap between these types and the private, mandrill-specific types in mandrill.go

// Mailer is a generic mail interface for common use cases of sending email from an application
type Mailer interface {
	BulkMail(recipients []MailRecipient, message *MailMessage, params *SendParams) ([]MailRecipientResponse, error)
	TemplateMail(toEmail string, subject string, template *template.Template, vars map[string]string) (*MailRecipientResponse, error)
	SimpleMail(from, to, subject, body string) (*MailRecipientResponse, error)
}

type MailRecipientType string

const (
	MAIL_TO  MailRecipientType = `to`
	MAIL_CC  MailRecipientType = `cc`
	MAIL_BCC MailRecipientType = `bcc`
)

type MailStatus string

const (
	MAIL_MESSAGE_SENT      MailStatus = `sent`
	MAIL_MESSAGE_QUEUED    MailStatus = `queued`
	MAIL_MESSAGE_SCHEDULED MailStatus = `scheduled`
	MAIL_MESSAGE_REJECTED  MailStatus = `rejected`
	MAIL_MESSAGE_INVALID   MailStatus = `invalid`
	MAIL_MESSAGE_UNKNOWN   MailStatus = `unknown`
)

type MailRecipient struct {
	Name          string
	Email         string
	RecipientType MailRecipientType
	// only really useful when tracking is on
	Metadata map[string]string
}

func (mr *MailRecipient) validate() error {

	if mr.Email == `` {
		return errors.New("The recipient email is required")
	}

	if mr.RecipientType == `` {
		return errors.New("The recipient type must be set")
	}

	return nil
}

type EmailAttachment struct {
	Name          string
	MimeType      string
	Base64Content string
}

// tags vs metadata (@see https://mandrill.zendesk.com/hc/en-us/articles/205582467-How-to-Use-Tags-in-Mandrill)
// tags : mandrill aggregates stats for tags, but not for metadata, tags are limited to 100 lifetime so big static categories, kept indefinitely
// uses : email type, region, customer type
//
// metadata : searchable, returned in webhooks, but doesn't aggregate with stats, has finite lifetime
// uses : country, model data like booking #, cancellation #, merchant/user id, language, etc.
type MailMessage struct {
	HTMLTemplate  *template.Template
	TextTemplate  *template.Template
	TemplateVars  map[string]string
	AutoText      bool
	Subject       string
	From          *MailRecipient
	ReplyTo       string
	Attachments   []EmailAttachment
	Images        []EmailAttachment
	MarkImportant bool
	Tags          []string
	Metadata      map[string]string
}

func (mm *MailMessage) validate() error {

	if mm.HTMLTemplate == nil && mm.TextTemplate == nil {
		return errors.New("Must set HTMLTemplate or TextTemplate for message")
	}

	if mm.Subject == `` {
		return errors.New("Must set Subject for message")
	}

	if mm.From == nil {
		return errors.New("Must set From for message")
	}

	return nil
}

type SendParams struct {
	SendAsync   bool
	SendAt      *time.Time
	IpPool      string
	TrackOpens  bool
	TrackClicks bool
}

// DefaultSendMailParams returns a default set of SendParams that are good for TemplateMail and SimpleMail. You
// probably want custom params for BulkMail.
func DefaultSendMailParams() *SendParams {

	n := time.Now()
	return &SendParams{
		SendAsync:   true,
		SendAt:      &n,
		IpPool:      ``,
		TrackOpens:  false,
		TrackClicks: false,
	}
}

type MailRecipientResponse struct {
	Id     string
	Email  string
	Status MailStatus
	Error  string
}
