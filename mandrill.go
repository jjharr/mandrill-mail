package mandrillmail

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"html/template"
	"net/http"
	"strings"
	"time"
)

// @see https://www.mandrillapp.com/api/docs/messages.JSON.html for API documentation

// todo better email and struct validation

const (
	MANDRILL_BASE_URL     = `https://mandrillapp.com/api/1.0`
	MANDRILL_MESSAGE_PATH = `/messages/send.json`
)

type mandrillParams struct {
	Key       string           `json:"key"`
	Message   *mandrillMessage `json:"message"`
	Async     bool             `json:"async"`
	IpPool    string           `json:"ip_pool"`
	SendAtTxt string           `json:"send_at"`
}

// validate checks required parameters
func (mp *mandrillParams) validate() error {
	if len(mp.Key) == 0 {
		return errors.New("The message key identifies this transaction and must be set")
	}

	if mp.Message == nil {
		return errors.New("The message must be set")
	}

	return nil
}

type mandrillMessage struct {
	Html                    string                      `json:"html"`
	Text                    string                      `json:"text"`
	Subject                 string                      `json:"subject"`
	FromEmail               string                      `json:"from_email"`
	FromName                string                      `json:"from_name"`
	To                      []mandrillRecipient         `json:"to"`
	Headers                 map[string]string           `json:"headers"`
	MarkImportant           bool                        `json:"important"`
	TrackOpens              bool                        `json:"track_opens"`
	TrackClicks             bool                        `json:"track_clicks"`
	AutoText                bool                        `json:"auto_text"`
	AutoHtml                bool                        `json:"auto_html"`
	InlineCss               bool                        `json:"inline_css"`
	StripQueryString        bool                        `json:"url_strip_qs"`
	PreserveRecipients      bool                        `json:"preserve_recipients"`
	ViewContentLink         bool                        `json:"view_content_link"`
	BccAddress              string                      `json:"bcc_address"`
	TrackingDomain          string                      `json:"tracking_domain"`
	SigningDomain           string                      `json:"signing_domain"`
	ReturnPathDomain        string                      `json:"return_path_domain"`
	Merge                   bool                        `json:"merge"`
	MergeLanguage           string                      `json:"merge_language"`
	GlobalMergeVars         []mandrillMergeVar          `json:"global_merge_vars"`
	MergeVars               []mandrillRecipientMergeVar `json:"merge_vars"`
	Tags                    []string                    `json:"tags"`
	GoogleAnalyticsDomains  []string                    `json:"google_analytics_domains"`
	GoogleAnalyticsCampaign string                      `json:"google_analytics_campaign"`
	Metadata                map[string]string           `json:"metadata"`
	RecipientMetadata       []mandrillRecipientMetadata `json:"recipient_metadata"`
	Attachments             []mandrillAttachment        `json:"attachments"`
	Images                  []mandrillAttachment        `json:"images"`
}

// validate checks required parameters
func (mm *mandrillMessage) validate() error {

	if mm.Html == `` && mm.Text == `` {
		return errors.New("Must set Html or Text for message")
	}

	if mm.Subject == `` {
		return errors.New("Must set Subject for message")
	}

	if mm.FromEmail == `` {
		return errors.New("Must set FromEmail for message")
	}

	return nil
}

type mandrillRecipient struct {
	Email         string            `json:"email"`
	Name          string            `json:"name"`
	RecipientType MailRecipientType `json:"type"`
}

// validate checks required parameters
func (mr *mandrillRecipient) validate() error {
	if mr.Email == `` {
		return errors.New("The recipient email is required")
	}

	if mr.RecipientType == `` {
		return errors.New("The recipient type must be set")
	}

	return nil
}

type mandrillMergeVar struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}

type mandrillRecipientMergeVar struct {
	Rcpt string             `json:"rcpt"`
	Vars []mandrillMergeVar `json:"vars"`
}

type mandrillRecipientMetadata struct {
	Rcpt   string            `json:"rcpt"`
	Values map[string]string `json:"values"`
}

type mandrillAttachment struct {
	MimeType string `json:"type"`
	// the Content ID of the image - use <img src="cid:THIS_VALUE"> to reference the image in your HTML content
	Name          string `json:"name"`
	Base64Content string `json:"content"`
}

type mandrillRecipientResponse struct {
	Email        string     `json:"email"`
	Status       MailStatus `json:"status"`
	RejectReason string     `json:"reject_reason"`
	Id           string     `json:"_id"`
}

type mandrillResponse struct {
	response []mandrillRecipientResponse
}

