package middleware

import (
	"strings"
	"testing"

	"github.com/Anthony-Bible/password-exchange/app/internal/domains/message/adapters/primary/api/models"
	"github.com/stretchr/testify/assert"
	"pgregory.net/rapid"
)

// Postgres rejects NUL bytes (22021) and invalid UTF-8, and rejects strings
// that exceed a column's varchar limit (22001), where MariaDB used to
// truncate silently. Binding decision (epic #562, open question 4): such
// input is rejected at the API layer, before it can reach the database.

func okRequest() *models.MessageSubmissionRequest {
	return &models.MessageSubmissionRequest{Content: "hello"}
}

// emailOfLen builds a syntactically valid address of exactly n characters.
func emailOfLen(n int) string {
	const local = "a@"
	remaining := n - len(local) - len(".com")
	var labels []string
	for remaining > 0 {
		l := 63
		if remaining < l+1 {
			l = remaining
		}
		labels = append(labels, strings.Repeat("b", l))
		remaining -= l + 1
	}
	return local + strings.Join(labels, ".") + ".com"
}

func TestValidateMessageSubmission_RejectsNULInContent(t *testing.T) {
	req := okRequest()
	req.Content = "before\x00after"
	assert.Contains(t, ValidateMessageSubmission(req), "content")
}

func TestValidateMessageSubmission_RejectsInvalidUTF8InContent(t *testing.T) {
	req := okRequest()
	req.Content = "before\xff\xfeafter"
	assert.Contains(t, ValidateMessageSubmission(req), "content")
}

func TestValidateMessageSubmission_RejectsNULAndInvalidUTF8InPassphrase(t *testing.T) {
	for name, pass := range map[string]string{"nul": "pa\x00ss", "invalid utf8": "pa\xc3\x28ss"} {
		t.Run(name, func(t *testing.T) {
			req := okRequest()
			req.Passphrase = pass
			assert.Contains(t, ValidateMessageSubmission(req), "passphrase")
		})
	}
}

func TestValidateMessageSubmission_RecipientEmailBoundaryIs255Characters(t *testing.T) {
	notify := func(email string) map[string]interface{} {
		qid := 0
		return ValidateMessageSubmission(&models.MessageSubmissionRequest{
			Content:          "hello",
			SendNotification: true,
			Sender:           &models.Sender{Name: "S", Email: "s@example.com"},
			Recipient:        &models.Recipient{Name: "R", Email: email},
			AntiSpamAnswer:   "blue",
			QuestionID:       &qid,
		})
	}
	assert.NotContains(t, notify(emailOfLen(255)), "recipient.email", "255 characters fits other_email varchar(255)")
	assert.Contains(t, notify(emailOfLen(256)), "recipient.email", "256 characters would overflow other_email")
}

// Property: any plaintext content containing a NUL byte is rejected on "content".
func TestValidateMessageSubmission_Property_AnyNULInContentIsRejected(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		pre := rapid.StringN(0, 200, -1).Draw(rt, "pre")
		post := rapid.StringN(0, 200, -1).Draw(rt, "post")
		req := okRequest()
		req.Content = pre + "\x00" + post
		assert.Contains(rt, ValidateMessageSubmission(req), "content")
	})
}

// Property: well-formed, NUL-free content within the limit is never rejected
// for storage-safety reasons (guards against over-eager validation).
func TestValidateMessageSubmission_Property_CleanContentWithinLimitIsAccepted(t *testing.T) {
	clean := rapid.StringOfN(rapid.Rune().Filter(func(r rune) bool { return r != 0 }), 1, 500, -1).
		Filter(func(s string) bool { return strings.TrimSpace(s) != "" })
	rapid.Check(t, func(rt *rapid.T) {
		req := okRequest()
		req.Content = clean.Draw(rt, "content")
		assert.NotContains(rt, ValidateMessageSubmission(req), "content")
	})
}
