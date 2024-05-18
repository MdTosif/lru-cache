package lru

type Linkedlist struct {
	Value interface{}
	Next *Linkedlist
	Prev *Linkedlist
}

func