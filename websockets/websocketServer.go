package websockets

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/KinMod-ui/thelaLocator/db"
	"github.com/KinMod-ui/thelaLocator/helper"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/websocket"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/lib/pq"
)

type webSocketHandler struct {
	upgrader websocket.Upgrader
	dbPool   *pgxpool.Conn
}

func makeRedisClient() *redis.Client {

	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // default DB
	})

	return rdb
}

func (wsh webSocketHandler) findFriends(fid string, rdb *redis.Client) []string {

	helper.Mylog.Println("reached here")
	var ctx = context.Background()

	retFriends := []string{}
	encFriends := []byte{}
	err := rdb.Get(ctx, string(fid)).Scan(&encFriends)

	if err == redis.Nil {
		helper.Mylog.Println("fid not found in redis, Get from db")

		friends, err := wsh.dbPool.Query(ctx, "SELECT f2 FROM friends WHERE f1=$1", fid)
		if err != nil {
			helper.Mylog.Println(err)
		} else {
			for friends.Next() {
				friend, err := friends.Values()
				if err != nil {
					helper.Mylog.Println("Error while iterating db ", err)
				}
				retFriends = append(retFriends, friend[0].(string))
			}

			encFriendList, err := json.Marshal(retFriends)
			if err != nil {
				helper.Mylog.Println("Error Marshalling data", err)
			}

			str, err := rdb.SetEX(ctx, string(fid), encFriendList, time.Hour).Result()
			if err != nil {
				helper.Mylog.Println("Error writing to redis ", err)
			} else {
				helper.Mylog.Println(str)
			}
		}
	} else if err != nil {
		helper.Mylog.Println("Error in redis fetch : ", err)
		return []string{}
	} else {
		err := json.Unmarshal(encFriends, &retFriends)
		if err != nil {
			helper.Mylog.Println("Error unmarshalling data ", err)
			return []string{}
		}
		helper.Mylog.Println("Got from redis Fetch ,", retFriends)
	}
	return retFriends
}

var usrCntG = 0

func (wsh webSocketHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	helper.Mylog.Println(r.Header)

	c, err := wsh.upgrader.Upgrade(w, r, nil)
	if err != nil {
		helper.Mylog.Printf("Error %s when upgrading connection to websocket", err)
		return
	}

	usrCntG++
	usrCnt := usrCntG
	defer c.Close()

	helper.Mylog.Println("Connected to : ", r.RemoteAddr, " usrCnt : ", usrCnt, " usrCntG", usrCntG)

	rdb := makeRedisClient()

	pubsub := rdb.Subscribe(context.Background(), wsh.findFriends(
		strconv.Itoa(usrCnt), rdb)...)
	defer pubsub.Close()

	wsMessages := make(chan string)

	go func() {
		for {
			mt, message, err := c.ReadMessage()
			if err != nil {
				helper.Mylog.Printf("Error %s when reading message from client", err)
				close(wsMessages)
				return
			}

			if mt == websocket.BinaryMessage {
				err = c.WriteMessage(websocket.TextMessage,
					[]byte("server doesn't support binary messages"))
				if err != nil {
					helper.Mylog.Printf("Error %s when sending message to client", err)
				}
				return
			}

			helper.Mylog.Printf("Receive message %s", string(message))
			wsMessages <- string(message)
		}
	}()

	var long, lat float64

	for {
		select {
		case msg := <-pubsub.Channel():
			helper.Mylog.Printf("Recieved message from redis: %s", msg.Payload)
			locStr := strings.Split(msg.Payload, " ")
			friendLat, err := strconv.ParseFloat(locStr[1], 64)
			if lat != 0 && long != 0 {

				if err != nil {
					helper.Mylog.Println("Error parsing friendLat", err)
					return
				}
				friendLong, err := strconv.ParseFloat(locStr[2], 64)
				if err != nil {
					helper.Mylog.Println("Error parsing friendLat", err)
					return
				}
				friendId := locStr[3]
				helper.Mylog.Println("the Haversine distance of ", friendId, " and ", usrCnt, " : ", helper.Haversine(lat, long, friendLat, friendLong))
				if helper.Haversine(lat, long, friendLat, friendLong) < 10.0 {
					helper.Mylog.Printf("%s", friendId)

					response := fmt.Sprintf("[%s] friend is in radius. Mate him %s",
						friendId, strconv.Itoa(usrCnt))
					err = c.WriteMessage(websocket.TextMessage, []byte(response))
					if err != nil {
						helper.Mylog.Printf("Error %s when sending message to client", err)
						return
					}

				}
			}

		case wsMsg, ok := <-wsMessages:
			if !ok {
				helper.Mylog.Println("Websocket channel closed")
				return
			}
			helper.Mylog.Println("Recieved message from websocket", wsMsg)
			if strings.HasPrefix(strings.Trim(string(wsMsg), "\n"), "location") {
				helper.Mylog.Println("start responding to client...")

				locStr := strings.Split(wsMsg, " ")
				lat, err = strconv.ParseFloat(locStr[1], 64)
				if err != nil {
					helper.Mylog.Fatalln("error parsing to float.", err.Error())
				}

				long, err = strconv.ParseFloat(locStr[2], 64)
				if err != nil {
					helper.Mylog.Fatalln("error parsing to float.", err.Error())
				}

				rdb.Publish(context.Background(), strconv.Itoa(usrCnt),
					wsMsg+" "+strconv.Itoa(usrCnt))

			} else if strings.Trim(string(wsMsg), "\n") == "close" {
				helper.Mylog.Println("Closing WS connection")
				return
			}
		}
	}
}

func serveHTTP(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello world"))
}

func setupHTTPServer(port string) {
	helper.Mylog.Println("Setting up http server: ", port)

	mux := http.NewServeMux()
	mux.Handle("/", http.HandlerFunc(serveHTTP))

	server := &http.Server{
		Addr:    "127.0.0.1:" + port,
		Handler: mux,
	}

	helper.Mylog.Println("Starting server on port: ", port)
	server.ListenAndServe()
}

func SetupWSServer(port string) {

	//go setupHTTPServer(port)
	helper.Mylog.Println("Reached here for port : ", port)
	connPool, err := pgxpool.NewWithConfig(context.Background(), db.Config())

	if err != nil {
		helper.Mylog.Fatal("Error while creating connection to database!")
	}

	connection, err := connPool.Acquire(context.Background())
	if err != nil {
		helper.Mylog.Fatal("Error while acquiring connection from database pool")
	}
	helper.Mylog.Println("Db setup for server: ", port)

	defer connection.Release()
	defer connPool.Close()

	webSocketHandler := webSocketHandler{
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		dbPool: connection,
	}

	helper.Mylog.Println("Websocket setup for server: ", port)

	mux := http.NewServeMux()
	mux.Handle("/", webSocketHandler)

	server := &http.Server{
		Addr:    "127.0.0.1:" + port,
		Handler: mux,
	}

	helper.Mylog.Println("Starting server on port: ", port)
	server.ListenAndServe()
}
