package main

import (
	"bytes"
	"io"
	"sync"
)

type addressStore struct {
	addresses map[string]bool
	maxCount  int
	count     int
	lock      sync.Mutex
	c         chan string
}

// newAddressStore makes a new addressStore
func newAddressStore(maxCount int) *addressStore {
	return &addressStore{
		addresses: make(map[string]bool),
		maxCount:  maxCount,
		count:     0,
		lock:      sync.Mutex{},
		c:         make(chan string, maxCount),
	}
}

// next searches `addresses` for an unvisited address
func (s *addressStore) next() string {
	return <-s.c
}

// add adds an address to the addressStore if it isnt already in it and
// increments `count`
func (s *addressStore) add(address string) bool {
	s.lock.Lock()
	if s.count >= s.maxCount {
		s.lock.Unlock()
		return true
	}
	if _, ok := s.addresses[address]; ok {
		s.lock.Unlock()
		return false
	}
	s.addresses[address] = true
	s.count++
	s.lock.Unlock()
	s.c <- address
	return false
}

// dumpTo joins all addresses into one large string and writes it to the writer
func (s *addressStore) dumpTo(writer io.Writer) (int, error) {
	s.lock.Lock()
	b := bytes.Buffer{}
	for address := range s.addresses {
		b.WriteString(address)
		b.WriteString("\n")
	}
	s.lock.Unlock()
	return writer.Write(b.Bytes())
}
