package scc

import (
	"fmt"
	"sankofa/ow"
	"sync"
)

// a thread-safe stack of 64-bit integers
type Stack struct {
	slice []int64
	hash  map[int64]int64
	mutex sync.RWMutex
}

func NewStack() *Stack {
	stack := new(Stack)
	stack.mutex.Lock()
	stack.hash = make(map[int64]int64)
	stack.mutex.Unlock()
	return stack
}

func (stack *Stack) String() string {
	ret := fmt.Sprint("L=", len(stack.slice), ":")
	first := true
	for _, element := range stack.slice {
		if first {
			ret += fmt.Sprint(element)
			first = false
		} else {
			ret += fmt.Sprint(",", element)
		}
	}
	return ret
}

func (stack *Stack) Size() int {
	stack.mutex.Lock()
	size := len(stack.slice)
	stack.mutex.Unlock()
	return size

}

func (stack *Stack) First() int64 {
	stack.mutex.Lock()
	first := stack.slice[0]
	stack.mutex.Unlock()
	return first
}

func (stack *Stack) Last() int64 {
	stack.mutex.Lock()
	last := stack.slice[len(stack.slice)-1]
	stack.mutex.Unlock()
	return last

}

// stack.Member(element) ⇢ position, ok?
func (stack *Stack) Member(element int64) bool {
	stack.mutex.Lock()
	_, ok := stack.hash[element]
	stack.mutex.Unlock()
	return ok
}

// stack.Push(element) ⇢ position
func (stack *Stack) Push(element int64) int64 {
	stack.mutex.Lock()

	// sanity check
	_, ok := stack.hash[element]
	if ok {
		ow.Panic("stack allows no duplicates:", element)
	}

	stack.slice = append(stack.slice, element)
	last := int64(len(stack.slice) - 1)
	stack.hash[element] = last

	stack.mutex.Unlock()
	return last
}

// Stack.Pop() ⇢ element, ok?
func (stack *Stack) Pop() (int64, bool) {
	var pop int64
	stack.mutex.Lock()

	last := int64(len(stack.slice) - 1)

	if last >= 0 {
		pop = stack.slice[last]
		stack.slice = stack.slice[:last]

		// sanity check
		position, ok := stack.hash[pop]
		if !ok || position != last {
			ow.Panic("hash does not match:", last)
		}

		delete(stack.hash, pop)
	}
	stack.mutex.Unlock()
	return pop, last >= 0
}
