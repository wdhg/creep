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

func newAddressStore(queueSize int) *addressStore {
	return &addressStore{
		addresses: make(map[string]bool),
		lock:      sync.Mutex{},
		count:     0,
	}
}

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

func (s *addressStore) add(address string) {
	s.lock.Lock()
	defer s.lock.Unlock()
	if _, ok := s.addresses[address]; ok {
		return
	}
	s.addresses[address] = false
	s.count++
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
