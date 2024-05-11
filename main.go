package main

import (
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/mdtosif/lru-go/lru"
)

type ApiEntry struct {
	Key     string `json:"key"`
	Value   string `json:"value"`
	Timeout string `json:"timeout"`
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,// read buffer size of the websocket
	WriteBufferSize: 1024,// write buffer size of websocket
	CheckOrigin:     func(r *http.Request) bool { return true }, // allow connection from all origin
}

var cache = lru.NewLRUCache(40, 5*time.Second)
var clients []*websocket.Conn

// broadcast the cache details to all the connected ws client
func BroadCast() {
	var entries []ApiEntry
	ent, err := cache.GetAll()
	if err != nil{
		return
	}
	for _, v := range ent {
		entries = append(entries, ApiEntry{Key: v.Key, Value: v.Value, Timeout: v.Timestamp.Add(5 * time.Second).Format("3:04:05 PM")})
	}

	// send message to the clients
	for _, client := range clients {
		client.WriteJSON(entries)
	}

}

func main() {

	// Create a new Gin router
	router := gin.Default()

	config := cors.DefaultConfig()
	config.AddAllowHeaders("Authorization")
	config.AllowCredentials = true
	config.AllowAllOrigins = false
	// allowing all origin
	config.AllowOriginFunc = func(origin string) bool {
		return true
	}
	router.Use(cors.New(config))

	// POST handler for creating cache
	router.POST("/cache", func(ctx *gin.Context) {
		var entry ApiEntry

		// Bind the JSON data from the request body into the cache struct
		if err := ctx.BindJSON(&entry); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		cache.Set(entry.Key, entry.Value)

		// send the data to the user
		ctx.JSON(http.StatusOK, entry)
	})

	router.GET("/cache", func(ctx *gin.Context) {
		var entry ApiEntry

		// Bind the JSON data from the request body into the cache struct
		if err := ctx.BindJSON(&entry); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		value, err := cache.Get(entry.Key)

		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		entry.Value = value.Value
		entry.Timeout = (value.Timestamp.Add(5 * time.Second)).Format("3:04:05 PM")

		ctx.JSON(http.StatusOK, entry)
	})

	router.GET("/cache/all", func(ctx *gin.Context) {
		var entries []ApiEntry

		ent, err := cache.GetAll()

		if ent == nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "no entry"})
			return
		}

		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		for _, v := range ent {

			entries = append(entries, ApiEntry{Key: v.Key, Value: v.Value, Timeout: v.Timestamp.Add(5 * time.Second).Format("3:04:05 PM")})
		}

		ctx.JSON(http.StatusOK, entries)
	})

	// delete the cache
	router.DELETE("/cache", func(ctx *gin.Context) {
		var entry ApiEntry

		// Bind the JSON data from the request body into the job struct
		if err := ctx.BindJSON(&entry); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		cache.Delete(entry.Key)

		ctx.JSON(http.StatusOK, entry)
	})

	router.Any("/ws", func(ctx *gin.Context) {
		conn, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			println("error: ", err.Error())
			return
		}

		clients = append(clients, conn)
	})


	// Start a goroutine to call the function every second
    go func() {
        for {
            // broadcasting the message to all user
            BroadCast()

            // Wait for one second
            time.Sleep(time.Second)
        }
    }()
	// Run the server on port 8080
	router.Run(":8080")

}
