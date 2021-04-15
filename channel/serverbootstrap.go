package channel

import (
	"net"
	"reflect"

	"github.com/kklab-com/gone/concurrent"
)

type ServerBootstrap interface {
	Bootstrap
	ChildHandler(handler Handler) ServerBootstrap
	SetChildParams(key ParamKey, value interface{})
	ChildParams() *Params
	Bind(localAddr net.Addr) Future
}

type DefaultServerBootstrap struct {
	DefaultBootstrap
	childHandler Handler
	childParams  Params
}

func (d *DefaultServerBootstrap) ChildHandler(handler Handler) ServerBootstrap {
	d.childHandler = handler
	return d
}

func (d *DefaultServerBootstrap) SetChildParams(key ParamKey, value interface{}) {
	d.childParams.Store(key, value)
}

func (d *DefaultServerBootstrap) ChildParams() *Params {
	return &d.childParams
}

func (d *DefaultServerBootstrap) Bind(localAddr net.Addr) Future {
	var serverChannel = reflect.New(d.channelType).Interface().(ServerChannel)
	serverChannel.Init()
	if d.handler != nil {
		serverChannel.Pipeline().AddLast("ROOT", d.handler)
	}

	if d.childHandler != nil {
		serverChannel.setChildHandler(d.childHandler)
	}

	d.Params().Range(func(k ParamKey, v interface{}) bool {
		serverChannel.SetParam(k, v)
		return true
	})

	future := NewChannelFuture(serverChannel, func(f concurrent.Future) interface{} {
		serverChannel.Bind(localAddr)
		return serverChannel
	})

	return future
}

func NewServerBootstrap() ServerBootstrap {
	bootstrap := DefaultServerBootstrap{}
	return &bootstrap
}
