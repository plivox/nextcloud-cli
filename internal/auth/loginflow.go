package auth

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/plivox/nextcloud-cli/internal/api"
)

type LoginFlow struct {
	LoginURL string
	token    string
	endpoint string
	r        *resty.Client
}

func StartLoginFlow(server string) (*LoginFlow, error) {
	var init struct {
		Poll struct {
			Token    string `json:"token"`
			Endpoint string `json:"endpoint"`
		} `json:"poll"`
		Login string `json:"login"`
	}

	r := resty.New().SetHeader("User-Agent", api.UserAgent)

	resp, err := r.R().SetResult(&init).Post(server + "/index.php/login/v2")
	if err != nil {
		return nil, err
	}

	if resp.StatusCode() != http.StatusOK || init.Login == "" {
		return nil, fmt.Errorf(
			"starting the login flow on %s failed: %s",
			server,
			resp.Status(),
		)
	}

	return &LoginFlow{
		LoginURL: init.Login,
		token:    init.Poll.Token,
		endpoint: init.Poll.Endpoint,
		r:        r,
	}, nil
}

func (f *LoginFlow) Wait(timeout time.Duration) (Credentials, error) {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		var creds Credentials

		resp, err := f.r.R().
			SetFormData(map[string]string{"token": f.token}).
			SetResult(&creds).
			Post(f.endpoint)
		if err != nil {
			return Credentials{}, err
		}

		switch resp.StatusCode() {
		case http.StatusOK:
			return creds, nil
		case http.StatusNotFound:
			time.Sleep(2 * time.Second)
		default:
			return Credentials{}, fmt.Errorf(
				"polling the login flow failed: %s",
				resp.Status(),
			)
		}
	}

	return Credentials{}, errors.New("login timed out")
}
