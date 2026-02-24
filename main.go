package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

func main() {

	serverURL := "http://localhost:8080/_stats"
	// serverURL := "http://localhost:8080/"
	interval := 2 * time.Second
	fmt.Println("We start polling the server")

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		sendGetRequest(serverURL)
	}

}

func sendGetRequest(url string) {

	resp, err := http.Get(url)
	if err != nil {
		log.Printf("Error sending request: %v", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response: %v", err)
		return
	}

	status := resp.Status
	fmt.Printf("Status: %s\n", status)

	// check type and pritnt --
	fmt.Printf("Type of body: %T\n", body)
	fmt.Printf("Body: %s\n", body)
	fmt.Printf("Body: %s\n", body[0:2])
	fmt.Printf("Body as string: %s\n", string(body))
	fmt.Printf("---\n\n")

}
