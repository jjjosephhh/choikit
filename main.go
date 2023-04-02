package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
)

var osName string
var chatGptApiUrl string
var chatGptApiKey string
var validTools = []string{"one_line"}

func main() {
	toolPtr := flag.String("tool", "", fmt.Sprintf("Which tool you want to use, pick one from %v\n", validTools))
	inputTextPtr := flag.String("input_text", "", "Path to the string file you want to parse")
	outputTextPtr := flag.String("output_text", "out.txt", "Path to the string file you want write to")
	flag.Parse()
	parseEnv()

	switch *toolPtr {
	case "one_line":
		if *inputTextPtr == "" {
			speakAndExit("The input_text command line argument is required")
		}
		// Open the input file
		inputFile, err := os.Open(*inputTextPtr)
		handleError(err)
		defer inputFile.Close()
		// Read the contents of the input file
		inputBytes, err := ioutil.ReadAll(inputFile)
		handleError(err)
		// Remove new lines and tabs
		inputStr := strings.ReplaceAll(string(inputBytes), "\n", " ")
		inputStr = strings.ReplaceAll(inputStr, "\t", " ")
		// Remove multiple consecutive white spaces
		re := regexp.MustCompile(`\s+`)
		inputStr = strings.TrimSpace(re.ReplaceAllString(inputStr, " "))
		// Open the output file
		outputFile, err := os.Create(*outputTextPtr)
		handleError(err)
		defer outputFile.Close()
		// Write the modified text to the output file
		outputWriter := bufio.NewWriter(outputFile)
		_, err = outputWriter.WriteString(inputStr)
		handleError(err)
		// Flush the buffer to ensure all data has been written to the file
		err = outputWriter.Flush()
		handleError(err)
		askChatGPT(fmt.Sprintf("Summarize the following content in a way that a 12 year old would understand. %v\n", inputStr))
	default:
		speakAndExit("The tool type specified could not be handled")
	}

}

func speakAndExit(message string) {
	if osName == "" {
		osName = runtime.GOOS
	}
	// Check if the current operating system is macOS
	if osName == "darwin" {
		// Use the macOS "say" command to speak the message
		cmd := exec.Command("say", message)
		err := cmd.Run()
		if err != nil {
			log.Fatal(err)
		}
	}
	err := errors.New(message)
	log.Fatal(err)
}

func handleError(err error) {
	if err != nil {
		speakAndExit(err.Error())
	}
}

func askChatGPT(prompt string) {
	// Create the API request payload
	payload := map[string]interface{}{
		"prompt":      prompt,
		"max_tokens":  10,
		"temperature": 0.7,
		"n":           1,
		"stop":        "\n",
	}
	// Convert the payload to JSON
	jsonPayload, err := json.Marshal(payload)
	handleError(err)
	// Create the API request
	req, err := http.NewRequest("POST", chatGptApiUrl, bytes.NewBuffer(jsonPayload))
	handleError(err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+chatGptApiKey)
	// Send the API request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending API request:", err)
		return
	}
	defer resp.Body.Close()
	// Read the response body
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading API response:", err)
		return
	}
	// Print the API response
	fmt.Println(string(respBody))
}

func parseEnv() error {
	// Open the .env file
	file, err := os.Open(".env")
	if err != nil {
		return err
	}
	defer file.Close()
	// Parse the lines of the .env file
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, "=")
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		switch key {
		case "CHAT_GPT_API_KEY":
			chatGptApiKey = value
		case "CHAT_GPT_API_URL":
			chatGptApiUrl = value
		}
	}
	return nil
}
