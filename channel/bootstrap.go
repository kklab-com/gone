package channel

import (
	"net"
	"reflect"
)

type Bootstrap interface {
	Handler(handler Handler) Bootstrap
	ChannelType(ch Channel) Bootstrap
	Connect(localAddr net.Addr, remoteAddr net.Addr) Future
	SetParams(key ParamKey, value interface{})
	Params() *Params
}

type DefaultBootstrap struct {
	handler     Handler
	channelType reflect.Type
	params      Params
}

func (d *DefaultBootstrap) SetParams(key ParamKey, value interface{}) {
	d.params.Store(key, value)
}

func (d *DefaultBootstrap) Params() *Params {
	return &d.params
}

func NewBootstrap() Bootstrap {
	bootstrap := DefaultBootstrap{}
	return &bootstrap
}

func (d *DefaultBootstrap) Handler(handler Handler) Bootstrap {
	d.handler = handler
	return d
}

func (d *DefaultBootstrap) ChannelType(ch Channel) Bootstrap {
	d.channelType = reflect.ValueOf(ch).Elem().Type()
	return d
}

func (d *DefaultBootstrap) Connect(localAddr net.Addr, remoteAddr net.Addr) Future {
	channelType := reflect.New(d.channelType)
	var channel = channelType.Interface().(Channel)
	ValueSetFieldVal(&channelType, "pipeline", _NewDefaultPipeline(channel))
	channel.Init()
	d.Params().Range(func(k ParamKey, v interface{}) bool {
		channel.SetParam(k, v)
		return true
	})

	if d.handler != nil {
		channel.Pipeline().AddLast("ROOT", d.handler)
	}

	ValueSetFieldVal(&channelType, "closeFuture", channel.Pipeline().newFuture())
	channel.Pipeline().fireRegistered()
	return channel.Connect(localAddr, remoteAddr)
}

func ValueSetFieldVal(target *reflect.Value, field string, val interface{}) bool {
	if icc := target.Elem().FieldByName(field); icc.IsValid() && icc.CanSet() {
		icc.Set(reflect.ValueOf(val))
		return true
	} else {
		return false
	}
}
