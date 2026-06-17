package gmail

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"refer-mate/internal/domain"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const (
	ScopeGmailSend   = "https://www.googleapis.com/auth/gmail.send"
	ScopeUserEmail   = "https://www.googleapis.com/auth/userinfo.email"
	ScopeUserProfile = "https://www.googleapis.com/auth/userinfo.profile"
)

type Client struct {
	oauthCfg *oauth2.Config
}

func NewClient(clientID, clientSecret, redirectURL string) *Client {
	return &Client{
		oauthCfg: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  redirectURL,
			Scopes:       []string{ScopeGmailSend, ScopeUserEmail, ScopeUserProfile},
			Endpoint:     google.Endpoint,
		},
	}
}

func (c *Client) AuthCodeURL(state string) string {
	return c.oauthCfg.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.ApprovalForce)
}

type GoogleUserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
	VerifiedEmail bool   `json:"verified_email"`
}

type SendResult struct {
	MessageID string
	ThreadID  string
}

func (c *Client) Exchange(ctx context.Context, code string) (*oauth2.Token, error) {
	return c.oauthCfg.Exchange(ctx, code)
}

func (c *Client) GetUserInfo(ctx context.Context, token *oauth2.Token) (*GoogleUserInfo, error) {
	httpClient := c.oauthCfg.Client(ctx, token)
	resp, err := httpClient.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var info GoogleUserInfo
	if err := json.Unmarshal(body, &info); err != nil {
		return nil, err
	}
	return &info, nil
}

func (c *Client) TokenFromDomain(t *domain.OAuthToken) *oauth2.Token {
	return &oauth2.Token{
		AccessToken:  t.AccessToken,
		RefreshToken: t.RefreshToken,
		Expiry:       t.ExpiresAt,
		TokenType:    "Bearer",
	}
}

func (c *Client) SendEmail(ctx context.Context, oauthToken *oauth2.Token, from, to, subject, htmlBody string) (*SendResult, error) {
	httpClient := c.oauthCfg.Client(ctx, oauthToken)

	msg, err := buildMIMEMessage(from, to, subject, htmlBody)
	if err != nil {
		return nil, err
	}
	encoded := base64.URLEncoding.EncodeToString(msg)

	payload := map[string]string{"raw": encoded}
	jsonPayload, _ := json.Marshal(payload)

	req, err := http.NewRequestWithContext(ctx, "POST",
		"https://gmail.googleapis.com/gmail/v1/users/me/messages/send",
		bytes.NewReader(jsonPayload),
	)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("gmail API error %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		ID       string `json:"id"`
		ThreadID string `json:"threadId"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return &SendResult{MessageID: result.ID, ThreadID: result.ThreadID}, nil
}

func buildMIMEMessage(from, to, subject, htmlBody string) ([]byte, error) {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)

	header := make(textproto.MIMEHeader)
	header.Set("From", from)
	header.Set("To", to)
	header.Set("Subject", subject)
	header.Set("MIME-Version", "1.0")
	header.Set("Date", time.Now().Format(time.RFC1123Z))
	header.Set("Content-Type", fmt.Sprintf("multipart/alternative; boundary=%s", w.Boundary()))

	// Write headers manually
	var headerBuf bytes.Buffer
	for k, vals := range header {
		for _, v := range vals {
			headerBuf.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
		}
	}
	headerBuf.WriteString("\r\n")

	// HTML part
	partHeader := make(textproto.MIMEHeader)
	partHeader.Set("Content-Type", "text/html; charset=UTF-8")
	partHeader.Set("Content-Transfer-Encoding", "quoted-printable")
	pw, err := w.CreatePart(partHeader)
	if err != nil {
		return nil, err
	}
	pw.Write([]byte(htmlBody))
	w.Close()

	return append(headerBuf.Bytes(), buf.Bytes()...), nil
}
