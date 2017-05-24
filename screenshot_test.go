package screenshot

import (
	"io/ioutil"
	"os"
	"testing"
)

var (
	url = "https://google.com"
)

func TestScreenshot(t *testing.T) {
	s := NewScreenshot()

	file, err := s.Save(url)
	if err != nil {
		t.Error(err)
	}

	if len(file) == 0 {
		t.Error("Empty file")
	}

	if _, err := os.Stat(file); os.IsNotExist(err) {
		t.Error(err)
	}

	dat, err := ioutil.ReadFile(file)
	if err != nil {
		t.Error(err)
	}

	if len(dat) == 0 {
		t.Fatal("Empty file")
	}

	if s.ContentType() != "image/png" {
		t.Error("Content type is different from expected")
	}

	if s.Format() != "png" {
		t.Error("Format is different from expected")
	}
}