type testResponse map[string]interface{}

type mandrillErrorResponse struct {
	Status  string `json:"status"`
	Code    int    `json:"code"` // Mandrill docs say this is a string, but it's an int.
	Name    string `json:"name"`
	Message string `json:"message"`
}

type mandrill struct {
	key           string
	domain        string
	defaultSender *MailRecipient
	client        *http.Client
}

var _ Mailer = new(mandrill)

// Mandrill constructor. The signature is purposely kept minimal so it can be easily created
// from a variety of application contexts
func NewMandrill(apiKey string, domain string, sender *MailRecipient, client *http.Client) (*mandrill, error) {

	if len(apiKey) == 0 {
		return nil, errors.New("API key is required")
	}

	if len(domain) == 0 {
		return nil, errors.New("domain is required")
	}

	if sender == nil {
		return nil, errors.New("sender is required")
	}

	if client == nil {
		return nil, errors.New("must set non-nil http client")
	}

	return &mandrill{
		key:           apiKey,
		domain:        domain,
		defaultSender: sender,
		client:        client,
	}, nil
}

// SimpleMail just sends a very simple email with the body you supply
func (m *mandrill) SimpleMail(from, to, subject, body string) (*MailRecipientResponse, error) {

	from = strings.TrimSpace(from)
	to = strings.TrimSpace(to)
	subject = strings.TrimSpace(subject)

	if from == `` {
		return nil, errors.New("SimpleMail: Must specify source email address;")
	}
	if to == `` {
		return nil, errors.New("SimpleMail: Must specify destination email address;")
	}
	if subject == `` {
		return nil, errors.New("SimpleMail: Must specify subject;")
	}

	msg := &mandrillMessage{
		InlineCss:     false,
		TrackClicks:   false,
		Subject:       subject,
		To:            []mandrillRecipient{{Email: string(to)}},
		FromEmail:     string(from),
		FromName:      to,
		Headers:       map[string]string{`Reply-To`: from},
		MarkImportant: true,
		Tags:          []string{},
		Text:          body,
	}

	var params SendParams
	mandrillParams := m.buildParams(&params, msg)

	resp, err := m.send(mandrillParams)
	if err != nil {
		return nil, err
	}

	if len(resp) == 0 {
		return nil, errors.New("SimpleMail: Received zero length response;")
	}

	return &resp[0], nil
}

// TemplateMail sends a templated email to a single recipient. Interpolates the supplied vars into the supplied template.
func (m *mandrill) TemplateMail(toEmail string, subject string, template *template.Template, vars map[string]string) (*MailRecipientResponse, error) {

	if toEmail == `` {
		return nil, errors.New("TemplateMail: Must specify destination email address;")
	} else if template == nil {
		return nil, errors.New("TemplateMail: Must specify template;")
	}

	recipients := []MailRecipient{
		{
			Name:          ``,
			Email:         string(toEmail),
			RecipientType: MAIL_TO,
		},
	}

	message := &MailMessage{
		HTMLTemplate: template,
		TemplateVars: vars,
		Subject:      subject,
	}

	msg, err := m.buildMessage(recipients, message)
	if err != nil {
		return nil, err
	}

	var params SendParams
	mandrillParams := m.buildParams(&params, msg)
	if err := mandrillParams.validate(); err != nil {
		return nil, err
	}

	resp, err := m.send(mandrillParams)
	if err != nil {
		return nil, err
	}

	if len(resp) == 0 {
		return nil, errors.New("TemplateMail: Received zero length response;")
	}

	return &resp[0], nil
}

// BulkMail sends an email to potentially many recipients. Allows the sender to set SendParams to track opens
// and clicks and other settings.
func (m *mandrill) BulkMail(recipients []MailRecipient, message *MailMessage, params *SendParams) ([]MailRecipientResponse, error) {

	err := m.validateMessageAndRecipients(recipients, message)
	if err != nil {
		return nil, err
	}

	msg, err := m.buildMessage(recipients, message)
	if err != nil {
		return nil, err
	}
	msg.TrackOpens = params.TrackOpens
	msg.TrackClicks = params.TrackClicks

	mandrillParams := m.buildParams(params, msg)
	if err := mandrillParams.validate(); err != nil {
		return nil, err
	}

	resp, err := m.send(mandrillParams)
	if err != nil {
		return nil, err
	}

	if len(resp) == 0 {
		return nil, err
	}

	return resp, nil
}

