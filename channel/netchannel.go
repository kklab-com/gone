package channel

import (
	"errors"
	"io"
	"net"
	"os"
	"reflect"
	"time"

	"github.com/kklab-com/goth-kklogger"
	"github.com/kklab-com/goth-kkutil/buf"
	errors2 "github.com/pkg/errors"
)

type NetChannel interface {
	Channel
	Conn() Conn
	RemoteAddr() net.Addr
	setConn(conn net.Conn)
}

type NetChannelSetConn interface {
	SetConn(conn net.Conn)
}

type DefaultNetChannel struct {
	DefaultChannel
	conn         Conn
	BufferSize   int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

func (c *DefaultNetChannel) Init() Channel {
	c.BufferSize = GetParamIntDefault(c, ParamReadBufferSize, 1024)
	c.ReadTimeout = time.Duration(GetParamIntDefault(c, ParamReadTimeout, 1000)) * time.Millisecond
	c.WriteTimeout = time.Duration(GetParamIntDefault(c, ParamWriteTimeout, 100)) * time.Millisecond
	return c
}

func (c *DefaultNetChannel) Conn() Conn {
	return c.conn
}

func (c *DefaultNetChannel) RemoteAddr() net.Addr {
	if c.conn != nil {
		return c.conn.RemoteAddr()
	}

	return nil
}

func (c *DefaultNetChannel) LocalAddr() net.Addr {
	if c.localAddr == nil {
		if c.conn != nil {
			c.localAddr = c.conn.LocalAddr()
			return c.localAddr
		} else {
			return nil
		}
	}

	return c.localAddr
}

func (c *DefaultNetChannel) setConn(conn net.Conn) {
	c.conn = WrapConn(conn)
}

func (c *DefaultNetChannel) IsActive() bool {
	return c.active
}

func (c *DefaultNetChannel) SetConn(conn net.Conn) {
	c.setConn(conn)
}

func (c *DefaultNetChannel) UnsafeWrite(obj interface{}) error {
	if c.Conn() == nil {
		return ErrNilObject
	}

	if !c.Conn().IsActive() {
		return net.ErrClosed
	}

	var bs []byte
	switch v := obj.(type) {
	case buf.ByteBuf:
		bs = v.Bytes()
	case []byte:
		bs = v
	default:
		kklogger.ErrorJ("DefaultNetChannel.UnsafeWrite", errors2.Wrap(ErrUnknownObjectType, reflect.TypeOf(v).String()))
		return ErrUnknownObjectType
	}

	if c.WriteTimeout > 0 {
		c.Conn().SetWriteDeadline(time.Now().Add(c.WriteTimeout))
	}

	if _, err := c.Conn().Write(bs); err != nil {
		kklogger.WarnJ("DefaultNetChannel.UnsafeWrite", err.Error())
		return err
	}

	return nil
}

func (c *DefaultNetChannel) UnsafeRead() (interface{}, error) {
	if c.Conn() == nil {
		return nil, ErrNilObject
	}

	if !c.IsActive() {
		return nil, net.ErrClosed
	}

	bs := make([]byte, c.BufferSize)
	if c.ReadTimeout > 0 {
		c.Conn().SetReadDeadline(time.Now().Add(c.ReadTimeout))
	}

	rc, err := c.Conn().Read(bs)
	if err != nil {
		if errors.Is(err, os.ErrDeadlineExceeded) {
			if c.Conn().IsActive() {
				return nil, ErrSkip
			} else {
				c.Deregister()
				return nil, ErrNotActive
			}
		}

		if c.IsActive() {
			if err != io.EOF {
				kklogger.TraceJ("DefaultNetChannel.UnsafeRead", err.Error())
			}

			if !c.Conn().IsActive() {
				c.Deregister()
				return nil, err
			}
		} else if err == io.EOF {
			return nil, err
		}
	} else if rc == 0 {
		return nil, ErrSkip
	}

	return buf.NewByteBuf(bs[:rc]), nil
}

func (c *DefaultNetChannel) UnsafeDisconnect() error {
	if c.Conn() != nil {
		if c.Conn().IsActive() {
			return c.Conn().Close()
		}

		return nil
	}

	return ErrNilObject
}

func (c *DefaultNetChannel) UnsafeConnect(localAddr net.Addr, remoteAddr net.Addr) error {
	if remoteAddr == nil {
		return ErrNilObject
	}

	if conn, err := net.Dial(remoteAddr.Network(), remoteAddr.String()); err != nil {
		return err
	} else {
		c.conn = WrapConn(conn)
	}

	return nil
}
