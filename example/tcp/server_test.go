package tcp

import (
	"net"
	"reflect"
	"testing"
	"time"

	"github.com/kklab-com/gone/channel"
	"github.com/kklab-com/gone/tcp"
	"github.com/kklab-com/goth-kklogger"
	"github.com/kklab-com/goth-kkutil/buf"
)

func TestServer_Start(t *testing.T) {
	kklogger.SetLogLevel("TRACE")
	go func() {
		time.Sleep(time.Millisecond * 500)

		bootstrap := channel.NewBootstrap()
		bootstrap.ChannelType(reflect.TypeOf(tcp.DefaultTCPClientChannel{}))
		bootstrap.Handler(channel.NewInitializer(func(ch channel.Channel) {
			ch.Pipeline().AddLast("DECODE_HANDLER", NewDecodeHandler())
			ch.Pipeline().AddLast("HANDLER", &ClientHandler{})
		}))

		ch := bootstrap.Connect(&net.TCPAddr{IP: nil, Port: 18080}).Sync().Channel().(channel.ClientChannel)
		nch := bootstrap.Connect(&net.TCPAddr{IP: nil, Port: 18080}).Sync().Channel().(channel.ClientChannel)
		ch.Write(buf.NewByteBuf([]byte("o12b32c49")))
		time.Sleep(time.Second)
		ch.Write(buf.NewByteBuf([]byte("a42d22e41")))
		time.Sleep(time.Second)
		ch.Disconnect()
		nch.Write(buf.NewByteBuf([]byte("ccc")))
	}()

	(&Server{}).Start(&net.TCPAddr{IP: nil, Port: 18080})
}
