package main

import (
	"flag"
	"github.com/streadway/amqp"
	"net/http"
	"strconv"
	"sync"
)

var (
	addr        = flag.String(`addr`, `:4567`, `listen address`)
	users       = flag.Int(`users`, 100, `set your max users value and manager create rooms <=> users such length`)
	concurrency = flag.Int(`concurrency`, 100, `parallel create exs`)
	rabbirmq    = flag.String("rabbitmq", "amqp://guest:guest@localhost:5672/", "Host:Port to rabbitmq server")
)

func main() {
	flag.Parse()

	conn, err := amqp.Dial(*rabbirmq)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	println(`start build exs`)
	if err = buildRelations(conn, *concurrency); err != nil {
		panic(err)
	}
	println(`end build exs`)

	http.HandleFunc(`/bind`, func(rw http.ResponseWriter, r *http.Request) {
		ch, err := conn.Channel()
		if err != nil {

		}
		defer ch.Close()

		from := r.FormValue(`from`)
		to := r.FormValue(`to`)
		if len(from) == 0 || len(to) == 0 {
			rw.WriteHeader(http.StatusBadRequest)
		}

		err = ch.ExchangeBind(from, ``, to, false, nil)
		if err != nil {
			println(`bind:ERR:` + err.Error())
			rw.WriteHeader(http.StatusInternalServerError)
		}

		rw.WriteHeader(http.StatusNoContent)
	})

	http.HandleFunc(`/unbind`, func(rw http.ResponseWriter, r *http.Request) {
		ch, err := conn.Channel()
		if err != nil {

		}
		defer ch.Close()

		from := r.FormValue(`from`)
		to := r.FormValue(`to`)
		if len(from) == 0 || len(to) == 0 {
			rw.WriteHeader(http.StatusBadRequest)
		}

		err = ch.ExchangeUnbind(from, ``, to, false, nil)
		if err != nil {
			println(`unbind:ERR:` + err.Error())
			rw.WriteHeader(http.StatusInternalServerError)
		}

		rw.WriteHeader(http.StatusNoContent)
	})

	panic(http.ListenAndServe(*addr, nil))
}

func buildRelations(conn *amqp.Connection, parallel int) error {
	pool := []*amqp.Channel{}
	for i := 0; i < parallel; i++ {
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
		if p == parallel {
			p = 0
		}
		ch := pool[p]
		p++
		lock.Unlock()
		return ch
	}

	for i := 1; i <= *users; i += parallel {
		wg := &sync.WaitGroup{}
		for c := 0; c < parallel; c++ {
			wg.Add(1)

			go func(j int) {
				defer wg.Done()
				id := strconv.Itoa(j)
				sec := strconv.Itoa(j + 1)
				ch := getChan()
				err := ch.ExchangeDeclare(`U`+id, `fanout`, true, false, false, false, nil)
				if err != nil {
					panic(err)
				}

				err = ch.ExchangeDeclare(`U`+sec, `fanout`, true, false, false, false, nil)
				if err != nil {
					panic(err)
				}

				err = ch.ExchangeDeclare(`R`+id, `fanout`, true, false, false, false, nil)
				if err != nil {
					panic(err)
				}

				err = ch.ExchangeBind(`U`+id, ``, `R`+id, false, nil)
				if err != nil {
					panic(err)
				}
				err = ch.ExchangeBind(`U`+sec, ``, `R`+id, false, nil)
				if err != nil {
					panic(err)
				}
			}(i + c)
		}
		wg.Wait()
	}
	return nil
}
