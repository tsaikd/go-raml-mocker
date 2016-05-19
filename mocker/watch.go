package mocker

import (
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/tsaikd/KDGoLib/errutil"
	"gopkg.in/fsnotify.v1"
)

func watch(dir string) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			errutil.Trace(err)
			return
		}
		err = filepath.Walk(dir, func(fPath string, info os.FileInfo, ferr error) error {
			if ferr != nil {
				return ferr
			}
			if info.IsDir() {
				logger.Debugf("start watching %q", fPath)
				return watcher.Add(fPath)
			}
			return nil
		})
		if err != nil {
			errutil.Trace(err)
			return
		}

		for {
			select {
			case <-c:
				logger.Debugln("shutting down disk watcher ... done")
				err = watcher.Close()
				errutil.Trace(err)
				os.Exit(0)
			case evt := <-watcher.Events:
				switch evt.Op {
				case fsnotify.Create, fsnotify.Write:
					logger.Debugln("reloading", evt.String())
					errutil.Trace(reload())
				}
			}
		}
	}()
}
