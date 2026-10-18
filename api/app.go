package api

import (
	"api/internal/message"
	"api/internal/message/repository"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

const rabbitmqUrl = "0.0.0.0:5672"
const webPort = ":8080"

func main() {
	fmt.Println("Starting the server")

	var router = mux.NewRouter()
	messageRepo, err := repository.NewRabbitMQMessageRepository(rabbitmqUrl)
	if err != nil {
		panic(err.Error())
		return
	}
	messageService := message.NewDefaultMessageService(messageRepo)
	message.AddMessageHandlers(router, messageService)

	fmt.Println("Starting Web Server on port", webPort)
	log.Fatal(http.ListenAndServe(webPort, router))
}
