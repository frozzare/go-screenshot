# Screenshot [![Build Status](https://travis-ci.org/frozzare/go-screenshot.svg?branch=master)](https://travis-ci.org/frozzare/go-screenshot) [![GoDoc](https://godoc.org/github.com/frozzare/go-screenshot?status.svg)](https://godoc.org/github.com/frozzare/go-screenshot) [![Go Report Card](https://goreportcard.com/badge/github.com/frozzare/go-screenshot)](https://goreportcard.com/report/github.com/frozzare/go-screenshot)

Go package for capturing screenshots of websites in various resolutions. It uses [phantomjs](http://phantomjs.org/) in the background.

## Installation

First you will need to install phantomjs then you can run `go get`

```
go get github.com/frozzare/go-screenshot
```

## Example

```go
package main

import (
	"fmt"
	"log"

	"github.com/frozzare/go-screenshot"
)

func main() {
	s := screenshot.NewScreenshot(&screenshot.Options{
		URL: "http://google.com",
	})

	file, err := s.Save() // or s.Save("http://google.com")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(file)
}
```

## License

MIT © [Fredrik Forsmo](https://github.com/frozzare)