// validateMessageAndRecipients
func (m *mandrill) validateMessageAndRecipients(recipients []MailRecipient, message *MailMessage) error {

	var e error

	for _, v := range recipients {
		if e = v.validate(); e != nil {
			return e
		}
	}

	if e = message.validate(); e != nil {
		return e
	}

	return nil
}

// defaultMessage sets some common defaults for a mandrillMessage, but is not sufficient for sending.
func (m *mandrill) defaultMessage() *mandrillMessage {
	return &mandrillMessage{
		InlineCss:        true,
		TrackClicks:      true,
		TrackOpens:       true,
		StripQueryString: true,
		TrackingDomain:   m.domain,
		SigningDomain:    m.domain,
		ReturnPathDomain: m.domain,
	}
}

// buildMessageContent builds the content of the email from the supplied template.
func (m *mandrill) buildMessageContent(msg *MailMessage, vars map[string]string) (string, string, error) {

	var (
		htmlBuf = new(bytes.Buffer)
		textBuf = new(bytes.Buffer)
	)

	// validation ensures that either the HTML or the Text template is set
	if msg.HTMLTemplate != nil {
		if err := msg.HTMLTemplate.Execute(htmlBuf, vars); err != nil {
			return ``, ``, err
		}
	}

	// if autotext is true, we take that as precedence over a non-nil text template
	if msg.TextTemplate != nil && !msg.AutoText {
		if err := msg.HTMLTemplate.Execute(textBuf, vars); err != nil {
			return ``, ``, err
		}
	}

	return htmlBuf.String(), textBuf.String(), nil
}

// buildMessage builds a Mandrill-formatted message for sending
func (m *mandrill) buildMessage(recipients []MailRecipient, message *MailMessage) (*mandrillMessage, error) {

	msg := m.defaultMessage()

	msg.Subject = message.Subject

	// set from & related
	m.setMessageFrom(message, msg)

	// misc
	msg.MarkImportant = message.MarkImportant
	msg.Metadata = message.Metadata
	msg.Tags = message.Tags

	// set email content
	if err := m.setMessageContent(message, msg); err != nil {
		return nil, err
	}

	// set recipients
	if err := m.setMessageRecipients(recipients, message, msg); err != nil {
		return nil, err
	}

	// attachments
	var mattach []mandrillAttachment
	if len(message.Attachments) > 0 {
		mattach = make([]mandrillAttachment, len(message.Attachments), len(message.Attachments))
		for i := range message.Attachments {
			mattach[i] = m.mailAttachmentToMandrillAttachment(message.Attachments[i])
		}
		msg.Attachments = mattach
	}

	// images
	if len(message.Images) > 0 {
		mattach = make([]mandrillAttachment, len(message.Images), len(message.Images))
		for i := range message.Images {
			mattach[i] = m.mailAttachmentToMandrillAttachment(message.Images[i])
		}
		msg.Images = mattach
	}

	return msg, nil
}

// setMessageFrom set the "from" data and Reply-To header for the supplied mandrillMessage
func (m *mandrill) setMessageFrom(src *MailMessage, dest *mandrillMessage) {

	if len(dest.FromEmail) > 0 {
		dest.FromEmail = src.From.Email
		dest.FromName = src.From.Name
	} else {
		dest.FromEmail = m.defaultSender.Email
		dest.FromName = m.defaultSender.Name
	}

	if len(src.ReplyTo) > 0 {
		dest.Headers = map[string]string{
			`Reply-To`: src.ReplyTo,
		}
	}
}

// setMessageContent set the html or text content for the supplied mandrillMessage
func (m *mandrill) setMessageContent(src *MailMessage, dest *mandrillMessage) error {

	html, text, err := m.buildMessageContent(src, src.TemplateVars)
	if err != nil {
		return err
	}
	dest.Html = html
	if len(text) > 0 {
		dest.Text = text
	} else {
		dest.AutoText = true
	}

	return nil
}

// setMessageRecipients set the recipients and recipient metadata for the supplied mandrillMessage
func (m *mandrill) setMessageRecipients(recipients []MailRecipient, src *MailMessage, dest *mandrillMessage) error {

	mrcpt := make([]mandrillRecipient, len(recipients), len(recipients))
	for i := range recipients {
		mrcpt[i] = m.mailRecipientToMandrillRecipient(recipients[i])
	}
	dest.To = mrcpt

	// recipient metadata
	mrcptMeta := make([]mandrillRecipientMetadata, 0, len(recipients))
	var meta mandrillRecipientMetadata
	for _, v := range recipients {
		if len(v.Metadata) > 0 {
			meta = mandrillRecipientMetadata{
				Rcpt:   v.Email,
				Values: v.Metadata,
			}
			mrcptMeta = append(mrcptMeta, meta)
		}
	}
	if len(mrcptMeta) > 0 {
		dest.RecipientMetadata = mrcptMeta
	}

	return nil
}

