package fs

import (
	"time"

	"github.com/fsnotify/fsevents"
	"github.com/sclevine/cflocal/engine"
)

var changeEvents = []fsevents.EventFlags{
	fsevents.Mount, fsevents.Unmount,
	fsevents.ItemCreated, fsevents.ItemRemoved,
	fsevents.ItemRenamed, fsevents.ItemModified,
}

func (f *FS) Watch(dir string, wait time.Duration) (change <-chan map[string]string, done chan<- struct{}, err error) {
	dev, err := fsevents.DeviceForPath(dir)
	if err != nil {
		return nil, nil, err
	}
	stream := &fsevents.EventStream{
		Paths:   []string{dir},
		Latency: wait,
		Device:  dev,
		Flags:   fsevents.FileEvents,
	}
	stream.Start()
	source := stream.Events

	out := make(chan map[string]string)
	stop := make(chan struct{})
	go func() {
		for {
			select {
			case events := <-source:
				for _, e := range events {
					if hasFlags(e.Flags, changeEvents...) {
						out <- map[string]string{"time": time.Now().Format(time.RFC3339)}
						break
					}
				}
			case <-stop:
				stream.Stop()
				return
			}
		}
	}()

	return out, stop, nil
}

func hasFlags(flag fsevents.EventFlags, flags ...fsevents.EventFlags) bool {
	for _, f := range flags {
		if flag&f == f {
			return true
		}
	}
	return false
}
