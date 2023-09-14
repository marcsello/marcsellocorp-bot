package utils

func BoolToEmoji(val bool) string {
	if val {
		return "✅"
	} else {
		return "❌"
	}
}
