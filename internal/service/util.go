package service

// TruncateContent shortens content to fit within maxLen characters
func TruncateContent(content string, maxLen int) string {
	if len(content) <= maxLen {
		return content
	}
	return content[:maxLen-3] + "..."
}
