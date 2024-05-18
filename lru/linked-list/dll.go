package linkedlist

type Element[T interface{}] struct {
	Value T
	Next  *Element[T]
	Prev  *Element[T]
}

type DoublyLiknedList[T interface{}] struct {
	Head *Element[T]
	Tail *Element[T]
}

func NewDll[T interface{}]() *DoublyLiknedList[T] {
	return &DoublyLiknedList[T]{}
}

func (ll *DoublyLiknedList[T]) Remove(element *Element[T]) {

	prev := element.Prev
	next := element.Next
	if element == ll.Head {
		println("XX")
		ll.Head = ll.Head.Next
	}
	if element == ll.Tail {
		ll.Tail = ll.Tail.Prev
	}
	if prev != nil {
		prev.Next = next
	}
	if next != nil {
		next.Prev = prev
	}

}

func (ll *DoublyLiknedList[T]) PushFront(element *Element[T]) {
	if ll.Head != nil {
		ll.Head.Prev = element
		element.Next = ll.Head
	} else {
		ll.Tail = element
	}
	ll.Head = element

}

func (ll *DoublyLiknedList[T]) MoveToFront(element *Element[T]) {
	ll.Remove(element)
	ll.PushFront(element)

}
