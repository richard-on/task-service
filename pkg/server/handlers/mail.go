package handlers

import (
	"encoding/json"
	"github.com/gofiber/fiber/v2"
	"github.com/richard-on/auth-service/pkg/response"
	"github.com/richard-on/mail-service/pkg/server/request"
	"github.com/valyala/fasthttp"
)

func SendEmail(ctx *fiber.Ctx, mailReq request.SendMail) error {

	marshalled, _ := json.Marshal(mailReq)

	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)
	req.Header.SetMethod("POST")
	req.Header.SetContentType("application/json")
	req.SetRequestURI("http://localhost:3000/mail/v1/send")
	req.Header.SetCookie("accessToken", ctx.Cookies("accessToken"))
	req.Header.SetCookie("refreshToken", ctx.Cookies("refreshToken"))
	req.SetBody(marshalled)

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	err := fasthttp.Do(req, resp)
	if err != nil {
		return err
	}
	if resp.StatusCode() != fasthttp.StatusOK {
		return err
	}

	body := resp.Body()
	val := response.ValidateSuccess{}

	err = json.Unmarshal(body, &val)
	if err != nil {
		return err
	}

	return nil
}
