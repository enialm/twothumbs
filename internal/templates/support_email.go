// File: internal/templates/issue_email.go

// This file contains email templates.

package templates

import (
	"fmt"
)

// The contact support email template
func ContactSupportEmail(workspace string, msg string) (subject string, body string) {
	subject = fmt.Sprintf("A message from (work)space %s", workspace)
	body = fmt.Sprintf(`Greetings,

you've got mail:

<msg>
%s
</msg>

Regards,

The Two Thumbs backend`, msg)

	return subject, body
}

// The issue email template
func IssueEmail(workspace string, issue string) (subject string, body string) {
	subject = fmt.Sprintf("Issue report from workspace %s", workspace)
	body = fmt.Sprintf(`Greetings,

I carry a message:

<issue>
%s
</issue>

Regards,

The Two Thumbs backend`, issue)

	return subject, body
}
