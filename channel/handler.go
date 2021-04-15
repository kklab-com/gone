package channel

import (
	"net"
)

type Handler interface {
	Added(ctx HandlerContext)
	Removed(ctx HandlerContext)
	Active(ctx HandlerContext)
	Inactive(ctx HandlerContext)
	Read(ctx HandlerContext, obj interface{})
	ReadCompleted(ctx HandlerContext)
	Write(ctx HandlerContext, obj interface{})
	Bind(ctx HandlerContext, localAddr net.Addr)
	Close(ctx HandlerContext)
	Connect(ctx HandlerContext, remoteAddr net.Addr)
	Disconnect(ctx HandlerContext)
	ErrorCaught(ctx HandlerContext, err error)
}

type DefaultHandler struct {
}

func NewDefaultHandler() *DefaultHandler {
	return new(DefaultHandler)
}

func (h *DefaultHandler) Active(ctx HandlerContext) {
	ctx.FireActive()
}

func (h *DefaultHandler) Inactive(ctx HandlerContext) {
	ctx.FireInactive()
}

func (h *DefaultHandler) Added(ctx HandlerContext) {
}

func (h *DefaultHandler) Removed(ctx HandlerContext) {
}

func (h *DefaultHandler) Read(ctx HandlerContext, obj interface{}) {
	(ctx).FireRead(obj)
}

func (h *DefaultHandler) ReadCompleted(ctx HandlerContext) {
	(ctx).FireReadCompleted()
}

func (h *DefaultHandler) Write(ctx HandlerContext, obj interface{}) {
	(ctx).FireWrite(obj)
}

func (h *DefaultHandler) Bind(ctx HandlerContext, localAddr net.Addr) {
	ctx.Bind(localAddr)
}

func (h *DefaultHandler) Close(ctx HandlerContext) {
	ctx.Close()
}

func (h *DefaultHandler) Connect(ctx HandlerContext, remoteAddr net.Addr) {
	ctx.Connect(remoteAddr)
}

func (h *DefaultHandler) Disconnect(ctx HandlerContext) {
	ctx.Disconnect()
}

func (h *DefaultHandler) ErrorCaught(ctx HandlerContext, err error) {
	(ctx).FireErrorCaught(err)
}
