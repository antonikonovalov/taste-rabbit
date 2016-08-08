# taste-rabbit


# base usage

```
# start manager
$ go run ./manager/main.go

# bind users and rooms
$ for i in {20..80}; do ./manager/bind.sh U$i R78; done

## for example add second user to room for 100 users
$ for i in {1..99}; do ./manager/bind.sh U$i R$((i+1)); done
```