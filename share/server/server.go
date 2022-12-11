package server

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/fioncat/wshare/pkg/log"
	"github.com/gorilla/websocket"
)

type Distributor struct {
	mu sync.RWMutex

	clients map[string]chan []byte
}

func NewDistributor() *Distributor {
	return &Distributor{
		clients: make(map[string]chan []byte),
	}
}

func (d *Distributor) Register(name string) (string, <-chan []byte) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if _, ok := d.clients[name]; ok {
		var idx = 1
		for {
			name = fmt.Sprintf("%s%d", name, idx)
			if _, ok := d.clients[name]; !ok {
				break
			}
			idx++
		}
	}

	ch := make(chan []byte, 800)
	d.clients[name] = ch
	return name, ch
}

func (d *Distributor) Deregister(name string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	ch := d.clients[name]
	if ch == nil {
		return
	}
	delete(d.clients, name)
	close(ch)
}

func (d *Distributor) Notify(name string, data []byte) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	for target, notify := range d.clients {
		if target != name {
			notify <- data
		}
	}
}

var (
	upgrader    = websocket.Upgrader{}
	distributor = NewDistributor()
)

func handle(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Get().Errorf("failed to upgrade connection: %v", err)
		return
	}
	defer conn.Close()
	addr := conn.RemoteAddr().String()

	name := r.Header.Get("client-name")
	if name == "" {
		log.Get().Warnf("client does not have a name, use remote addr")
		name = addr
	}
	var notify <-chan []byte
	name, notify = distributor.Register(name)

	logger := log.Get().WithField("client", name)
	if name != addr {
		logger = logger.WithField("addr", addr)
	}
	logger.Info("new client connected to server")

	defer distributor.Deregister(name)

	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			mt, data, err := conn.ReadMessage()
			if err != nil {
				logger.Errorf("failed to read message from client: %v", err)
				return
			}
			if mt != websocket.BinaryMessage {
				continue
			}
			size := log.BytesSize(data)
			logger.Infof("recv %s data", size)
			distributor.Notify(name, data)
		}
	}()

	for {
		select {
		case <-done:
			logger.Info("connection closed")
			return

		case data := <-notify:
			if conn == nil {
				return
			}
			err := conn.WriteMessage(websocket.BinaryMessage, data)
			if err != nil {
				logger.Errorf("failed to write message: %v", err)
				continue
			}
			size := log.BytesSize(data)
			log.Get().Infof("send %s data", size)
		}
	}
}
