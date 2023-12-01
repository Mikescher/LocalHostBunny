package bunny

import "strings"

func FilenameToMime(fn string, fallback string) string {
	lowerFN := strings.ToLower(fn)
	if strings.HasSuffix(lowerFN, ".html") || strings.HasSuffix(lowerFN, ".htm") {
		return "text/html"
	}
	if strings.HasSuffix(lowerFN, ".css") {
		return "text/css"
	}
	if strings.HasSuffix(lowerFN, ".js") {
		return "text/javascript"
	}
	if strings.HasSuffix(lowerFN, ".json") {
		return "application/json"
	}
	if strings.HasSuffix(lowerFN, ".jpeg") || strings.HasSuffix(lowerFN, ".jpg") {
		return "image/jpeg"
	}
	if strings.HasSuffix(lowerFN, ".png") {
		return "image/png"
	}
	if strings.HasSuffix(lowerFN, ".svg") {
		return "image/svg+xml"
	}
	if strings.HasSuffix(lowerFN, ".gif") {
		return "image/gif"
	}
	if strings.HasSuffix(lowerFN, ".webp") {
		return "audio/webm"
	}
	if strings.HasSuffix(lowerFN, ".bmp") {
		return "image/bmp"
	}

	return fallback
}
