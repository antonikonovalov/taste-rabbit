package main

import (
	"flag"
	"github.com/streadway/amqp"
	"os"
	"os/signal"
	"strconv"
	"syscall"
)

var (
	start    = flag.Int(`start`, 10, `start from users listeners`)
	end      = flag.Int(`end`, 20, `end from users listeners`)
	rabbirmq = flag.String("rabbitmq", "amqp://guest:guest@localhost:5672/", "Host:Port to rabbitmq server")
)

func main() {
	flag.Parse()

	conn, err := amqp.Dial(*rabbirmq)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	sigs := make(chan os.Signal, 1)
	done := make(chan struct{})

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigs
		println(sig.String())
		close(done)
	}()

	for i := *start; i <= *end; i++ {
		go func(id string) {
			ch, err := conn.Channel()
			if err != nil {
				panic(err)
			}
			defer ch.Close()

			queue, err := ch.QueueDeclare(`QU`+id, false, true, false, false, nil)
			if err != nil {
				panic(err)
			}
			if err = ch.QueueBind(queue.Name, "", `U`+id, false, nil); err != nil {
				panic(err)
			}
			msg, err := ch.Consume(queue.Name, `consumer`, true, false, false, false, nil)
			if err != nil {
				panic(err)
			}
			for {
				select {
				case m, ok := <-msg:
					if ok {
						println(`U`+id, `:`, string(m.Body))
					}
				case <-done:
					println(`exit consumer`, id)
					return
				}
			}

		}(strconv.Itoa(i))
	}

	<-done
	println(`exit`)
}
