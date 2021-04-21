package channel

import (
	"context"
	"net"
	"reflect"
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
	serverChannelType := reflect.New(d.channelType)
	var serverChannel = serverChannelType.Interface().(ServerChannel)
	if preInit, ok := serverChannel.(BootstrapChannelPreInit); ok {
		preInit.BootstrapPreInit()
	}

	serverChannel.setPipeline(_NewDefaultPipeline(serverChannel))
	cancel, cancelFunc := context.WithCancel(context.Background())
	serverChannel.setContext(cancel)
	serverChannel.setContextCancelFunc(cancelFunc)
	d.Params().Range(func(k ParamKey, v interface{}) bool {
		serverChannel.SetParam(k, v)
		return true
	})

	serverChannel.Init()
	if d.handler != nil {
		serverChannel.Pipeline().AddLast("ROOT", d.handler)
	}

	if d.childHandler != nil {
		serverChannel.setChildHandler(d.childHandler)
	}

	serverChannel.setLocalAddr(localAddr)
	if postInit, ok := serverChannel.(BootstrapChannelPostInit); ok {
		postInit.BootstrapPostInit()
	}

	serverChannel.setCloseFuture(serverChannel.Pipeline().newFuture())
	serverChannel.Pipeline().fireRegistered()
	return serverChannel.Bind(localAddr)
}

func NewServerBootstrap() ServerBootstrap {
	bootstrap := DefaultServerBootstrap{}
	return &bootstrap
}
