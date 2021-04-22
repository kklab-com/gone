package http

import (
	"bytes"
	"time"

	"github.com/kklab-com/gone-httpstatus"
	"github.com/kklab-com/gone/channel"
	"github.com/kklab-com/gone/http"
)

type DefaultTask struct {
	http.DefaultHTTPTask
}

func (l *DefaultTask) Get(ctx channel.HandlerContext, req *http.Request, resp *http.Response, params map[string]interface{}) http.ErrorResponse {
	resp.SetStatusCode(httpstatus.OK)
	resp.TextResponse(bytes.NewBufferString("feeling good"))
	return nil
}

type DefaultHomeTask struct {
	http.DefaultHTTPTask
}

func (l *DefaultHomeTask) Get(ctx channel.HandlerContext, req *http.Request, resp *http.Response, params map[string]interface{}) http.ErrorResponse {
	resp.SetStatusCode(httpstatus.OK)
	resp.TextResponse(bytes.NewBufferString(req.RequestURI))
	go func() {
		<-time.After(time.Millisecond * 100)
		ctx.Channel().Disconnect()
	}()

	return nil
}

type CloseTask struct {
	http.DefaultHTTPTask
}

func (l *CloseTask) Get(ctx channel.HandlerContext, req *http.Request, resp *http.Response, params map[string]interface{}) http.ErrorResponse {
	resp.SetStatusCode(httpstatus.OK)
	resp.TextResponse(bytes.NewBufferString(req.RequestURI))
	go func() {
		<-time.After(time.Second)
		ctx.Channel().Parent().Close()
	}()

	return nil
}
