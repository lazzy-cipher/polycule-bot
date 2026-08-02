package main

import (
	"math/rand/v2"
)

type MessagePool struct {
	messages []string
	shuffler *IndexShuffler // currently unused
}

func NewMessagePool(messages []string) *MessagePool {
	var mp = MessagePool{messages: messages}
	mp.shuffler = NewIndexShuffler(len(messages))
	return &mp
}

func (mp *MessagePool) Next() string {
	// return mp.messages[mp.shuffler.Next()]
	return mp.messages[rand.IntN(len(mp.messages))]
}
