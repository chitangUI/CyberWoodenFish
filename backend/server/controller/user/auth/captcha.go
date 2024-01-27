package auth

import (
	"errors"
	"github.com/chitangUI/electronic-wooden-fish/config"
	"github.com/cloudwego/hertz/pkg/common/json"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
)

type CaptchaValidation struct {
	Success    bool     `json:"success"`
	ErrorCodes []string `json:"error-codes"`
}

func ValidateCaptcha(reCaptchaResponse, secretKey string) (bool, error) {
	request, err := http.PostForm(config.TurnstileApi, url.Values{
		"secret":   {secretKey},
		"response": {reCaptchaResponse},
	})

	if err != nil {
		return false, err
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Fatal("close io failed")
		}
	}(request.Body)
	body, err := io.ReadAll(request.Body)

	if err != nil {
		return false, err
	}

	var captchaValidation CaptchaValidation
	if err = json.Unmarshal(body, &captchaValidation); err != nil {
		return false, err
	}

	if captchaValidation.Success {
		return true, nil
	} else {
		return false, errors.New(strings.Join(captchaValidation.ErrorCodes, ","))
	}
}
