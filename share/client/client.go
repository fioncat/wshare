package client

import (
	"bytes"
	"encoding/gob"
	"net/http"
	"reflect"

	"github.com/fioncat/wshare/pkg/log"
	"github.com/fioncat/wshare/share"
	"github.com/gorilla/websocket"
)

const (
	retryDialMinPeriodSeconds = 3
	retryDialMaxPeriodSeconds = 20
)

type Client struct {
	header http.Header

	history *share.History

	url string
}

func (c *Client) Start() {
	handlers := share.ListHandlers()
	handlerNames := make([]string, 0, len(handlers))
	selectCases := make([]reflect.SelectCase, 0, len(handlers)+1)
	for name, handler := range handlers {
		ch := make(chan *share.Packet, 500)
		go handler.Notify(ch)
		handlerNames = append(handlerNames, name)
		selectCases = append(selectCases, reflect.SelectCase{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(ch),
		})
	}

reentry:
	conn := c.dial()

	done := make(chan struct{})
	selectCases = append(selectCases, reflect.SelectCase{
		Dir:  reflect.SelectRecv,
		Chan: reflect.ValueOf(done),
	})
	go func() {
		defer close(done)
		log.Get().Info("begin to recv message")
		for {
			mt, data, err := conn.ReadMessage()
			if err != nil {
				log.Get().Errorf("failed to recv message from server: %v", err)
				return
			}
			if mt != websocket.BinaryMessage {
				continue
			}

			var buff bytes.Buffer
			buff.Write(data)
			decoder := gob.NewDecoder(&buff)

			var pack share.Packet
			err = decoder.Decode(&pack)
			if err != nil {
				log.Get().Errorf("failed to decode packet from server: %v", err)
				continue
			}

			if pack.Type == "" {
				log.Get().Warn("recv an invalid packet without type, discarded it")
				continue
			}

			handler := share.GetHandler(pack.Type)
			if handler == nil {
				log.Get().Warnf("recv an invalid packet with an unknown type %q, discarded it", pack.Type)
				continue
			}

			entry := log.Get().WithField("handler", pack.Type)
			size := log.BytesSize(pack.Data)
			entry.Infof("recv %s data from server, meta: %s", size, string(pack.Metadata))
			ctx := &share.Context{
				Entry:   entry,
				History: c.history,
				Pack:    &pack,
			}
			err = handler.Recv(ctx)
			if err != nil {
				entry.Errorf("failed to handle packet: %v", err)
				continue
			}
		}
	}()

	for {
		chosen, value, _ := reflect.Select(selectCases)
		if chosen == len(selectCases)-1 {
			goto reentry
		}

		handler := handlerNames[chosen]
		pack := value.Interface().(*share.Packet)
		pack.Type = handler

		var buffer bytes.Buffer
		encoder := gob.NewEncoder(&buffer)
		err := encoder.Encode(pack)
		if err != nil {
			log.Get().Errorf("failed to encode packet: %v", err)
			continue
		}
		err = conn.WriteMessage(websocket.BinaryMessage, buffer.Bytes())
		if err != nil {
			log.Get().Errorf("failed to send data to server: %v", err)
			continue
		}
		size := log.BytesSize(pack.Data)
		log.Get().Infof("%s: send %s data to server, meta: %s", pack.Type, size, string(pack.Metadata))
	}
}

func (c *Client) dial() *websocket.Conn {
	retrySeconds := retryDialMinPeriodSeconds
	for {
		conn, _, err := websocket.DefaultDialer.Dial(c.url, c.header)
		if err != nil {
			log.Get().Errorf("failed to dial server: %v, we will retry in %d seconds", err, retrySeconds)
			// Increment retrySeconds, so that if the server is
			// disconnected for a long time, do not retry too much.
			// But retrySeconds won't be bigger than max threshold.
			if retrySeconds <= retryDialMaxPeriodSeconds {
				retrySeconds++
			}
			continue
		}
		log.Get().Info("connected to server")
		return conn
	}
}
