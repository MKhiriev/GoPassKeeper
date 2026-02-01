package workers

type Workers struct {
	workers []Worker
}

func (w *Workers) Run() {
	for _, worker := range w.workers {
		worker.Run()
	}
}
