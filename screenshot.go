package screenshot

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	gourl "net/url"
	"os/exec"
	"path/filepath"
	"strings"
)

// Phantomjs script.
var script = `var webpage = require('webpage');
var args = require('system').args;
var noop = function () {};
var url = '%s';
var width = %d;
var height = %d;
var timeout = %d;
var format = '%s';
var clip = '%s';
var page = webpage.create();
page.viewportSize = {
	width: width,
	height: height
};
page.clipRect = {
	top: 0,
	left: 0,
	width: (clip === 'true' ? width : 0),
	height: (clip === 'true' ? height : 0)
};

page.onConsoleMessage =
page.onConfirm =
page.onPrompt =
page.onError = noop;

page.open(url, function (status) {
	if (status !== 'success') {
		console.error('Unable to load');
		phantom.exit();
	}
	window.setTimeout(function () {
		page.evaluate(function () {
			if (!document.body.style.background) {
				document.body.style.backgroundColor = 'white';
			}
		});
		console.log(page.renderBase64(format));
		phantom.exit();
	}, timeout);
});
`

// Screenshot represents a screenshot.
type Screenshot struct {
	opts *Options
}

// Options is the screenshot options.
type Options struct {
	PhantomjsBin    string // Path to phantomjs binary.
	Clip            bool   // Clip rectangle. Default false.
	Format          string // Image format. Can be jpg, gif or png.
	Height          int    // Image height. Default 768.
	IgnoreSSLErrors bool   // Ignore SSL errors. Default false.
	Dir             string // Directory to save the image in.
	SSLProtocol     string // SSLProtocol. Default sslv3.
	Timeout         int    // Timeout. Default 0.
	URL             string // URL to save screenshot from.
	Width           int    // Image width. Default 1024.
}

// NewScreenshot creates a new screenshot struct with default options.
func NewScreenshot(args ...*Options) *Screenshot {
	var o *Options

	if len(args) > 0 && args[0] != nil {
		o = args[0]
	}

	if o == nil {
		o = &Options{}
	}

	if o.Height == 0 {
		o.Height = 768
	}

	if len(o.Format) == 0 {
		o.Format = "png"
	}

	if len(o.SSLProtocol) == 0 {
		o.SSLProtocol = "sslv3"
	}

	if o.Width == 0 {
		o.Width = 1024
	}

	if len(o.PhantomjsBin) == 0 {
		o.PhantomjsBin = "phantomjs"
	}

	return &Screenshot{o}
}

// Save saves a image.
func (s *Screenshot) Save(args ...string) (string, error) {
	url := s.opts.URL

	if len(args) > 0 && args[0] != "" {
		url = args[0]
	}

	var outb, errb bytes.Buffer

	// Parse input to head and parts.
	parts := []string{"/dev/stdin"}

	// Add ignore ssl errors if true.
	if s.opts.IgnoreSSLErrors {
		parts = append(parts, "--ignore-ssl-errors=true")
	}

	// Add ssl protocol is specified.
	if len(s.opts.SSLProtocol) != 0 {
		parts = append(parts, "--ssl-protocol="+s.opts.SSLProtocol)
	}

	// Prepare script.
	script = fmt.Sprintf(script,
		url,
		s.opts.Width,
		s.opts.Height,
		s.opts.Timeout,
		s.opts.Format,
		fmt.Sprintf("%t", s.opts.Clip))

	// Prepare command.
	cmd := exec.Command(s.opts.PhantomjsBin, parts...)
	cmd.Stdin = strings.NewReader(script)
	cmd.Stdout = &outb
	cmd.Stderr = &errb

	// Execute command.
	if err := cmd.Start(); err != nil {
		return "", err
	}

	// Wait for phantomjs do be done.
	cmd.Wait()

	// Return error if any.
	if err := errb.String(); len(err) > 0 {
		return "", errors.New(err)
	}

	return s.saveImage(url, outb.String())
}

// saveImage decode base64 output and saves it to a image.
func (s *Screenshot) saveImage(url, out string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(out)
	if err != nil {
		return "", err
	}

	u, err := gourl.Parse(url)
	if err != nil {
		return "", err
	}

	file := fmt.Sprintf("%s.%s", u.Host, s.opts.Format)

	if len(s.opts.Dir) != 0 {
		file = filepath.Join(s.opts.Dir, file)
	}

	if err := ioutil.WriteFile(file, data, 0644); err != nil {
		return "", err
	}

	return file, nil
}
