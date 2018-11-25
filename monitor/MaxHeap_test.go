package monitor

import (
	"testing"
)

var heap MaxHeap

func createHeap() {
	heap = NewMaxHeap()
}

func TestHeapInsert(t *testing.T) {
	createHeap()
	expectedOrder := []int{0, 4, 2, 3, 1, 5}
	for _, el := range insert {
		heap.insert(el)
	}
	for idx, value := range heap.values {
		if value.Value.Status != expectedOrder[idx] {
			t.Error("Order doesn't match expectation !")
		}
	}
}

func TestHeapDelete(t *testing.T) {
	createHeap()
	var target1 *QueueElement
	var target2 *QueueElement

	target1Status := 2
	target2Status := 5

	expectedOrder1 := []int{0, 4, 5, 3, 1}
	expectedOrder2 := []int{0, 4, 1, 3}

	for _, el := range insert {
		t.Log(*el.Value)
		if el.Value.Status == target1Status {
			target1 = el
		}
		if el.Value.Status == target2Status {
			target2 = el
		}
		heap.insert(el)
	}
	for _, value := range heap.values {
		t.Log("Actual:", value.Value.Status)
	}
	heap.delete(target1)
	for i, value := range heap.values {
		if value.Value.Status != expectedOrder1[i] {
			t.Error("Fail: actual", value.Value.Status, "expected", expectedOrder1[i])
		} else {
			t.Log("actual", value.Value.Status, "expected", expectedOrder1[i])
		}
	}

	heap.delete(target2)
	for i, value := range heap.values {
		if value.Value.Status != expectedOrder2[i] {
			t.Error("Fail: actual", value.Value.Status, "expected", expectedOrder2[i])
		} else {
			t.Log("actual", value.Value.Status, "expected", expectedOrder2[i])
		}
	}
}

var insert = []*QueueElement{
	&QueueElement{
		Value: &PingLog{
			ResponseTime: 2,
			Status:       1,
		},
	},
	&QueueElement{
		Value: &PingLog{
			ResponseTime: 7,
			Status:       2,
		},
	},
	&QueueElement{
		Value: &PingLog{
			ResponseTime: 3,
			Status:       3,
		},
	},
	&QueueElement{
		Value: &PingLog{
			ResponseTime: 24,
			Status:       4,
		},
	},
	&QueueElement{
		Value: &PingLog{
			ResponseTime: 3,
			Status:       5,
		},
	},
}
