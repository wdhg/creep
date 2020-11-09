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

// dumpToFile saves all found addresses to a file
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

// dumpToTerminal outputs all found addresses to the terminal
func (s *addressStore) dumpToTerminal() {
	fmt.Println(s.dumpToString())
}

// dumpToString joins all addresses into one large string
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
