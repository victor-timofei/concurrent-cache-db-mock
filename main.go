package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

var cache = map[int]Book{}
var rnd = rand.New(rand.NewSource(time.Now().UnixNano()))

func main() {
	wg := &sync.WaitGroup{}
	m := &sync.RWMutex{}
	cacheCh := make(chan Book)
	dbCh := make(chan Book)
	for i := 0; i < 10; i++ {
		id := rnd.Intn(10) + 1
		wg.Add(3)

		go func(id int, group *sync.WaitGroup, mutex *sync.RWMutex, ch chan<- Book) {
			if b, ok := queryCache(id, mutex); ok {
				ch <- b
			}
			wg.Done()
		}(id, wg, m, cacheCh)

		go func(id int, group *sync.WaitGroup, mutex *sync.RWMutex, ch chan<- Book) {
			if b, ok := queryDatabase(id, mutex); ok {
				m.Lock()
				cache[id] = b
				m.Unlock()
				ch <- b
			}
			wg.Done()
		}(id, wg, m, dbCh)

		go func(wg *sync.WaitGroup, cacheCh, dbCh <-chan Book) {
			select {
			case b := <-cacheCh:
				fmt.Println("from cache")
				fmt.Println(b)
				<-dbCh
			case b := <-dbCh:
				fmt.Println("from db")
				fmt.Println(b)
			}
			wg.Done()
		}(wg, cacheCh, dbCh)
		wg.Wait()
	}
}

func queryCache(id int, mutex *sync.RWMutex) (Book, bool) {
	mutex.RLock()
	b, ok := cache[id]
	mutex.RUnlock()
	return b, ok
}

func queryDatabase(id int, mutex *sync.RWMutex) (Book, bool) {
	time.Sleep(100 * time.Millisecond)
	for _, b := range books {
		if b.ID == id {
			return b, true
		}
	}
	return Book{}, false
}
