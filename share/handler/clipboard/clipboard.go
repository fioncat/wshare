package clipboard

import (
	"context"
	"errors"
	"fmt"

	"github.com/fioncat/wshare/config"
	"github.com/fioncat/wshare/pkg/log"
	"github.com/fioncat/wshare/share"
	"golang.design/x/clipboard"
)

type Handler struct{}

func (h *Handler) Notify(ch chan *share.Packet) {
	ctx := context.Background()
	imageWatcher := clipboard.Watch(ctx, clipboard.FmtImage)
	textWatcher := clipboard.Watch(ctx, clipboard.FmtText)

	for {
		var data []byte
		var fmtStr string
		var cooldown *cooldownSet
		select {
		case data = <-imageWatcher:
			fmtStr = "image"
			cooldown = imageCooldown

		case data = <-textWatcher:
			fmtStr = "text"
			cooldown = textCooldown
		}
		if cooldown.Exists(data) {
			continue
		}
		ch <- &share.Packet{
			Metadata: []byte(fmtStr),
			Data:     data,
		}
	}
}

func (h *Handler) Recv(ctx *share.Context) error {
	pack := ctx.Pack
	if len(pack.Data) == 0 {
		return errors.New("received empty data")
	}
	fmtStr := string(pack.Metadata)
	size := log.BytesSize(pack.Data)
	var dataFmt clipboard.Format
	var cooldown *cooldownSet
	switch fmtStr {
	case "image":
		dataFmt = clipboard.FmtImage
		cooldown = imageCooldown
		ctx.History.Write("clipboard-image", "%s size of image", size)

	case "text":
		dataFmt = clipboard.FmtText
		cooldown = textCooldown
		ctx.History.Write("clipboard-text", string(pack.Data))

	default:
		return fmt.Errorf("unknown clipboard fmt %s", fmtStr)
	}
	cooldown.Set(pack.Data)

	if !config.Get().Clipboard.Readonly {
		clipboard.Write(dataFmt, pack.Data)
		ctx.Infof("write %s %s data to clipboard", size, fmtStr)
	}

	return nil
}
