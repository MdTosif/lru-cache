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
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

var cache = lru.NewLRUCache(5, 60*time.Second) // capacity = 2, expiration = 5 seconds
var clients []*websocket.Conn

func BroadCast() {
	var entries []ApiEntry
	ent, err := cache.GetAll()
	if err != nil{
		return
	}
	for _, v := range ent {
		entries = append(entries, ApiEntry{Key: v.Key, Value: v.Value, Timeout: v.Timestamp.Add(30 * time.Second).Format("3:04:05 PM")})
	}

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
	// I think you should whitelist a limited origins instead:
	//  config.AllowAllOrigins = []{"xxxx", "xxxx"}
	config.AllowOriginFunc = func(origin string) bool {
		return true
	}
	router.Use(cors.New(config))

	// POST handler for creating a job
	router.POST("/cache", func(ctx *gin.Context) {
		var entry ApiEntry

		// Bind the JSON data from the request body into the job struct
		if err := ctx.BindJSON(&entry); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		cache.Set(entry.Key, entry.Value)

		// Process the job (e.g., save it to a database)
		// Here, we simply echo back the received job
		ctx.JSON(http.StatusOK, entry)
	})

	router.GET("/cache", func(ctx *gin.Context) {
		var entry ApiEntry

		// Bind the JSON data from the request body into the job struct
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
		entry.Timeout = (value.Timestamp.Add(30 * time.Second)).Format("2006-01-02 15:04:05")

		// Process the job (e.g., save it to a database)
		// Here, we simply echo back the received job
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

			entries = append(entries, ApiEntry{Key: v.Key, Value: v.Value, Timeout: v.Timestamp.Add(30 * time.Second).Format("3:04:05 PM")})
		}

		// Process the job (e.g., save it to a database)
		// Here, we simply echo back the received job
		ctx.JSON(http.StatusOK, entries)
	})

	// WebSocket handler
	router.DELETE("/cache", func(ctx *gin.Context) {
		var entry ApiEntry

		// Bind the JSON data from the request body into the job struct
		if err := ctx.BindJSON(&entry); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		cache.Delete(entry.Key)

		// Process the job (e.g., save it to a database)
		// Here, we simply echo back the received job
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
            // Call your function here
            BroadCast()

            // Wait for one second
            time.Sleep(time.Second)
        }
    }()
	// Run the server on port 8080
	router.Run(":8080")

}
