package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

// Unauthorized represents an unauthorized error response from the server.
type Unauthorized struct {
	Errors []struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"errors"`
}

// getParts splits a line into three parts: username, password, and cookie.
func getParts(line string) []string {
	return strings.SplitN(line, ":", 3)
}

func checkCookie(client *http.Client, cookie string, username string, password string, outputFile *os.File, channel chan string) {
	// Create a new request with the cookie
	req, err := http.NewRequest("GET", "https://users.roblox.com/v1/users/authenticated", nil)
	if err != nil {
		fmt.Println(err)
		channel <- ""
		return
	}
	req.AddCookie(&http.Cookie{Name: ".ROBLOSECURITY", Value: cookie})

	// Send the request and read the response body
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		channel <- ""
		return
	}
	body, berr := io.ReadAll(resp.Body)
	if berr != nil {
		fmt.Println(berr)
		channel <- ""
		return
	}

	// Unmarshal the response body into the Unauthorized struct
	var accountResponse Unauthorized
	if err := json.Unmarshal(body, &accountResponse); err != nil {
		fmt.Println(err)
		channel <- ""
		return
	}

	// If there are no errors, write the line to the output file
	if accountResponse.Errors == nil {
		var buffer bytes.Buffer
		buffer.WriteString(username)
		buffer.WriteString(":")
		buffer.WriteString(password)
		buffer.WriteString(":")
		buffer.WriteString(cookie)
		buffer.WriteString("\n")

		if _, err := outputFile.Write(buffer.Bytes()); err != nil {
			panic(err)
		}
	}

	channel <- ""
}

func main() {
	// Set the timeout for the http client
	client := &http.Client{Transport: &http.Transport{
		Dial: (&net.Dialer{
			Timeout: 5 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 5 * time.Second,
	}}

	checked := 0

	// Open the input file
	inputFile, err := os.Open("cookies.txt")
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer inputFile.Close()

	// Open the output file for writing
	outputFile, err := os.OpenFile("output.txt", os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer outputFile.Close()

	fmt.Println("Starting...")

	channel := make(chan string)

	scanner := bufio.NewScanner(inputFile)
	for scanner.Scan() {
		parts := getParts(scanner.Text())
		cookie, username, password := parts[2], parts[0], parts[1]

		go checkCookie(client, cookie, username, password, outputFile, channel)
	}

	for range channel {
		checked++
		fmt.Printf("Checked %d cookies\n", checked)
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading file:", err)
		return
	}
}
