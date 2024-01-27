package model

type RegisterRequest struct {
	Username          string `json:"username,required"`
	Password          string `json:"password,required" vd:"len($)>8 && len($)<14"`
	TelegramId        uint64 `json:"telegram_id,required"`
	ReCaptchaResponse string `json:"re_captcha_response"`
}

type LoginRequest struct {
	Username          string `json:"username,required"`
	Password          string `json:"password,required"`
	ReCaptchaResponse string `json:"re_captcha_response"`
}