// mailRecipientToMandrillRecipient converts a generic MailRecipient to a mandrillRecipient
// see Note at top of mail.go
func (m *mandrill) mailRecipientToMandrillRecipient(r MailRecipient) mandrillRecipient {

	return mandrillRecipient{
		Email:         r.Email,
		Name:          r.Name,
		RecipientType: r.RecipientType,
	}
}

// mailRecipientToMandrillRecipientMetadata converts a generic MailRecipient to a mandrillRecipientMetadata
// see Note at top of mail.go
func (m *mandrill) mailRecipientToMandrillRecipientMetadata(r MailRecipient) mandrillRecipientMetadata {

	return mandrillRecipientMetadata{
		Rcpt:   r.Email,
		Values: r.Metadata,
	}
}

// mailAttachmentToMandrillAttachment converts a generic EmailAttachment to a mandrillAttachment
// see Note at top of mail.go
func (m *mandrill) mailAttachmentToMandrillAttachment(a EmailAttachment) mandrillAttachment {

	return mandrillAttachment{
		Name:          a.Name,
		MimeType:      a.MimeType,
		Base64Content: a.Base64Content,
	}
}

// buildParams builds the message paramaters for sending the email
func (m *mandrill) buildParams(params *SendParams, msg *mandrillMessage) *mandrillParams {

	p := &mandrillParams{
		Key:     m.key,
		Message: msg,
		Async:   params.SendAsync,
		IpPool:  params.IpPool,
	}

	if !(params.SendAt == nil || params.SendAt.IsZero()) {

		_, offset := params.SendAt.Zone()

		// adjust for timezone
		if offset != 0 {
			seconds := params.SendAt.Unix()
			seconds = seconds - int64(offset)

			ut := time.Unix(seconds, 0)
			params.SendAt = &ut
		}

		p.SendAtTxt = params.SendAt.Format(`2006-01-02 15:04:05`)
	}

	return p
}

// send submits the email to Mandrill
func (m *mandrill) send(params *mandrillParams) ([]MailRecipientResponse, error) {

	//spew.Dump(params)
	sendJson, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}

	url := MANDRILL_BASE_URL + MANDRILL_MESSAGE_PATH
	reader := strings.NewReader(string(sendJson))

	response, err := m.client.Post(url, `application/json`, reader)
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()
	decoder := json.NewDecoder(response.Body)

	var mandrillResponse = new(mandrillResponse)

	err = decoder.Decode(&mandrillResponse.response)
	spew.Dump(mandrillResponse)
	if err != nil {
		fmt.Printf("ERR: %s\n", err.Error())
		return nil, err
	}

	return m.handleApiSuccess(mandrillResponse)
}

// handleApiSuccess created a structured response for a successful mandrill call. Note that the
// length of the MailRecipientResponse slice will always be one for SimpleMail and TemplateMail
func (m *mandrill) handleApiSuccess(response *mandrillResponse) ([]MailRecipientResponse, error) {

	var (
		resp   = make([]MailRecipientResponse, len(response.response), len(response.response))
		status MailStatus
	)
	for i, v := range response.response {

		switch v.Status {
		case MAIL_MESSAGE_SENT:
			status = MAIL_MESSAGE_SENT
		case MAIL_MESSAGE_QUEUED:
			status = MAIL_MESSAGE_SENT
		case MAIL_MESSAGE_SCHEDULED:
			status = MAIL_MESSAGE_SENT
		case MAIL_MESSAGE_REJECTED:
			status = MAIL_MESSAGE_SENT
		case MAIL_MESSAGE_INVALID:
			status = MAIL_MESSAGE_SENT
		default:
			status = MAIL_MESSAGE_UNKNOWN
		}

		resp[i] = MailRecipientResponse{
			Id:     v.Id,
			Email:  v.Email,
			Status: status,
			Error:  v.RejectReason,
		}
	}

	return resp, nil
}

// handleApiError formats proper message for api call that failed
func (m *mandrill) handleApiError(response *http.Response, params *mandrillParams) error {

	var e mandrillErrorResponse
	decoder := json.NewDecoder(response.Body)
	err := decoder.Decode(&e)
	if err != nil {
		return err
	}

	return fmt.Errorf("Email, sent to %d recipients starting with %s, failed : %s\n",
		len(params.Message.To), params.Message.To[0].Email, e.Message)
}
