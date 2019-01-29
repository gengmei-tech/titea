package main

// not ready for use

import (
	"flag"
	"fmt"
	"github.com/go-redis/redis"
	"strings"
	"sync"
	"time"
)

var (
	host       = flag.String("h", "127.0.0.1", "gm-kv server host")
	port       = flag.String("p", "5379", "gm-kv server port")
	clientCnt  = flag.Int("c", 50, "Number of parallel connections (default 50)")
	requestCnt = flag.Int("n", 2000, "per client send requests (default 2000)")
	loop       = flag.Bool("l", false, "Loop. Run the tests forever")
	tests      = flag.String("t", "set,get", "only run the comma separated list of tests")
)

var clients []*redis.Client

var setAlready = false

func newClient() *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", *host, *port),
		DB:   0,
	})
	err := client.Ping().Err()
	if err != nil {
		fmt.Printf("error:%s", err)
	}
	return client
}

func initClient() {
	for i := 0; i < *clientCnt; i++ {
		clients = append(clients, newClient())
	}
}
func main() {
	flag.Parse()

	fmt.Printf("the host is: %s, the port is: %s\n", *host, *port)

	initClient()

	tst := strings.Split(*tests, ",")
	if len(tst) > 0 {
		for _, cmd := range tst {
			cmd = strings.ToLower(strings.TrimSpace(cmd))
			switch cmd {
			case "set":
				set()
				break
			case "get":
				get()
				break
			}
		}
	}

	if *loop {
		for {
			get()
			time.Sleep(time.Duration(30000000000))
		}
	}
}

func set() {

	fmt.Println("[test set begin]")

	wg := sync.WaitGroup{}
	wg.Add(*clientCnt)

	start := time.Now()

	for i := 0; i < *clientCnt; i++ {
		go func(i int) {
			defer wg.Done()
			for j := 0; j < *requestCnt; j++ {
				key := fmt.Sprintf("key_%d_%d", i, j)
				value := fmt.Sprintf("val_%d_%d", i, j)
				clients[i].Set(key, value, 0)
			}
		}(i)
	}
	wg.Wait()

	totalTime := time.Since(start).Nanoseconds()

	totalReq := int64((*requestCnt) * (*clientCnt))

	avg := float64(totalTime) / float64(totalReq)

	qps := 1000000000 * totalReq / totalTime

	fmt.Printf("%s done, totalTime:%d (ns), clientCnt: %d, requestCnt: %d, avg:%.4f (ns), requestPerSecond:%d \n", "set", totalTime, *clientCnt, totalReq, avg, qps)

	fmt.Println("[test set end]")

	setAlready = true

}

func get() {

	if !setAlready {
		set()
	}

	fmt.Println("[test get begin]")

	wg := sync.WaitGroup{}
	wg.Add(*clientCnt)

	start := time.Now()

	for i := 0; i < *clientCnt; i++ {
		go func(i int) {
			defer wg.Done()
			for j := 0; j < *requestCnt; j++ {
				key := fmt.Sprintf("key_%d_%d", i, j)
				clients[i].Get(key)
			}
		}(i)
	}
	wg.Wait()

	totalTime := time.Since(start).Nanoseconds()

	totalReq := int64((*requestCnt) * (*clientCnt))

	avg := float64(totalTime) / float64(totalReq)

	qps := 1000000000 * totalReq / totalTime

	fmt.Printf("%s done, totalTime:%d (ns), clientCnt: %d, requestCnt: %d, avg:%.4f (ns), requestPerSecond:%d \n", "get", totalTime, *clientCnt, totalReq, avg, qps)

	fmt.Println("[test get end]")

}
