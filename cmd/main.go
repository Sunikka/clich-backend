package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/sunikka/clich-backend/internal/database"
	userService "github.com/sunikka/clich-backend/internal/http/user"
	"github.com/sunikka/clich-backend/internal/http/ws"
	"golang.org/x/net/websocket"

	_ "github.com/lib/pq"
)

func main() {

	// Env variables
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}
	serverPort := os.Getenv("SERVER_PORT")
	loginPort := os.Getenv("AUTH_PORT")

	dbURL := os.Getenv("DB_URL")

	conn, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("connecting to database failed:", err)
	}
	db := database.New(conn)
	loginService := userService.NewUserService(loginPort, db)

	go loginService.Run(loginPort)

	startMainService(serverPort)
}

// func startLoginService(port string) {
// 	http.HandleFunc("/login", routes.HandleLogin)
// 	log.Println("Login Service listening on port", port)
// 	http.ListenAndServe(port, nil)
// }

func startMainService(port string) {
	server := ws.NewServer()
	http.Handle("/ws", websocket.Handler(server.HandleConn))

	log.Println("Server listening on port", port)
	http.ListenAndServe(":"+port, nil)
}
