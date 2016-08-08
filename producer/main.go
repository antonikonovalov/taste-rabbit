package main

import (
	"flag"
	"github.com/streadway/amqp"
	"strconv"
	"sync"
	"time"
)

var (
	start       = flag.Int(`start`, 10, `start room pushers`)
	end         = flag.Int(`end`, 50, `end from room listeners`)
	concurrency = flag.Int(`concurrency`, 20, `parallel publish messages`)
	duration    = flag.Duration(`duration`, 10*time.Second, `how long send message`)
	rabbirmq    = flag.String("rabbitmq", "amqp://guest:guest@localhost:5672/", "Host:Port to rabbitmq server")
)

func main() {
	flag.Parse()

	conn, err := amqp.Dial(*rabbirmq)
	if err != nil {
		panic(err)
	}

	defer conn.Close()

	pool := []*amqp.Channel{}
	for i := 0; i < *concurrency; i++ {
		ch, err := conn.Channel()
		if err != nil {
			panic(err)
		}

		pool = append(pool, ch)
	}

	p := 0
	lock := sync.Mutex{}
	getChan := func() *amqp.Channel {
		lock.Lock()
		if p == *concurrency {
			p = 0
		}
		ch := pool[p]
		p++
		lock.Unlock()
		return ch
	}

	body := `test messages`
	msg := amqp.Publishing{
		ContentType:     "text/plain",
		ContentEncoding: "",
		Body:            []byte(body),
		DeliveryMode:    1, // 1=non-persistent, 2=persistent
		Priority:        0, // 0-9
		// a bunch of application/implementation-specific fields
	}

	timer := time.NewTimer(*duration)
	println(`start`)
	for {
		for i := *start; i < *end; i += *concurrency {
			wg := sync.WaitGroup{}
			for c := 0; c <= *concurrency; c++ {
				wg.Add(1)
				go func(exchange string) {
					defer wg.Done()
					ch := getChan()
					err = ch.Publish(
						exchange, // publish to an exchange
						"",       // routing to 0 or more queues
						false,    // mandatory
						false,    // immediate
						msg,
					)
					if err != nil {
						panic(err)
					}
				}(`R` + strconv.Itoa(i))
			}
			wg.Wait()
		}

		select {
		case <-timer.C:
			println(`exit`)
			return
		default:
		}
	}
}
