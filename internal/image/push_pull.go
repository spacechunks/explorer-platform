package image

import (
	"crypto/tls"
	"fmt"
	"net/http"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	ociv1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
)

func Push(img ociv1.Image, imgRef, user, pass string) error {
	ref, err := name.ParseReference(imgRef)
	if err != nil {
		return fmt.Errorf("push: parse image ref: %w", err)
	}
	// TODO: view remote.DefaultTransport
	tp := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
		ForceAttemptHTTP2: true,
	}
	auth := auther{
		username: user,
		password: pass,
	}
	if err := remote.Write(ref, img, remote.WithAuth(auth), remote.WithTransport(tp)); err != nil {
		return fmt.Errorf("push image: %w", err)
	}
	return nil
}

func Pull(imgRef, user, pass string) (ociv1.Image, error) {
	ref, err := name.ParseReference(imgRef)
	if err != nil {
		return nil, fmt.Errorf("pull: parse image ref: %w", err)
	}
	// TODO: view remote.DefaultTransport
	tp := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
		ForceAttemptHTTP2: true,
	}
	auth := auther{
		username: user,
		password: pass,
	}
	img, err := remote.Image(ref, remote.WithAuth(auth), remote.WithTransport(tp))
	if err != nil {
		return nil, fmt.Errorf("pull image: %w", err)
	}
	return img, nil
}

// hack to avoid having to rely on keychain stuff
type auther struct {
	username string
	password string
}

func (a auther) Authorization() (*authn.AuthConfig, error) {
	return &authn.AuthConfig{
		Username: a.username,
		Password: a.password,
	}, nil
}
