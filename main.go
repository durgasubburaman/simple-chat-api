package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/garyburd/redigo/redis"
)

// SimpleChatAPIVersion defines the version of the API
const SimpleChatAPIVersion = "1.1"

func main() {
	// handle POST /api/messages/new
	http.HandleFunc("/api/messages/new", PostNewMessage)

	// handle GET /api/messages
	http.HandleFunc("/api/messages", GetMessages)

	// handle GET /api/version
	http.HandleFunc("/api/version", GetVersion)

	// start the http server
	log.Printf("The Simple Chat API server has started and is listening on port %d...", 82)
	log.Fatal(http.ListenAndServe(":82", nil))
}

// PostNewMessage posts a new message to the chat
func PostNewMessage(writer http.ResponseWriter, request *http.Request) {
	// ensure that the http method used is POST
	if strings.ToUpper(request.Method) != "POST" {
		writer.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintln(writer, "This API only supports POST method")
		return
	}

	// create a new JSON decode on the request body
	decoder := json.NewDecoder(request.Body)
	var newMessage Message
	err := decoder.Decode(&newMessage)

	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(writer, "Something went wrong with your request: "+err.Error())
		return
	}

	jsonMessage, jsonErr := json.Marshal(newMessage)
	if jsonErr != nil {
		panic(jsonErr)
	}

	redisClient := getRedisClient()
	redisClient.Do("LPUSH", "messages", string(jsonMessage))

	writer.WriteHeader(http.StatusCreated)
}

// GetMessages returns the list of messages on the chat
func GetMessages(writer http.ResponseWriter, request *http.Request) {
	redisClient := getRedisClient()
	jsonMessages, err := redis.Strings(redisClient.Do("LRANGE", "messages", "0", "100"))
	if err != nil {
		panic(err)
	}

	writer.WriteHeader(http.StatusOK)
	fmt.Fprint(writer, jsonMessages)
}

// GetVersion returns the version of the API
func GetVersion(writer http.ResponseWriter, request *http.Request) {
	writer.WriteHeader(http.StatusOK)
	fmt.Fprint(writer, "Simple Chat - API Version: "+SimpleChatAPIVersion)
}

func getRedisClient() redis.Conn {
	redisEndpoint := os.Getenv("SIMPLE_CHAT_REDIS_ENDPOINT")
	redisClient, err := redis.Dial("tcp", redisEndpoint)

	if err != nil {
		panic(err)
	}

	return redisClient
}

// Message represents an instance of a message in the chat
type Message struct {
	// Content represents the content of the message
	Content string

	// Username is the name of the user who posts the message
	Username string

	// MessageTime is the time when the message has been posted
	MessageTime time.Time
}
