package channel

import (
	"net"

	"github.com/kklab-com/goth-kklogger"
	kkpanic "github.com/kklab-com/goth-panic"
)

type HandlerContext interface {
	Name() string
	Channel() Channel
	FireRegistered() HandlerContext
	FireUnregistered() HandlerContext
	FireActive() HandlerContext
	FireInactive() HandlerContext
	FireRead(obj interface{}) HandlerContext
	FireReadCompleted() HandlerContext
	FireErrorCaught(err error) HandlerContext
	Write(obj interface{}, future Future) Future
	Bind(localAddr net.Addr, future Future) Future
	Close(future Future) Future
	Connect(localAddr net.Addr, remoteAddr net.Addr, future Future) Future
	Disconnect(future Future) Future
	Deregister(future Future) Future
	prev() HandlerContext
	setPrev(prev HandlerContext) HandlerContext
	next() HandlerContext
	setNext(prev HandlerContext) HandlerContext
	setChannel(channel Channel) HandlerContext
	deferErrorCaught()
	handler() Handler
	read() HandlerContext
}

type DefaultHandlerContext struct {
	name     string
	channel  Channel
	_handler Handler
	nextCtx  HandlerContext
	prevCtx  HandlerContext
}

func (d *DefaultHandlerContext) setPrev(prev HandlerContext) HandlerContext {
	d.prevCtx = prev
	return d
}

func (d *DefaultHandlerContext) next() HandlerContext {
	return d.nextCtx
}

func (d *DefaultHandlerContext) setNext(next HandlerContext) HandlerContext {
	d.nextCtx = next
	return d
}

func (d *DefaultHandlerContext) setChannel(channel Channel) HandlerContext {
	d.channel = channel
	return d
}

func (d *DefaultHandlerContext) handler() Handler {
	return d._handler
}

func NewHandlerContext() *DefaultHandlerContext {
	context := new(DefaultHandlerContext)
	return context
}

func (d *DefaultHandlerContext) Name() string {
	return d.name
}

func (d *DefaultHandlerContext) Channel() Channel {
	return d.channel
}

func (d *DefaultHandlerContext) FireRegistered() HandlerContext {
	if d.next() != nil {
		defer d.next().deferErrorCaught()
		d.next().handler().Registered(d.next())
	}

	return d
}

func (d *DefaultHandlerContext) FireUnregistered() HandlerContext {
	if d.next() != nil {
		defer d.next().deferErrorCaught()
		d.next().handler().Unregistered(d.next())
	}

	return d
}

func (d *DefaultHandlerContext) FireActive() HandlerContext {
	if d.next() != nil {
		defer d.next().deferErrorCaught()
		d.next().handler().Active(d.next())
	}

	return d
}

func (d *DefaultHandlerContext) FireInactive() HandlerContext {
	if d.next() != nil {
		defer d.next().deferErrorCaught()
		d.next().handler().Inactive(d.next())
	}

	return d
}

func (d *DefaultHandlerContext) FireRead(obj interface{}) HandlerContext {
	if d.next() != nil {
		defer d.next().deferErrorCaught()
		d.next().handler().Read(d.next(), obj)
	}

	return d
}

func (d *DefaultHandlerContext) FireReadCompleted() HandlerContext {
	if d.next() != nil {
		defer d.next().deferErrorCaught()
		d.next().handler().ReadCompleted(d.next())
	}

	return d
}

func (d *DefaultHandlerContext) FireErrorCaught(err error) HandlerContext {
	if d.prev() != nil {
		defer d.prev().deferErrorCaught()
		d.prev().handler().ErrorCaught(d.prev(), err)
	}

	return d
}

func (d *DefaultHandlerContext) Write(obj interface{}, future Future) Future {
	future = d.checkFuture(future)
	if d.prev() != nil {
		defer d.prev().deferErrorCaught()
		d.prev().handler().Write(d.prev(), obj, future)
	}

	return future
}

func (d *DefaultHandlerContext) Bind(localAddr net.Addr, future Future) Future {
	future = d.checkFuture(future)
	if d.prev() != nil {
		defer d.prev().deferErrorCaught()
		d.prev().handler().Bind(d.prev(), localAddr, future)
	}

	return future
}

func (d *DefaultHandlerContext) Close(future Future) Future {
	future = d.checkFuture(future)
	if d.prev() != nil {
		defer d.prev().deferErrorCaught()
		d.prev().handler().Close(d.prev(), future)
	}

	return future
}

func (d *DefaultHandlerContext) Connect(localAddr net.Addr, remoteAddr net.Addr, future Future) Future {
	future = d.checkFuture(future)
	if d.prev() != nil {
		defer d.prev().deferErrorCaught()
		d.prev().handler().Connect(d.prev(), localAddr, remoteAddr, future)
	}

	return future
}

func (d *DefaultHandlerContext) Disconnect(future Future) Future {
	future = d.checkFuture(future)
	if d.prev() != nil {
		defer d.prev().deferErrorCaught()
		d.prev().handler().Disconnect(d.prev(), future)
	}

	return future
}

func (d *DefaultHandlerContext) Deregister(future Future) Future {
	future = d.checkFuture(future)
	if d.next() != nil {
		defer d.prev().deferErrorCaught()
		d.next().handler().Deregister(d.next(), future)
	}

	return future
}

func (d *DefaultHandlerContext) prev() HandlerContext {
	return d.prevCtx
}

func (d *DefaultHandlerContext) read() HandlerContext {
	if d.prev() != nil {
		defer d.prev().deferErrorCaught()
		d.prev().handler().read(d.prev())
	}

	return d
}

func (d *DefaultHandlerContext) deferErrorCaught() {
	if v := recover(); v != nil {
		caught := kkpanic.Convert(v)
		kklogger.ErrorJ("HandlerContext.ErrorCaught", caught.Error())
		d.handler().ErrorCaught(d, caught)
	}
}

func (d *DefaultHandlerContext) checkFuture(future Future) Future {
	if future != nil {
		future = d.Channel().Pipeline().newFuture()
	}

	return future
}

type LogStruct struct {
	Action  string
	Handler string
}
