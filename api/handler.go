package api

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"proxytrack/store"
	"proxytrack/types"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

type ProxyHandler struct {
	SStore     store.SessionStore
	targetURL  string
	reqTimeout time.Duration
}

func NewProxyHandler(store store.SessionStore, t time.Duration, url string) *ProxyHandler {
	return &ProxyHandler{
		SStore:     store,
		targetURL:  url,
		reqTimeout: t,
	}
}

func (h *ProxyHandler) Proxy(c *fiber.Ctx) error {
	start := time.Now()

	var params types.RequestParams
	if err := c.BodyParser(&params); err != nil {
		return ErrBadRequest()
	}
	// if errors := validate(&params); len(errors) > 0 {
	// 	return NewValidationError(errors)
	// }

	session, err := types.NewSessionFromParams(params, c.Body())
	if err != nil {
		return err
	}

	url, err := url.Parse(h.targetURL + c.OriginalURL())
	if err != nil {
		return ErrInternalServerError()
	}

	body := bytes.NewReader(c.Body())

	req, err := http.NewRequest(c.Method(), url.String(), body)
	if err != nil {
		return ErrInternalServerError()
	}
	c.Request().Header.VisitAll(func(key, value []byte) {
		req.Header.Add(string(key), string(value))
	})

	client := &http.Client{
		Timeout: h.reqTimeout,
	}

	resp, err := client.Do(req)
	if err != nil {
		respObj := map[string]string{"error": "Timeout"}
		respBytes, _ := json.Marshal(respObj)
		session.Response = respBytes
		session.Error = err.Error()
		session.ResponseTime = time.Now()
		session.Status = fiber.StatusBadGateway
		session.DurationMs = int64(time.Since(start))

		// Save session even on timeout
		if insertErr := h.SStore.InsertSession(c.Context(), session); insertErr != nil {
			fmt.Println("failed to insert session:", insertErr)
			//return ErrInternalServerError()
		}
		return ErrBadGateway(err)
	}
	defer resp.Body.Close()

	// Copy headers from upstream to client
	for name, values := range resp.Header {
		for _, v := range values {
			c.Set(name, v)
		}
	}
	c.Status(resp.StatusCode)

	var reader io.Reader = resp.Body
	if strings.EqualFold(resp.Header.Get("Content-Encoding"), "gzip") {
		gz, err := gzip.NewReader(resp.Body)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to decompress gzip"})
		}
		defer gz.Close()
		reader = gz

		c.Response().Header.Del("Content-Encoding")
	}

	var buf bytes.Buffer
	mw := io.MultiWriter(&buf, c)

	if _, err := io.Copy(mw, reader); err != nil {
		return err
	}

	session.Response = buf.Bytes()
	session.ResponseTime = time.Now()
	session.Status = resp.StatusCode
	session.DurationMs = int64(time.Since(start))

	err = h.SStore.InsertSession(c.Context(), session)
	if err != nil {
		fmt.Println("failed to insert session:", err)
		//return ErrInternalServerError()
	}
	return nil
}

func validate(params *types.RequestParams) map[string]string {
	validate := validator.New()
	if err := validate.Struct(params); err != nil {
		errs := err.(validator.ValidationErrors)
		errors := make(map[string]string)
		for _, e := range errs {
			errors[e.Field()] = fmt.Sprintf("failed on '%s' tag", e.Tag())
		}
		//Err := NewValidationError(errors)
		// _ = c.Status(fiber.StatusUnprocessableEntity).JSON(Err)
		return errors
	}
	return nil
}
