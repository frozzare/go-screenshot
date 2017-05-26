package screenshot

import (
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	gourl "net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// ErrUnableToLoad is the error that are return when it can't load the url.
var ErrUnableToLoad = errors.New("Unable to load")

// Phantomjs script.
var script = `var webpage = require('webpage');
var noop = function () {};
var url = '%s';
var width = %d;
var height = %d;
var timeout = %d;
var format = '%s';
var clip = '%s';
var page = webpage.create();
page.settings.resourceTimeout = timeout;
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

page.onResourceTimeout = function(e) {
	console.error('unable to load');
	phantom.exit();
};

page.open(url, function (status) {
	if (status !== 'success') {
		console.error('unable to load');
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
	CacheTime       time.Time // Time to cache the images (e.g time.Now().Add(1 * time.Hour)). Default is zero.
	PhantomjsBin    string    // Path to phantomjs binary. Default is "phantomjs".
	Clip            bool      // Clip rectangle. Default false.
	Format          string    // Image format. Can be jpg, gif or png. Default is "png".
	Height          int       // Image height. Default 768.
	IgnoreSSLErrors bool      // Ignore SSL errors. Default false.
	Dir             string    // Directory to save the image in.
	SSLProtocol     string    // SSLProtocol. Default "sslv3".
	Timeout         int       // Timeout. Default 5000 (5s).
	URL             string    // URL to save screenshot from.
	Width           int       // Image width. Default 1024.
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

	if o.Timeout == 0 {
		o.Timeout = 5000
	}

	return &Screenshot{o}
}

// Bytes will take a screenshot of a url and return it as bytes or a error.
func (s *Screenshot) Bytes(args ...string) ([]byte, error) {
	if bytes := s.cached(args...); len(bytes) > 0 {
		return bytes, nil
	}

	url := s.url(args...)

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

	// Prepare stdin script.
	stdin := fmt.Sprintf(script,
		url,
		s.opts.Width,
		s.opts.Height,
		s.opts.Timeout,
		s.Format(),
		fmt.Sprintf("%t", s.opts.Clip))

	// Prepare command.
	cmd := exec.Command(s.opts.PhantomjsBin, parts...)
	cmd.Stdin = strings.NewReader(stdin)
	cmd.Stdout = &outb
	cmd.Stderr = &errb

	// Execute command.
	if err := cmd.Start(); err != nil {
		return []byte{}, err
	}

	// Wait for phantomjs do be done or kill cmd process after 15s.
	timer := time.AfterFunc(time.Duration(s.opts.Timeout*2)*time.Millisecond, func() {
		cmd.Process.Kill()
	})
	err := cmd.Wait()
	timer.Stop()

	if err != nil {
		return []byte{}, ErrUnableToLoad
	}

	// Return error if any.
	if err := errb.String(); len(err) > 0 {
		return []byte{}, ErrUnableToLoad
	}

	// Bail if we can't load the url.
	if strings.Contains(outb.String(), "unable to load") {
		return []byte{}, ErrUnableToLoad
	}

	dat, err := base64.StdEncoding.DecodeString(outb.String())
	if err != nil {
		return []byte{}, err
	}

	return dat, nil
}

// cached will determine if the url is cached already.
func (s *Screenshot) cached(args ...string) []byte {
	path := s.path(args...)

	if path == "" {
		return []byte{}
	}

	stat, err := os.Stat(path)

	if os.IsNotExist(err) {
		return []byte{}
	}

	if stat != nil && !stat.ModTime().Before(s.opts.CacheTime) {
		os.Remove(path)
		return []byte{}
	}

	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return []byte{}
	}

	return bytes
}

// Format returns the image format.
func (s *Screenshot) Format() string {
	format := s.opts.Format
	format = strings.ToLower(format)

	if format == "jpeg" {
		format = "jpg"
	}

	if format != "jpg" && format != "png" {
		format = "png"
	}

	return format
}

// ContentType returns image content type.
func (s *Screenshot) ContentType() string {
	switch s.Format() {
	case "jpg":
		return "image/jpeg"
	default:
		return "image/png"
	}
}

// path will return the image name path.
func (s *Screenshot) path(args ...string) string {
	url := s.url(args...)

	u, err := gourl.Parse(url)
	if err != nil {
		return ""
	}

	file := fmt.Sprintf("%s-%x.%s", u.Host, md5.Sum([]byte(url)), s.Format())

	if len(s.opts.Dir) != 0 {
		return filepath.Join(s.opts.Dir, file)
	}

	path, err := os.Getwd()
	if err != nil {
		return ""
	}

	return filepath.Join(path, file)
}

// Save saves a image.
func (s *Screenshot) Save(args ...string) (string, error) {
	bytes, err := s.Bytes(args...)
	if err != nil {
		return "", err
	}

	path := s.path(args...)

	if path == "" {
		return path, errors.New("Cannot create filename")
	}

	if err := ioutil.WriteFile(path, bytes, 0644); err != nil {
		return "", err
	}

	return path, nil
}

// url will return the url that should be used.
func (s *Screenshot) url(args ...string) string {
	url := s.opts.URL

	if len(args) > 0 && args[0] != "" {
		url = args[0]
	}

	return url
}
