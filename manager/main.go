package main

import (
	"flag"
	"github.com/streadway/amqp"
	"net/http"
	"strconv"
)

var (
	addr     = flag.String(`addr`, `:4567`, `listen address`)
	users    = flag.Int(`users`, 100, `set your max users value and manager create rooms <=> users such length`)
	rabbirmq = flag.String("rabbitmq", "amqp://guest:guest@localhost:5672/", "Host:Port to rabbitmq server")
)

func main() {
	flag.Parse()

	conn, err := amqp.Dial(*rabbirmq)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	if err = buildRelations(conn); err != nil {
		panic(err)
	}

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

func buildRelations(conn *amqp.Connection) error {
	ch, err := conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	for i := 1; i <= *users; i++ {
		id := strconv.Itoa(i)
		err = ch.ExchangeDeclare(`U`+id, `fanout`, true, false, false, false, nil)
		if err != nil {
			return err
		}

		err = ch.ExchangeDeclare(`R`+id, `fanout`, true, false, false, false, nil)
		if err != nil {
			return err
		}

		err = ch.ExchangeBind(`U`+id, ``, `R`+id, false, nil)
		if err != nil {
			return err
		}
	}
	return nil
}
