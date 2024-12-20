package runners

import "time"

type Task struct {
	Ticker time.Ticker
	Done   chan bool
}

func NewTask(seconds int) Task {
	var T Task

	T.Ticker = *time.NewTicker(time.Duration(seconds) * time.Second)
	T.Done = make(chan bool)

	return T
}
