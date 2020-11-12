package main

import (
	"bytes"
	"io"
	"sync"
)

type visited bool

const (
	beenVisited = true
	notVisited  = false
)

type addressStore struct {
	addresses map[string]visited
	maxCount  int
	count     int
	lock      sync.Mutex
}

// newAddressStore makes a new addressStore
func newAddressStore(maxCount int) *addressStore {
	return &addressStore{
		addresses: make(map[string]visited),
		maxCount:  maxCount,
		count:     0,
		lock:      sync.Mutex{},
	}
}

// next searches `addresses` for an unvisited address
func (s *addressStore) next() (string, bool) {
	s.lock.Lock()
	defer s.lock.Unlock()
	for address, hasBeenVisited := range s.addresses {
		if !hasBeenVisited {
			s.addresses[address] = beenVisited
			return address, true
		}
	}
	return "", false
}

// add adds an address to the addressStore if it isnt already in it and
// increments `count`
func (s *addressStore) add(address string) bool {
	s.lock.Lock()
	defer s.lock.Unlock()
	if s.count >= s.maxCount {
		return true
	}
	if _, ok := s.addresses[address]; ok {
		return false
	}
	s.addresses[address] = notVisited
	s.count++
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
