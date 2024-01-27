package model

type Response struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

func ReturnSuccess(msg string, data any) Response {
	return Response{
		Code:    0,
		Message: msg,
		Data:    data,
	}
}

func ReturnError(msg any) Response {
	switch t := msg.(type) {
	case error:
		return Response{
			Code:    1,
			Message: t.Error(),
			Data:    nil,
		}
	case string:
		return Response{
			Code:    1,
			Message: t,
			Data:    nil,
		}
	default:
		return Response{
			Code:    1,
			Message: t.(string),
			Data:    nil,
		}
	}

}
