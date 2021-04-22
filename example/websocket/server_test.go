package websocket

import (
	"net"
	"testing"
	"time"

	"github.com/kklab-com/gone/channel"
	"github.com/kklab-com/gone/http"
	"github.com/kklab-com/gone/websocket"
	"github.com/kklab-com/goth-kklogger"
	"github.com/kklab-com/goth-kkutil/sync"
)

func TestServer_Start(t *testing.T) {
	kklogger.SetLogLevel("TRACE")
	serverBootstrap := channel.NewServerBootstrap()
	serverBootstrap.
		ChannelType(&http.ServerChannel{}).
		SetParams(websocket.ParamCheckOrigin, false)

	server := serverBootstrap.
		ChildHandler(channel.NewInitializer(func(ch channel.Channel) {
			ch.Pipeline().AddLast("DISPATCHER", http.NewDispatchHandler(NewRoute())).
				AddLast("WS_UPGRADE", &websocket.WSUpgradeProcessor{})
		})).
		Bind(&net.TCPAddr{IP: nil, Port: 18081}).Sync().Channel()

	go func() {
		time.Sleep(time.Minute * 1)
		server.Close()
	}()

	bootstrap := channel.NewBootstrap()
	bootstrap.ChannelType(&websocket.Channel{})
	bootstrap.Handler(channel.NewInitializer(func(ch channel.Channel) {
		ch.Pipeline().AddLast("HANDLER", websocket.NewInvokeHandler(&ClientHandlerTask{}, nil))
	}))

	ch := bootstrap.Connect(nil, &websocket.WSCustomConnectConfig{Url: "ws://localhost:18081/echo", Header: nil}).Sync().Channel()
	ch.Write(&websocket.DefaultMessage{
		MessageType: websocket.TextMessageType,
		Message:     []byte("write data"),
	})

	bwg := sync.BurstWaitGroup{}
	for i := 0; i < 100; i++ {
		bwg.Add(1)
		go func(i int) {
			chs := bootstrap.Connect(nil, &websocket.WSCustomConnectConfig{Url: "ws://localhost:18081/echo", Header: nil}).Sync().Channel()
			time.Sleep(time.Millisecond * 500)
			if i%2 == 0 {
				chs.Disconnect()
			}

			bwg.Done()
		}(i)
	}

	bwg.Wait()
	ch.Write(&websocket.CloseMessage{
		DefaultMessage: websocket.DefaultMessage{
			MessageType: websocket.CloseMessageType,
			Message:     []byte("text"),
		},
		CloseCode: websocket.CloseNormalClosure,
	})

	time.Sleep(time.Millisecond * 500)
	server.CloseFuture().Sync()
}
