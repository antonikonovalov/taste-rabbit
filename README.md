# taste-rabbit


# base usage

```
## start manager and generate 100000 exchanges {u1,u2 -> r1},{u2,u3 -> r2}
$ go run ./manager/main.go -users=50000 -concurrency=1000

## bind users and rooms
$ for i in {20..80}; do ./manager/bind.sh U$i R78; done

## for example add second user to room for 100 users
$ for i in {1..99}; do ./manager/bind.sh U$i R$((i+1)); done

## start listeners 5000 consumers
$ go run ./consumer/main.go -start=1 -end=5000

## start generate messages to 50000 rooms on 5 min's by 1000 chunk
$ go run ./producer/main.go -concurrency=1000 -start=1 -end=50000 -duration=5m

```