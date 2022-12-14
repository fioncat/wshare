package share

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/fioncat/wshare/config"
	"github.com/fioncat/wshare/pkg/crypto"
	"github.com/fioncat/wshare/pkg/log"
	"github.com/fioncat/wshare/pkg/osutil"
	"github.com/sirupsen/logrus"
)

type Packet struct {
	Type string

	Metadata []byte

	Data []byte
}

func (p *Packet) Encode() ([]byte, error) {
	var buff bytes.Buffer
	encoder := gob.NewEncoder(&buff)
	err := encoder.Encode(p)
	if err != nil {
		return nil, err
	}

	data := buff.Bytes()
	return crypto.Encrypt(data), nil
}

func DecodePack(data []byte) (*Packet, error) {
	data, err := crypto.Decrypt(data)
	if err != nil {
		return nil, fmt.Errorf("decrypt data failed: %v", err)
	}

	var buff bytes.Buffer
	buff.Write(data)
	decoder := gob.NewDecoder(&buff)

	var p Packet
	err = decoder.Decode(&p)
	if err != nil {
		return nil, fmt.Errorf("failed to decode packet: %v", err)
	}

	return &p, nil
}

type History struct {
	target io.Writer

	mu sync.Mutex
}

func OpenHistory() (*History, error) {
	path, err := config.LocalFile("history")
	if err != nil {
		return nil, err
	}
	dst, err := osutil.OpenAppend(path)
	if err != nil {
		return nil, err
	}
	return &History{target: dst}, nil
}

func (his *History) Write(name string, msg string, args ...any) {
	his.mu.Lock()
	defer his.mu.Unlock()

	now := time.Now().Format("2006-01-02 15:04:05")
	header := fmt.Sprintf("===> %s [%s]", name, now)
	msg = fmt.Sprintf(msg, args...)

	var buff bytes.Buffer
	buff.WriteString(header + "\n")
	buff.WriteString(msg + "\n")

	_, err := his.target.Write(buff.Bytes())
	if err != nil {
		log.Get().Warnf("write history file error: %v", err)
	}
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

func RegisterHandler(name string, b HandlerBuilder) {
	handlerBuilders[name] = b
}

func InitHandlers() error {
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

func ListHandlers() map[string]Handler {
	return handlers
}
