package ai

import (
	"fmt"
	"strings"
)

const maxAttachmentPromptChars = 50000

// AppendAttachmentsToPrompt appends attachment content as raw text to the prompt.
// This is the fallback for providers that don't support file attachments.
func AppendAttachmentsToPrompt(prompt string, attachments []Attachment) string {
	if len(attachments) == 0 {
		return prompt
	}

	var b strings.Builder
	b.WriteString(prompt)
	b.WriteString("\n\n[Attachments]\n")

	for i, att := range attachments {
		if att.Content == "" {
			continue
		}
		b.WriteString(fmt.Sprintf("\nAttachment %d: %s (%s, %d bytes, %s)\n", i+1, att.Name, att.MimeType, att.Size, att.Encoding))
		b.WriteString("---\n")

		content := att.Content
		if len(content) > maxAttachmentPromptChars {
			content = content[:maxAttachmentPromptChars] + "\n[TRUNCATED]"
		}
		b.WriteString(content)
		b.WriteString("\n---\n")
	}

	return b.String()
}
