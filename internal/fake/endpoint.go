package fake

import (
	"errors"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

type Endpoint struct{}

func (e *Endpoint) Upload(endpoint string, content io.Reader) (*http.Response, error) {
	p := rand.Intn(20)
	// sleep that many seconds
	time.Sleep(time.Duration(p) * time.Second)

	_, err := ioutil.ReadAll(content)
	if err != nil {
		panic(err)
	}

	// with 5% probability, return an error
	// with 15% probability, return a 500
	// with 80% probability, return all good.
	switch true {
	case p < 1:
		return nil, errors.New("something awful has happened")
	case p < 4:
		return &http.Response{
			Status:     "500 Internal Server Error",
			StatusCode: 500,
			Body:       ioutil.NopCloser(strings.NewReader("")),
		}, nil
	default:
		return &http.Response{
			Status:     "200 OK",
			StatusCode: 200,
			Body:       ioutil.NopCloser(strings.NewReader("")),
		}, nil
	}
}

func (e *Endpoint) Download(endpoint string) (*http.Response, error) {
	p := rand.Intn(1000)
	// sleep that many milliseconds
	time.Sleep(time.Duration(p) * time.Millisecond)

	// with 5% probability, return an error
	// with 15% probability, return a 500
	// with 80% probability, return all good.
	switch true {
	case p < 50:
		return nil, errors.New("something awful has happened")
	case p < 200:
		return &http.Response{
			Status:     "500 Internal Server Error",
			StatusCode: 500,
			Body:       ioutil.NopCloser(strings.NewReader("")),
		}, nil
	default:
		return &http.Response{
			Status:     "200 OK",
			StatusCode: 200,
			Body:       ioutil.NopCloser(strings.NewReader("")),
		}, nil
	}
}
