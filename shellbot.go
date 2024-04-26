package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/logrusorgru/aurora"
	"github.com/spf13/cobra"
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OpenAIResponse struct {
	Id      string `json:"id"`
	Object  string `json:"object"`
	Created int    `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

var mainCmd = &cobra.Command{
	Use:   "shellbot [query]",
	Short: "Ask a virtual Bash and GNU utilities expert for help",
	Args:  cobra.MinimumNArgs(1),
	Run:   runQuery,
}

func runQuery(cmd *cobra.Command, args []string) {
	query := strings.Join(args, " ")

	response, err := queryOpenAI(query)
	if err != nil {
		log.Fatalf("Error querying OpenAI: %v", err)
	}

	formattedResponse, err := formatResponse(response)
	if err != nil {
		log.Fatalf("Error formatting response: %v", err)
	}

	fmt.Println(formattedResponse)
}

func queryOpenAI(query string) (string, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("OPENAI_API_KEY not set in environment variables")
	}

	requestPayload := struct {
		Model    string    `json:"model"`
		Messages []Message `json:"messages"`
	}{
		Model: "gpt-3.5-turbo",
		Messages: []Message{
			{Role: "system", Content: "You are an assistant trained as an expert in Bash and GNU utilities."},
			{Role: "user", Content: query},
		},
	}

	jsonData, err := json.Marshal(requestPayload)
	if err != nil {
		return "", fmt.Errorf("error marshaling JSON: %v", err)
	}

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func formatResponse(jsonStr string) (string, error) {
	var resp OpenAIResponse
	err := json.Unmarshal([]byte(jsonStr), &resp)
	if err != nil {
		return "", fmt.Errorf("Failed to parse JSON response: %v", err)
	}

	if len(resp.Choices) == 0 {
		return "No response from API", nil
	}

	// Use aurora library to apply color
	response := resp.Choices[0].Message.Content
	response = strings.ReplaceAll(response, "```bash\n", aurora.Index(105, "").String()) // Apply a specific color
	response = strings.ReplaceAll(response, "\n```", aurora.Reset("").String())
	return response, nil
}

func main() {
	cobra.CheckErr(mainCmd.Execute())
}
