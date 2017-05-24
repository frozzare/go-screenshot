package screenshot

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestScreenshot(t *testing.T) {
	s := NewScreenshot()

	file, err := s.Save("https://google.com")
	if err != nil {
		t.Fatal(err)
	}

	if len(file) == 0 {
		t.Fatal("Empty file")
	}

	if _, err := os.Stat(file); os.IsNotExist(err) {
		t.Fatal(err)
	}

	dat, err := ioutil.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}

	if len(dat) == 0 {
		t.Fatal("Empty file")
	}
}
