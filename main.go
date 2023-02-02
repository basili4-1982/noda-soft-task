package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// ЗАДАНИЕ:
// * сделать из плохого кода хороший;
// * важно сохранить логику появления ошибочных тасков;
// * сделать правильную мультипоточность обработки заданий.
// Обновленный код отправить через merge-request.

// приложение эмулирует получение и обработку тасков, пытается и получать и обрабатывать в многопоточном режиме
// В конце должно выводить успешные таски и ошибки выполнены остальных тасков

// A Ttype represents a meaninglessness of our life
type Ttype struct {
	id         int
	cT         string // время создания
	fT         string // время выполнения
	taskRESULT []byte
	success    bool
}

func main() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	superChan := make(chan Ttype, 10)
	defer close(superChan)

	go taskCreator(superChan)

	doneTasks := make(chan Ttype)
	defer close(doneTasks)
	undoneTasks := make(chan error)
	defer close(undoneTasks)

	go func() {
		// получение тасков
		for t := range superChan {
			t = taskWorker(t)
			go taskSorter(t, doneTasks, undoneTasks)
		}
	}()

loop:
	for {
		select {
		case r := <-doneTasks:
			fmt.Printf("Done: %d \n", r.id)
		case r := <-undoneTasks:
			fmt.Printf("Undone: %s \n", r)
		case <-sigChan:
			fmt.Println("Bye-byes")
			break loop
		}

	}
}

func taskCreator(a chan Ttype) {
	go func() {
		for {
			cT := time.Now().Format(time.RFC3339)
			if time.Now().Nanosecond()%2 > 0 { // вот такое условие появления ошибочных тасков
				cT = "Some error occured"
			}
			a <- Ttype{cT: cT, id: int(time.Now().Unix())} // передаем таск на выполнение
		}
	}()
}

func taskWorker(a Ttype) Ttype {
	tt, _ := time.Parse(time.RFC3339, a.cT)
	if tt.After(time.Now().Add(-20 * time.Second)) {
		a.taskRESULT = []byte("task has been successed")
		a.success = true
	} else {
		a.taskRESULT = []byte("something went wrong")
		a.success = false
	}
	a.fT = time.Now().Format(time.RFC3339Nano)

	time.Sleep(time.Millisecond * 150)

	return a
}

func taskSorter(t Ttype, doneTasks chan Ttype, undoneTasks chan error) {
	if t.success {
		doneTasks <- t
	} else {
		undoneTasks <- fmt.Errorf("task id %d time %s, error %s", t.id, t.cT, t.taskRESULT)
	}
}
