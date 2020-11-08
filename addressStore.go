package main

import (
	"fmt"
	"os"
	"strings"
	"sync"
)

type page struct {
	address string
	visited bool
	linksTo []string
}

type addressStore struct {
	pages map[string]page
	count int
	lock  sync.RWMutex
}

// newAddressStore makes a new addressStore
func newAddressStore(queueSize int) *addressStore {
	return &addressStore{
		pages: make(map[string]page),
		lock:  sync.RWMutex{},
		count: 0,
	}
}

// next searches `addresses` for an unvisited address, flags it as visited, and
// returns it
func (s *addressStore) next() (string, bool) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	for _, p := range s.pages {
		if !p.visited {
			return p.address, true
		}
	}
	return "", false
}

// add adds an address to the addressStore if it isnt already in it and
// increments `count`
func (s *addressStore) add(address string, linkedFrom string) {
	s.lock.Lock()
	defer s.lock.Unlock()
	// add address to linkedFrom's linksTo
	if p, ok := s.pages[linkedFrom]; ok {
		p.linksTo = append(p.linksTo, address)
		p.visited = true
		s.pages[linkedFrom] = p
	} else {
		s.pages[linkedFrom] = page{
			address: linkedFrom,
			visited: true,
			linksTo: []string{address},
		}
	}
	// keep track of address
	if _, ok := s.pages[address]; !ok {
		s.pages[address] = page{
			address: address,
			visited: false,
			linksTo: []string{},
		}
		s.count++
	}
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
	s.lock.RLock()
	defer s.lock.RUnlock()
	builder := strings.Builder{}
	for _, p := range s.pages {
		builder.WriteString(p.address)
		builder.WriteString("\n")
		for _, address := range p.linksTo {
			builder.WriteString(address)
			builder.WriteString("\n")
		}
	}
	return builder.String()
}
