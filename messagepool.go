package main

type MessagePool struct {
	messages []string
	shuffler *IndexShuffler
}

func NewMessagePool(messages []string) *MessagePool {
	var mp = MessagePool{messages: messages}
	mp.shuffler = NewIndexShuffler(len(messages))
	return &mp
}

func (mp *MessagePool) Next() string {
	return mp.messages[mp.shuffler.Next()]
}
