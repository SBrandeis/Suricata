package monitor

import (
	"math"
	"time"
)

type MaxHeap struct {
	values []*QueueElement
}

func NewMaxHeap() MaxHeap {
	heap := MaxHeap{
		values: make([]*QueueElement, 1),
	}
	heap.values[0] = &QueueElement{
		Value: &PingLog{
			ResponseTime: time.Duration(math.MaxInt64),
		},
	}
	return heap
}

func (m *MaxHeap) insert(el *QueueElement) {
	el.heapIndex = len(m.values)
	m.values = append(m.values, el)
	m.bubbleUp(len(m.values)-1, el)
}

func (m *MaxHeap) delete(e *QueueElement) {
	lastEl := m.values[len(m.values)-1]
	m.swap(e.heapIndex, lastEl.heapIndex)
	m.values = m.values[:len(m.values)-1]
	m.bubbleDown(lastEl.heapIndex, lastEl)
}

func (m *MaxHeap) getMax() *QueueElement {
	return m.values[1]
}

func (m *MaxHeap) bubbleUp(idx int, el *QueueElement) {
	parentIdx := idx / 2
	currentIdx := idx
	for currentIdx > 0 && m.values[currentIdx].gt(m.values[parentIdx]) {
		m.swap(currentIdx, parentIdx)
		currentIdx = parentIdx
		parentIdx = parentIdx / 2
	}
}

func (m *MaxHeap) bubbleDown(idx int, el *QueueElement) {
	smallest := idx
	leftChild := 2 * idx
	rightChild := 2*idx + 1
	if leftChild < len(m.values) && m.values[leftChild].gt(m.values[smallest]) {
		smallest = leftChild
	}
	if rightChild < len(m.values) && m.values[rightChild].gt(m.values[smallest]) {
		smallest = rightChild
	}
	if smallest != idx {
		m.swap(idx, smallest)
		m.bubbleDown(smallest, el)
	}

}

// Greater Than
func (e *QueueElement) gt(other *QueueElement) bool {
	return e.Value.ResponseTime > other.Value.ResponseTime
}

// Swap two elements in the heap
func (m *MaxHeap) swap(idx1 int, idx2 int) {
	v1 := m.values[idx1]
	v2 := m.values[idx2]
	v1.heapIndex, v2.heapIndex = v2.heapIndex, v1.heapIndex
	m.values[idx1], m.values[idx2] = m.values[idx2], m.values[idx1]
}
