package ui

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"golang.org/x/term"
)

var frames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

type Spinner struct {
	w     io.Writer
	mu    *sync.Mutex
	stop  chan struct{}
	done  chan struct{}
	label string
	start time.Time
}

func Start(label string, w io.Writer, mu *sync.Mutex) *Spinner {
	if w == nil {
		w = os.Stdout
	}
	if mu == nil {
		mu = &sync.Mutex{}
	}
	if !term.IsTerminal(int(os.Stdout.Fd())) {
		return nil
	}
	s := &Spinner{w: w, mu: mu, stop: make(chan struct{}), done: make(chan struct{}), label: label, start: time.Now()}
	go s.run()
	return s
}

func (s *Spinner) run() {
	defer close(s.done)
	ticker := time.NewTicker(120 * time.Millisecond)
	defer ticker.Stop()
	i := 0
	for {
		select {
		case <-s.stop:
			return
		case <-ticker.C:
			s.mu.Lock()
			fmt.Fprintf(s.w, "\r  %s %s… %v  ", frames[i%len(frames)], s.label, time.Since(s.start).Round(time.Second))
			s.mu.Unlock()
			i++
		}
	}
}

func (s *Spinner) Cancel() {
	if s == nil || s.stop == nil {
		return
	}
	select {
	case <-s.stop:
	default:
		close(s.stop)
	}
	<-s.done
}
