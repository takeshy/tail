package tail

import (
	"bufio"
	"log"
	"os"
	"time"
	"github.com/go-fsnotify/fsnotify"
)

type Tail struct {
	path	string
	watcher *fsnotify.Watcher
	file	*os.File
	mTime	time.Time
	c		chan string
}

func (t *Tail) read() {
	fi, err := t.file.Stat()
	if err != nil {
		log.Fatal(err)
	}
	t.mTime = fi.ModTime()
	s := bufio.NewScanner(t.file)
	for s.Scan() {
		t.c <- s.Text()
	}
	return
}

func (t *Tail) waitOpenFile() {
	for {
		if err := t.watcher.Add(t.path); err != nil {
			time.Sleep(time.Second)
		} else {
			file, err := os.Open(t.path)
			if err != nil {
				log.Fatal(err);
			}
			if t.file != nil {
				t.file.Close()
				t.file = file
			} else {
				file.Seek(0, os.SEEK_END)
				t.file = file
			}
			return
		}
	}
}

func Watch(path string) chan string {
	t := new(Tail)
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	c := make(chan string)
	t.watcher = watcher
	t.path = path
	t.waitOpenFile()
	t.c = c
	go func() {
		defer t.watcher.Close()
		for {
			select {
			case event := <-t.watcher.Events:
				switch {
				case event.Op&fsnotify.Write == fsnotify.Write:
					t.read()
				case event.Op&fsnotify.Remove == fsnotify.Remove:
					t.waitOpenFile()
					t.read()
				default:
					fi, err := t.file.Stat()
					if err != nil {
						log.Fatal(err)
					}
					mTime := fi.ModTime()
					if mTime.After(t.mTime) {
						t.waitOpenFile()
						t.read()
					}
				}
			case err := <-watcher.Errors:
				log.Println("error:", err)
			}
		}
	}()
	return c
}
