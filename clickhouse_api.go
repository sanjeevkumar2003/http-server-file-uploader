package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
)

type UserData struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

var chConn clickhouse.Conn

func main() {
	var err error
	chConn, err = clickhouse.Open(&clickhouse.Options{
		Addr: []string{"localhost:9000"},
		Auth: clickhouse.Auth{
			Database: "default",
			Username: "default",
			Password: "default",
		},
		Debug: true,
	})

	if err != nil {
		log.Fatalf("Failed to connect to ClickHouse : %v", err)
	}

	err = chConn.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS users (
			id UInt32,
			name String,
			email String
        ) ENGINE = 	MergeTree()
		ORDER BY id
    `)

	if err != nil {
		log.Fatalf("Failed to connect to ClickHouse : %v", err)
	}

	http.HandleFunc("/insert", insertHandler)

	fmt.Println("Server Started on : 8082")
	log.Fatal(http.ListenAndServe(":8082", nil))

}

func insertHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only Post Method is Allowed", http.StatusMethodNotAllowed)
		return
	}

	var data UserData
	err := json.NewDecoder(r.Body).Decode(&data)

	if err != nil {
		log.Printf("Json decode error : %v", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	log.Printf("Decoded Json : %+v", data)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	batch, err := chConn.PrepareBatch(ctx, "INSERT INTO users")
	if err != nil {
		log.Printf("PrepareBatch error : %v", err)
		http.Error(w, "Failed to prepare batch", http.StatusInternalServerError)
		return
	}

	err = batch.Append(data.ID, data.Name, data.Email)
	if err != nil {
		log.Printf("Append error: %v", err)
		http.Error(w, "Failed to append data", http.StatusInternalServerError)
		return
	}
	log.Println("Appended to batch")

	if err := batch.Send(); err != nil {
		log.Printf("Send error : %v", err)
		http.Error(w, "Failed to insert data", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Data Inserted Successfully"))
}
