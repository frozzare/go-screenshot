package screenshot

import (
	"net/http"
	"strconv"
)

func toInt(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}

	return i
}

func toBool(s string) bool {
	b, err := strconv.ParseBool(s)
	if err != nil {
		return false
	}

	return b
}

// Handler will render the screenshot of the given url direct to the browser.
func Handler(w http.ResponseWriter, r *http.Request) {
	s := NewScreenshot(&Options{
		Clip:            toBool(r.URL.Query().Get("clip")),
		Format:          r.URL.Query().Get("format"),
		Height:          toInt(r.URL.Query().Get("height")),
		IgnoreSSLErrors: toBool(r.URL.Query().Get("ignoresslerror")),
		SSLProtocol:     r.URL.Query().Get("sslprotocol"),
		Timeout:         toInt(r.URL.Query().Get("timeout")),
		URL:             r.URL.Query().Get("url"),
		Width:           toInt(r.URL.Query().Get("width")),
	})

	bytes, err := s.Bytes()
	if err != nil {
		http.Error(w, err.Error(), 500)
	}

	w.Header().Set("Content-Type", s.ContentType())
	w.Header().Set("Content-Length", strconv.Itoa(len(bytes)))

	w.Write(bytes)
}
