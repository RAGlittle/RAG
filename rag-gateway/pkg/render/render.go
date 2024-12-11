package render

func ClampString(s string, maxLength int, ellipsis string) string {
	if len(s) <= maxLength {
		return s
	}
	if maxLength < len(ellipsis) {
		// Not enough space for even the ellipsis, return a truncated ellipsis
		return ellipsis[:maxLength]
	}
	return s[:maxLength-len(ellipsis)] + ellipsis
}
