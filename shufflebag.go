package main

import (
	"math/rand"
	"sync"
)

type IndexShuffler struct {
	mu      sync.Mutex
	deck    []int
	length  int
	deckIdx int
}

func NewIndexShuffler(length int) *IndexShuffler {
	var s = IndexShuffler{
		length:  length,
		deck:    make([]int, length),
		deckIdx: 0,
	}

	for i := 0; i < length; i++ {
		s.deck[i] = i
	}

	s.shuffle()

	return &s
}

func (s *IndexShuffler) Next() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.deckIdx >= len(s.deck) {
		s.deckIdx = 0
		s.shuffle()
	}

	var idx = s.deck[s.deckIdx]
	s.deckIdx++

	return idx
}

func (s *IndexShuffler) shuffle() {
	rand.Shuffle(len(s.deck), func(i, j int) {
		s.deck[i], s.deck[j] = s.deck[j], s.deck[i]
	})
}
