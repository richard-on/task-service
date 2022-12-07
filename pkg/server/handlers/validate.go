package handlers

import (
	"encoding/json"
	"github.com/gofiber/fiber/v2"
	"github.com/richard-on/auth-service/pkg/response"
	"github.com/valyala/fasthttp"
)

func Validate(ctx *fiber.Ctx) (response.ValidateSuccess, error) {
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)
	req.Header.SetMethod("POST")
	req.SetRequestURI("http://localhost:80/api/v1/validate")
	req.Header.SetCookie("accessToken", ctx.Cookies("accessToken"))
	req.Header.SetCookie("refreshToken", ctx.Cookies("refreshToken"))

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	err := fasthttp.Do(req, resp)
	if err != nil {
		return response.ValidateSuccess{}, err
	}
	if resp.StatusCode() != fasthttp.StatusOK {
		return response.ValidateSuccess{}, err
	}

	body := resp.Body()
	val := response.ValidateSuccess{}

	err = json.Unmarshal(body, &val)
	if err != nil {
		return response.ValidateSuccess{}, err
	}

	return val, nil
}
