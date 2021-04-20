package websocket

import (
	"encoding/binary"
	"time"

	"github.com/gorilla/websocket"
	"github.com/kklab-com/gone/http"
)

const (
	TextMessageType   MessageType = 1
	BinaryMessageType MessageType = 2
	CloseMessageType  MessageType = 8
	PingMessageType   MessageType = 9
	PongMessageType   MessageType = 10

	CloseNormalClosure           CloseCode = 1000
	CloseGoingAway               CloseCode = 1001
	CloseProtocolError           CloseCode = 1002
	CloseUnsupportedData         CloseCode = 1003
	CloseNoStatusReceived        CloseCode = 1005
	CloseAbnormalClosure         CloseCode = 1006
	CloseInvalidFramePayloadData CloseCode = 1007
	ClosePolicyViolation         CloseCode = 1008
	CloseMessageTooBig           CloseCode = 1009
	CloseMandatoryExtension      CloseCode = 1010
	CloseInternalServerErr       CloseCode = 1011
	CloseServiceRestart          CloseCode = 1012
	CloseTryAgainLater           CloseCode = 1013
	CloseTLSHandshake            CloseCode = 1015
)

type Message interface {
	Type() MessageType
	Encoded() []byte
	Deadline() *time.Time
}

type MessageType int

func (m MessageType) wsLibType() int {
	return int(m)
}

type CloseCode int

type DefaultMessage struct {
	MessageType MessageType `json:"message_type,omitempty"`
	Message     []byte      `json:"message,omitempty"`
	Dead        *time.Time  `json:"dead,omitempty"`
}

func (m *DefaultMessage) Type() MessageType {
	return m.MessageType
}

func (m *DefaultMessage) Encoded() []byte {
	return m.Message
}

func (m *DefaultMessage) Deadline() *time.Time {
	return m.Dead
}

func (m *DefaultMessage) StringMessage() string {
	return string(m.Message)
}

type CloseMessage struct {
	DefaultMessage
	CloseCode CloseCode `json:"close_code,omitempty"`
}

func (m *CloseMessage) Encoded() []byte {
	if m.CloseCode == CloseNoStatusReceived {
		return []byte{}
	}

	if m.Message == nil {
		m.Message = []byte{}
	}

	buf := make([]byte, 2+len(m.Message))
	binary.BigEndian.PutUint16(buf, uint16(m.CloseCode))
	copy(buf[2:], m.Message)
	return buf
}

type PingMessage struct {
	DefaultMessage
}

type PongMessage PingMessage

func _ParseMessage(messageType int, bs []byte) *DefaultMessage {
	switch messageType {
	case websocket.TextMessage:
		return &DefaultMessage{
			MessageType: TextMessageType,
			Message:     bs,
		}
	case websocket.BinaryMessage:
		return &DefaultMessage{
			MessageType: BinaryMessageType,
			Message:     bs,
		}
	}

	return nil
}

type HttpWebsocketPack struct {
	Request     *http.Request          `json:"req"`
	HandlerTask ServerHandlerTask      `json:"task"`
	Message     Message                `json:"message"`
	Params      map[string]interface{} `json:"params"`
}

func _HttpWebsocketPackCast(obj interface{}) *HttpWebsocketPack {
	if pkg, true := obj.(*HttpWebsocketPack); true {
		return pkg
	}

	return nil
}
