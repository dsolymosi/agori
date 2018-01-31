package agori

import (
	"container/list"
)

type LRU struct {
	s int
	l *list.List
	m map[uint32]*list.Element
}

func NewLRU(size int) *LRU {
	lru := LRU{
		s: size,
		l: list.New(),
		m: make(map[uint32]*list.Element),
	}
	return &lru
}

func (l *LRU) IsFull() bool {
	return l.l.Len() >= l.s
}

func (l *LRU) Add(mip uint32) (bool, uint32) {
	//if contains, touch
	e := l.m[mip]
	if e != nil {
		l.l.MoveToFront(e)
		return false, 0
	}

	l.l.PushFront(mip)
	l.m[mip] = l.l.Front()

	if l.l.Len() > l.s {
		o := l.l.Back()
		l.l.Remove(o)
		delete(l.m, o.Value.(uint32))

		return true, o.Value.(uint32)
	}

	return false, 0
}

func (l *LRU) Delete(mip uint32) bool {
	e := l.m[mip]
	if e != nil {
		l.l.Remove(e)
		delete(l.m, e.Value.(uint32))
		return true
	}
	return false
}

//get nth element counting backwards
func (l *LRU) GetEnd(offset int) uint32 {
	e := l.l.Back()
	for i := 0; i < offset && e != nil; i++ {
		e = e.Prev()
	}
	if e == nil {
		return 0
	}
	return e.Value.(uint32)
}
