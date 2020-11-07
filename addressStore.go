package main

import (
	"fmt"
	"os"
	"strings"
	"sync"
)

type addressStore struct {
	addresses map[string]bool
	count     int
	lock      sync.Mutex
	queue     chan string
}

func newAddressStore(queueSize int) *addressStore {
	return &addressStore{
		addresses: make(map[string]bool),
		lock:      sync.Mutex{},
		count:     0,
		queue:     make(chan string, queueSize),
	}
}

func (s *addressStore) next() string {
	return <-s.queue
}

func (s *addressStore) add(address string) {
	s.lock.Lock()
	if _, ok := s.addresses[address]; ok {
		s.lock.Unlock()
		return
	}
	s.addresses[address] = false
	s.count++
	s.lock.Unlock()
	if len(s.queue) < cap(s.queue) {
		go func() {
			s.queue <- address
		}()
	}
}

func (s *addressStore) dumpToString() string {
	s.lock.Lock()
	defer s.lock.Unlock()
	builder := strings.Builder{}
	for address := range s.addresses {
		builder.WriteString(address)
		builder.WriteString("\n")
	}
	return builder.String()
}

func (s *addressStore) dumpToTerminal() {
	fmt.Println(s.dumpToString())
}

func (s *addressStore) dumpToFile(filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	_, err = f.WriteString(s.dumpToString())
	if err != nil {
		return err
	}
	err = f.Close()
	if err != nil {
		return err
	}
	return nil
}
