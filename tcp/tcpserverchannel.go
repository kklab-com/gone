package tcp

import (
	"fmt"
	"net"
	"os"
	"reflect"

	"github.com/kklab-com/gone/channel"
	"github.com/kklab-com/goth-kklogger"
)

type ServerChannel struct {
	channel.DefaultNetServerChannel
	listen net.Listener
	active bool
}

var ClientChannelType = reflect.TypeOf(Channel{})

func (c *ServerChannel) Init() channel.Channel {
	c.pipeline = channel._NewDefaultPipeline(c)
	c.unsafe.BindFunc = c.bind
	c.unsafe.CloseFunc = c.close
	c.unsafe.CloseLock.Lock()
	return c
}

func (c *ServerChannel) bind(localAddr net.Addr) error {
	if c.Name == "" {
		c.Name = fmt.Sprintf("TCPSERV_%s", localAddr.String())
	}

	if c.active {
		kklogger.Error("ServerChannel.bind", fmt.Sprintf("%s bind twice", c.Name))
		os.Exit(1)
	}

	if listen, err := net.Listen("tcp4", localAddr.String()); err != nil {
		kklogger.ErrorJ("ServerChannel.bind", fmt.Sprintf("bind fail, %s", err.Error()))
		return err
	} else {
		c.active = true
		c.listen = listen
	}

	go c.acceptLoop()
	return nil
}

func (c *ServerChannel) acceptLoop() {
	for c.active {
		if conn, err := c.listen.Accept(); err != nil {
			if !c.active {
				return
			}

			kklogger.ErrorJ("ServerChannel.acceptLoop", err.Error())
		} else {
			cc := c.DeriveNetChildChannel(ClientChannelType, conn)
			go cc.(*Channel).read()
		}
	}
}

func (c *ServerChannel) close() error {
	c.active = false
	c.listen.Close()
	c.unsafe.CloseLock.Unlock()
	return nil
}

func (c *ServerChannel) IsActive() bool {
	return c.active
}
