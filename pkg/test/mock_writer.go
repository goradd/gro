package test

import "fmt"

type FailingWriter struct {
	Count int
}

func (w *FailingWriter) Write(p []byte) (n int, err error) {
	if w.Count == 0 {
		return 0, fmt.Errorf("failed write")
	}
	w.Count--
	return len(p), nil
}
