package share

import (
	"fmt"
	"os"
	"sync"

	"github.com/sirupsen/logrus"
)

type Packet struct {
	// Type is the handler name to handle this packet.
	Type string

	// Metadata is the handler metadata for this packet. The style and
	// meaning of metadata are completely determined by the handler.
	// For example, Metadata can be a JSON type object.
	Metadata []byte

	// Data is the main packet data to share.
	// For example, Data can be a clipboard text.
	Data []byte
}

type History struct {
	File *os.File

	mu sync.Mutex
}

func (his *History) Write(name string, msg string, args ...any) {
}

type Context struct {
	*logrus.Entry

	History *History
	Pack    *Packet
}

type Handler interface {
	Notify(ch chan *Packet)
	Recv(ctx *Context) error
}

type HandlerBuilder func() (Handler, error)

var (
	handlers = map[string]Handler{}

	handlerBuilders = map[string]HandlerBuilder{}
)

func Init() error {
	for name, builder := range handlerBuilders {
		handler, err := builder()
		if err != nil {
			return fmt.Errorf("failed to init handler %q: %v", name, err)
		}
		handlers[name] = handler
	}
	return nil
}

func GetHandler(name string) Handler {
	return handlers[name]
}

func ListHandlers() []Handler {
	hs := make([]Handler, 0, len(handlers))
	for _, h := range handlers {
		hs = append(hs, h)
	}
	return hs
}
