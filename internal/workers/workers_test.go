// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Rasul Khiriev

package workers

import (
	"testing"
)

// mockWorker is a test implementation of the Worker interface
// that tracks how many times Run was called.
type mockWorker struct {
	runCount int
}

func (m *mockWorker) Run() {
	m.runCount++
}

func TestWorkers_Run_AllWorkersAreCalled(t *testing.T) {
	w1 := &mockWorker{}
	w2 := &mockWorker{}
	w3 := &mockWorker{}

	ws := &Workers{workers: []Worker{w1, w2, w3}}
	ws.Run()

	for i, w := range []*mockWorker{w1, w2, w3} {
		if w.runCount != 1 {
			t.Errorf("worker[%d]: expected runCount=1, got %d", i, w.runCount)
		}
	}
}

func TestWorkers_Run_Empty(t *testing.T) {
	ws := &Workers{workers: []Worker{}}

	// Should not panic on empty workers list
	ws.Run()
}

func TestWorkers_Run_Nil(t *testing.T) {
	ws := &Workers{}

	// Should not panic when workers field is nil
	ws.Run()
}

func TestWorkers_Run_Order(t *testing.T) {
	order := []int{}

	// orderWorker records its index into the shared order slice
	newOrderWorker := func(id int) Worker {
		return &orderWorker{id: id, order: &order}
	}

	ws := &Workers{workers: []Worker{
		newOrderWorker(1),
		newOrderWorker(2),
		newOrderWorker(3),
	}}
	ws.Run()

	expected := []int{1, 2, 3}
	for i, v := range expected {
		if order[i] != v {
			t.Errorf("expected order[%d]=%d, got %d", i, v, order[i])
		}
	}
}

func TestWorkers_Run_CalledOnce(t *testing.T) {
	w := &mockWorker{}
	ws := &Workers{workers: []Worker{w}}

	ws.Run()

	if w.runCount != 1 {
		t.Errorf("expected Run to be called exactly once, got %d", w.runCount)
	}
}

func TestWorkers_Run_MultipleRuns(t *testing.T) {
	w := &mockWorker{}
	ws := &Workers{workers: []Worker{w}}

	ws.Run()
	ws.Run()
	ws.Run()

	if w.runCount != 3 {
		t.Errorf("expected runCount=3 after 3 calls, got %d", w.runCount)
	}
}

// orderWorker is a helper that appends its ID to a shared slice on Run.
type orderWorker struct {
	id    int
	order *[]int
}

func (o *orderWorker) Run() {
	*o.order = append(*o.order, o.id)
}
