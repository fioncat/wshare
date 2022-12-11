package clipboard

import (
	"context"

	"github.com/fioncat/wshare/pkg/log"
	"github.com/fioncat/wshare/share"
	"github.com/gorilla/websocket"
	"golang.design/x/clipboard"
)

type Handler struct{}

func (h *Handler) Notify(ch chan *share.Packet) {
}

func (h *Handler) listen(notify chan *share.Packet) {
	ctx := context.Background()
	imageWatcher := clipboard.Watch(ctx, clipboard.FmtImage)
	textWatcher := clipboard.Watch(ctx, clipboard.FmtText)

	for {
		var mt int
		var data []byte
		var fmtStr string
		var cooldown *cooldownSet
		select {
		case data = <-imageWatcher:
			mt = websocket.BinaryMessage
			fmtStr = "image"
			cooldown = imageCooldown

		case data = <-textWatcher:
			mt = websocket.TextMessage
			fmtStr = "text"
			cooldown = textCooldown
		}
		size := log.BytesSize(data)
		if cooldown.Exists(data) {
			continue
		}
		log.Get().Info("clipboard: send %s %s to server", size, fmtStr)
		notify <- &share.Packet{

		}
	}
}
