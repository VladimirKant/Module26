package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type RingIntBuffer struct {
	array []int
	pos   int
	size  int
	m     sync.Mutex
}

func NewRingIntBuffer(size int) *RingIntBuffer {
	return &RingIntBuffer{make([]int, size), -1, size, sync.Mutex{}}
}

func (r *RingIntBuffer) Push(el int) {
	r.m.Lock()
	defer r.m.Unlock()
	if r.pos == r.size-1 {
		for i := 1; i <= r.size-1; i++ {
			r.array[i-1] = r.array[i]
		}
		r.array[r.pos] = el
	} else {
		r.pos++
		r.array[r.pos] = el
	}
}

func (r *RingIntBuffer) Get() []int {
	if r.pos <= 0 {
		return nil
	}
	r.m.Lock()
	defer r.m.Unlock()
	var output []int = r.array[:r.pos+1]
	r.pos = -1
	return output
}

func read(nextStage chan<- int, done chan bool) {
	scanner := bufio.NewScanner(os.Stdin)
	var data string
	for scanner.Scan() {
		data = scanner.Text()
		if strings.EqualFold(data, "exit") {
			fmt.Println("Программа завершила работу!")
			close(done)
			return
		}
		i, err := strconv.Atoi(data)
		if err != nil {
			fmt.Println("Программа обрабатывает только целые числа")
			continue
		}
		nextStage <- i
	}
}

func negativeFilterStageInt(previousStageChannel <-chan int, nextStageChannel chan<- int, done <-chan bool) {
	for {
		select {
		case data := <-previousStageChannel:
			if data > 0 {
				nextStageChannel <- data
			}
		case <-done:
			return
		}
	}
}

func notDividedThreeFunc(previousStageChannel <-chan int, nextStageChannel chan<- int, done <-chan bool) {
	for {
		select {
		case data := <-previousStageChannel:
			if data%3 == 0 {
				nextStageChannel <- data
			}
		case <-done:
			return
		}
	}
}

func bufferStageFunc(previousStageChannel <-chan int, nextStageChannel chan<- int, done <-chan bool, size int, interval time.Duration) {
	buffer := NewRingIntBuffer(size)
	for {
		select {
		case data := <-previousStageChannel:
			buffer.Push(data)
		case <-time.After(interval):
			bufferData := buffer.Get()
			if bufferData != nil {
				for _, data := range bufferData {
					nextStageChannel <- data
				}
			}
		case <-done:
			return
		}
	}
}

func main() {
	input := make(chan int)
	done := make(chan bool)
	go read(input, done)

	negativeFilterChannel := make(chan int)
	go negativeFilterStageInt(input, negativeFilterChannel, done)

	notDividedThreeChannel := make(chan int)
	go notDividedThreeFunc(negativeFilterChannel, notDividedThreeChannel, done)

	bufferedIntChannel := make(chan int)
	bufferSize := 10
	bufferDrainInterval := 30 * time.Second
	go bufferStageFunc(notDividedThreeChannel, bufferedIntChannel, done, bufferSize, bufferDrainInterval)

	for {
		select {
		case data := <-bufferedIntChannel:
			fmt.Println("Обработанные данные: ", data)
		case <-done:
			return
		}
	}
}
