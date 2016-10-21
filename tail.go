package tail

import (
	"bufio"
	"github.com/go-fsnotify/fsnotify"
	"log"
	"os"
	"path"
)

type Tail struct {
	path       string
	watcher    *fsnotify.Watcher
	file       *os.File
	beforeSize int64
	c          chan string
}

func (t *Tail) size() int64 {
	fileInfo, err := t.file.Stat()
	if err != nil {
		log.Fatal(err)
	}
	return fileInfo.Size()
}

func (t *Tail) read() {
	size := t.size()
	if size < t.beforeSize {
		t.file.Seek(0, os.SEEK_SET)
	}
	t.beforeSize = size
	s := bufio.NewScanner(t.file)
	for s.Scan() {
		t.c <- s.Text()
	}
	return
}

func (t *Tail) openFileAndSeekEnd() {
	file, err := os.Open(t.path)
	if err != nil {
		return
	}
	file.Seek(0, os.SEEK_END)
	t.file = file
	t.beforeSize = t.size()
}

func (t *Tail) openFile() {
	file, err := os.Open(t.path)
	if err != nil {
		log.Fatal(err)
	}
	if t.file != nil {
		t.file.Close()
	}
	t.file = file
	t.read()
	return
}

func Watch(filePath string) chan string {
	t := new(Tail)
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	c := make(chan string)
	t.watcher = watcher
	t.path = filePath
	if err = t.watcher.Add(path.Dir(t.path)); err != nil {
		log.Fatal(err)
	}
	t.openFileAndSeekEnd()
	t.c = c
	go func() {
		defer t.watcher.Close()
		for {
			select {
			case event := <-t.watcher.Events:
				switch {
				case event.Op&fsnotify.Write == fsnotify.Write:
					if path.Base(t.path) == path.Base(event.Name) {
						t.read()
					}
				case event.Op&fsnotify.Create == fsnotify.Create:
					if path.Base(t.path) == path.Base(event.Name) {
						t.openFile()
					}
				}
			case err := <-watcher.Errors:
				log.Println("error:", err)
			}
		}
	}()
	return c
}
