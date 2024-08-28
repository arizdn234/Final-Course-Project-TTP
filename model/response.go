package model

type ErrorResponse struct {
	Error string `json:"error"`
}

type SuccessResponse struct {
	Message string `json:"message"`
}

func NewErrorResponse(msg string) ErrorResponse {
	return ErrorResponse{
		Error: msg,
	}
}

func NewSuccessResponse(msg string) SuccessResponse {
	return SuccessResponse{
		Message: msg,
	}
}
