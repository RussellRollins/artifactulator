package artifactory

import (
	"errors"
	"fmt"
	"io"
	"net/http"
)

type Artifactory struct {
	endpoint string
	user     string
	token    string
}

type ArtifactoryOption func(*Artifactory) error

func NewArtifactory(endpoint string, user string, token string, options ...ArtifactoryOption) (*Artifactory, error) {
	afy := &Artifactory{
		endpoint: endpoint,
		user:     user,
		token:    token,
	}

	for _, o := range options {
		if err := o(afy); err != nil {
			return nil, err
		}
	}

	if afy.endpoint == "" {
		return nil, errors.New("artifactory endpoint cannot be empty")
	}

	if afy.user == "" {
		return nil, errors.New("artifactory user cannot be empty")
	}

	if afy.token == "" {
		return nil, errors.New("artifactory token cannot be empty")
	}
	return afy, nil
}

func (a *Artifactory) Upload(endpoint string, content io.Reader) (*http.Response, error) {
	client := &http.Client{}
	req, err := http.NewRequest("PUT", fmt.Sprintf("%s/%s", a.endpoint, endpoint), content)
	if err != nil {
		return nil, err
	}
	//fmt.Printf("%+v\n", req)
	req.SetBasicAuth(a.user, a.token)
	return client.Do(req)
}

func (a *Artifactory) Download(endpoint string) (*http.Response, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/%s", a.endpoint, endpoint), nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(a.user, a.token)
	return client.Do(req)
}
