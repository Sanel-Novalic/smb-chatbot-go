package main_test // Or another appropriate test package

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const baseAPIURL = "http://localhost:8080/api"

const testDbURL = "postgres://postgres:postgres@localhost:5433/postgres?sslmode=disable"

func TestE2EConversationFlow(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	testChatID := int64(3001)
	testUserID := int64(4001)

	db, err := sql.Open("pgx", testDbURL)
	require.NoError(err, "Failed to connect to test DB")
	defer db.Close()
	err = db.Ping()
	require.NoError(err, "Failed to ping test DB")

	_, err = db.Exec(`DELETE FROM message_history WHERE chat_id = $1;`, testChatID)
	require.NoError(err)
	_, err = db.Exec(`DELETE FROM reviews WHERE chat_id = $1;`, testChatID)
	require.NoError(err)
	_, err = db.Exec(`DELETE FROM conversations WHERE chat_id = $1;`, testChatID)
	require.NoError(err)

	// Step 1: Send initial message
	fmt.Println("E2E Test: Sending initial message...")
	respBody, statusCode, err := sendMessageAPI(testChatID, testUserID, "Hello there")
	require.NoError(err)
	require.Equal(http.StatusOK, statusCode)
	require.NotEmpty(respBody.Reply, "Expected a reply for initial message")
	fmt.Printf("E2E Test: Received reply: %s\n", respBody.Reply)

	// Step 2: Verify DB state after initial message (Optional) ---
	var historyCount int
	err = db.QueryRow(`SELECT count(*) FROM message_history WHERE chat_id = $1`, testChatID).Scan(&historyCount)
	require.NoError(err)
	// Expect 2 messages: user's "Hello there" and the bot's reply
	assert.Equal(2, historyCount, "Expected 2 history entries after first exchange")

	var convoState string
	err = db.QueryRow(`SELECT state FROM conversations WHERE chat_id = $1`, testChatID).Scan(&convoState)
	require.NoError(err)
	assert.Equal("Idle", convoState, "Expected conversation state to be Idle")

	// Step 3: Send "thank you" trigger message
	fmt.Println("E2E Test: Sending trigger message...")
	respBody, statusCode, err = sendMessageAPI(testChatID, testUserID, "That was really helpful, thank you!")
	require.NoError(err)
	require.Equal(http.StatusOK, statusCode)
	require.NotEmpty(respBody.Reply, "Expected a reply after thank you")
	fmt.Printf("E2E Test: Received reply: %s\n", respBody.Reply)

	// Step 4: Verify DB state after trigger
	err = db.QueryRow(`SELECT state FROM conversations WHERE chat_id = $1`, testChatID).Scan(&convoState)
	require.NoError(err)
	assert.Equal("AwaitingReview", convoState, "Expected conversation state to be AwaitingReview")

	// Step 5: Send the actual review
	fmt.Println("E2E Test: Sending review...")
	reviewText := "The service was excellent, very fast!"
	respBody, statusCode, err = sendMessageAPI(testChatID, testUserID, reviewText)
	require.NoError(err)
	require.Equal(http.StatusOK, statusCode)
	require.NotEmpty(respBody.Reply, "Expected a reply after sending review")
	fmt.Printf("E2E Test: Received reply: %s\n", respBody.Reply)

	// Step 6: Verify review saved in DB
	var savedReviewText string
	err = db.QueryRow(`SELECT text FROM reviews WHERE chat_id = $1 ORDER BY received_at DESC LIMIT 1`, testChatID).Scan(&savedReviewText)
	require.NoError(err, "Failed to find saved review in DB")
	assert.Equal(reviewText, savedReviewText, "Saved review text does not match")

	err = db.QueryRow(`SELECT state FROM conversations WHERE chat_id = $1`, testChatID).Scan(&convoState)
	require.NoError(err)
	assert.Equal("Idle", convoState, "Expected conversation state to return to Idle")
}

type apiResponseMessage struct {
	Reply string `json:"reply"`
}

func sendMessageAPI(chatID, userID int64, text string) (apiResponseMessage, int, error) {
	apiURL := fmt.Sprintf("%s/message", baseAPIURL)
	requestBody := map[string]interface{}{
		"chat_id":   chatID,
		"user_id":   userID,
		"user_name": fmt.Sprintf("E2E User %d", userID),
		"text":      text,
	}
	jsonBody, _ := json.Marshal(requestBody)

	req, err := http.NewRequestWithContext(context.Background(), "POST", apiURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return apiResponseMessage{}, 0, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return apiResponseMessage{}, 0, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	var responseBody apiResponseMessage
	if resp.StatusCode == http.StatusOK {
		if err := json.NewDecoder(resp.Body).Decode(&responseBody); err != nil {
			return apiResponseMessage{}, resp.StatusCode, fmt.Errorf("failed to decode JSON response: %w", err)
		}
	}

	return responseBody, resp.StatusCode, nil
}
