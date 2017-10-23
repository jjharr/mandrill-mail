# MandrillMail - Full featured client for Mandrill

Mandrill Mail is a full-featured implementation of the Send capabilities for Mandrill. It contains
an interface with functions for three common use cases with increasing complexity.

## Getting Started

You must set your own Mandrill key, test email addresses, and domain to use or test this package.
See notes in mandrill_test.go.

```
go get jjharr/mandrill-mail
```

## Usage

Send a simple freeform email.
```
func (m *mandrill) SimpleMail(from, to, subject, body string) (*MailRecipientResponse, error)
```

Send a templated email. Allows for template (i.e. merge) variables.
```
func (m *mandrill) TemplateMail(toEmail string, subject string, template string, vars map[string]string) (*MailRecipientResponse, error)
```

Send mail to multiple recipients, with template/merge variables. Allows for async sending, tracking
of opens and clicks, and more.
```
func (m *mandrill) BulkMail(recipients []MailRecipient, message *MailMessage, params *SendParams) ([]MailRecipientResponse, error)
```
