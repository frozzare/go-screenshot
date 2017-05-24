package screenshot

import (
	"testing"
)

func TestScreenshot(t *testing.T) {
	s := NewScreenshot()

	file, err := s.Save("http://google.com")
	if err != nil {
		t.Fatal(err)
	}

	if len(file) == 0 {
		t.Fatal("Empty file")
	}
}
