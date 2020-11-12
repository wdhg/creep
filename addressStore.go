package main

import (
	"bytes"
	"io"
	"sync"
)

type addressStore struct {
	addresses map[string]bool
	count     int
	lock      sync.Mutex
}

// newAddressStore makes a new addressStore
func newAddressStore(queueSize int) *addressStore {
	return &addressStore{
		addresses: make(map[string]bool),
		count:     0,
		lock:      sync.Mutex{},
	}
}

// next searches `addresses` for an unvisited address
func (s *addressStore) next() (string, bool) {
	s.lock.Lock()
	defer s.lock.Unlock()
	for address, hasBeenScraped := range s.addresses {
		if !hasBeenScraped {
			s.addresses[address] = true
			return address, true
		}
	}
	return "", false
}

// add adds an address to the addressStore if it isnt already in it and
// increments `count`
func (s *addressStore) add(address string) {
	s.lock.Lock()
	defer s.lock.Unlock()
	if _, ok := s.addresses[address]; ok {
		return
	}
	s.addresses[address] = false
	s.count++
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
