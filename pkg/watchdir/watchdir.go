package watchdir

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/fioncat/wshare/pkg/log"
	"github.com/fioncat/wshare/pkg/osutil"
	"github.com/fsnotify/fsnotify"
	"github.com/sirupsen/logrus"
)

const eventDeplay = time.Second * 2

type Notify struct {
	C chan struct{}

	rootDir string

	watcher *fsnotify.Watcher

	logger *logrus.Entry

	recursive bool
}

func NewNotify(rootDir string, rec bool) (*Notify, error) {
	var err error
	if !filepath.IsAbs(rootDir) {
		rootDir, err = filepath.Abs(rootDir)
		if err != nil {
			return nil, fmt.Errorf("failed to abs path %s: %v", rootDir, err)
		}
	}
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	err = w.Add(rootDir)
	if err != nil {
		return nil, fmt.Errorf("failed to watch %s: %v", rootDir, err)
	}

	logger := log.Get().WithFields(logrus.Fields{
		"Type":      "DirWatcher",
		"RootDir":   rootDir,
		"Recursive": rec,
	})
	n := &Notify{
		C: make(chan struct{}, 500),

		rootDir:   rootDir,
		watcher:   w,
		logger:    logger,
		recursive: rec,
	}
	if rec {
		err = n.flushDir()
		if err != nil {
			return nil, err
		}
	}
	go n.listen()

	return n, nil
}

func (n *Notify) listen() {
	n.logger.Info("begin to watch")
	for {
		select {
		case <-n.watcher.Events:
			deplayTimer := time.NewTimer(eventDeplay)
		delayLoop:
			for {
				select {
				case <-n.watcher.Events:

				case <-deplayTimer.C:
					break delayLoop
				}
			}
			n.C <- struct{}{}
			if n.recursive {
				err := n.flushDir()
				if err != nil {
					n.logger.Errorf("failed to flush dir: %v", err)
				}
			}

		case err := <-n.watcher.Errors:
			n.logger.Errorf("failed to watch %s: %v", n.rootDir, err)

		}
	}
}

func (n *Notify) flushDir() error {
	subDirs, err := osutil.WalkDirs(n.rootDir)
	if err != nil {
		return err
	}
	subMap := make(map[string]struct{}, len(subDirs))
	for _, dir := range subDirs {
		subMap[dir] = struct{}{}
	}

	curWatchDirs := n.watcher.WatchList()
	curWatchMap := make(map[string]struct{}, len(curWatchDirs))
	var toDelete []string
	for _, dir := range curWatchDirs {
		if dir == n.rootDir {
			continue
		}
		curWatchMap[dir] = struct{}{}
		if _, ok := subMap[dir]; !ok {
			toDelete = append(toDelete, dir)
		}
	}

	var toAdd []string
	for _, dir := range subDirs {
		if _, ok := curWatchMap[dir]; !ok {
			toAdd = append(toAdd, dir)
		}
	}

	for _, dir := range toDelete {
		err = n.watcher.Remove(dir)
		if err != nil {
			return fmt.Errorf("failed to remove watching %s: %v", dir, err)
		}
	}
	for _, dir := range toAdd {
		err = n.watcher.Add(dir)
		if err != nil {
			return fmt.Errorf("failed to add watching %s: %v", dir, err)
		}
	}
	return nil
}
