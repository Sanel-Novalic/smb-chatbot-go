<script setup>
import { ref, onMounted, nextTick } from 'vue';

// --- Configuration ---
const chatID = ref(2001); // <<<--- SET YOUR CHAT ID HERE
const userID = ref(456); // <<<--- SET YOUR USER ID HERE
const apiUrlMessage = 'http://localhost:8080/api/message';
const apiUrlHistoryBase = 'http://localhost:8080/api/history/';
// --- End Configuration ---

const messages = ref([]); // Array to hold chat messages: { id: number, text: string, is_user: boolean }
const newMessage = ref(''); // Input field model
const isLoadingHistory = ref(false);
const error = ref(null);
const messageListRef = ref(null); // Ref for scrolling

// --- Functions ---

const scrollToBottom = () => {
  // Use nextTick to wait for the DOM to update after adding a message
  nextTick(() => {
    const el = messageListRef.value;
    if (el) {
      el.scrollTop = el.scrollHeight;
    }
  });
};

const fetchHistory = async () => {
  if (!chatID.value) {
    error.value = "Chat ID is not set.";
    return;
  }
  isLoadingHistory.value = true;
  error.value = null;
  console.log(`Fetching history for chat ID: ${chatID.value}`);

  try {
    const response = await fetch(`${apiUrlHistoryBase}${chatID.value}/`);
    console.log(`History fetch status: ${response.status}`);

    if (!response.ok) {
      const errorText = await response.text();
      console.error(`History fetch failed: ${response.status}`, errorText);
      throw new Error(`Failed to fetch history: ${response.status} ${errorText}`);
    }

    const historyData = await response.json();
    console.log('Received history data:', historyData);

    // Assuming historyData is an array like [{ text: string, is_user: boolean }, ...]
    messages.value = historyData.map((msg, index) => ({
        id: Date.now() + index, // Simple unique ID generation
        text: msg.text,
        is_user: msg.is_user ?? false
    }));

    scrollToBottom(); // Scroll after loading history

  } catch (err) {
    console.error("Error fetching history:", err);
    error.value = `Error fetching history: ${err.message}`;
    messages.value = [];
  } finally {
    isLoadingHistory.value = false;
  }
};

const sendMessage = async () => {
  const textToSend = newMessage.value.trim();
  if (!textToSend) return;

  if (!chatID.value || !userID.value) {
      error.value = "Chat ID or User ID is not set.";
      return;
  }

  // 1. Optimistically display user message
  const userMessage = {
    id: Date.now(), // Use timestamp as a simple unique ID
    text: textToSend,
    is_user: true,
  };
  messages.value.push(userMessage);
  newMessage.value = ''; // Clear input field
  scrollToBottom(); // Scroll after adding user message

  // 2. Prepare payload for backend
  const payload = {
    chat_id: chatID.value,
    user_id: userID.value,
    text: textToSend,
  };

  console.log('Sending message:', payload);

  // 3. Send message to backend
  error.value = null;
  try {
    const response = await fetch(apiUrlMessage, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(payload),
    });

    console.log(`Send message status: ${response.status}`);

    if (!response.ok) {
       const errorText = await response.text();
       console.error(`Send message failed: ${response.status}`, errorText);
      throw new Error(`Failed to send message: ${response.status} ${errorText}`);
    }

    // --- *** NEW: Handle Bot Reply *** ---
    const result = await response.json(); // Parse the JSON response from the backend
    console.log('Send message response:', result);

    // Check if the response contains a 'reply' field with text
    if (result && typeof result.reply === 'string' && result.reply.trim() !== '') {
        // Create the bot message object
        const botMessage = {
            id: Date.now() + 1, // Slightly different ID from user message
            text: result.reply,
            is_user: false, // It's a message from the bot
        };
        // Add the bot's message to the chat display
        messages.value.push(botMessage);
        scrollToBottom(); // Scroll again to show the new bot message
    }
    // --- *** End NEW *** ---

  } catch (err) {
    console.error("Error sending message:", err);
    error.value = `Error sending message: ${err.message}`;
    // Optional: Remove the optimistically added user message on failure
    // messages.value.pop(); // Be careful if bot reply was already added
  }
};

// --- Lifecycle Hooks ---
onMounted(() => {
  fetchHistory();
});

</script>

<template>
  <div class="chat-container">
    <h2>Chat Bot (Chat ID: {{ chatID }})</h2>

    <div class="chat-window">
      <div class="message-list" ref="messageListRef">
         <div v-if="isLoadingHistory">Loading history...</div>
         <div v-else-if="error && messages.length === 0" class="error-message">Error loading history: {{ error }}</div>
         <div v-else-if="messages.length === 0 && !isLoadingHistory">No messages yet. Start chatting!</div>
        <div
          v-for="message in messages"
          :key="message.id"
          :class="['message', message.is_user ? 'user-message' : 'bot-message']"
        >
          <p>{{ message.text }}</p>
        </div>
      </div>

      <form @submit.prevent="sendMessage" class="input-area">
        <input
          type="text"
          v-model="newMessage"
          placeholder="Type your message..."
          :disabled="isLoadingHistory"
          aria-label="Chat message input"
        />
        <button type="submit" :disabled="!newMessage.trim() || isLoadingHistory">Send</button>
      </form>
       <div v-if="error && !isLoadingHistory" class="error-message sending-error">
            {{ error }}
        </div>
    </div>
  </div>
</template>

<style>
body {
  font-family: sans-serif;
  margin: 0;
  background-color: #f4f4f4;
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 100vh;
}

.chat-container {
  width: 100%;
  max-width: 600px;
  background-color: #fff;
  border-radius: 8px;
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.1);
  display: flex;
  flex-direction: column;
  overflow: hidden; /* Ensure child elements fit */
}

.chat-container h2 {
    text-align: center;
    padding: 15px;
    margin: 0;
    background-color: #4a90e2;
    color: white;
    font-size: 1.2em;
}

.chat-window {
  display: flex;
  flex-direction: column;
  height: 70vh; /* Fixed height for the window */
  max-height: 700px; /* Max height */
}

.message-list {
  flex-grow: 1; /* Takes available space */
  overflow-y: auto; /* Enables scrolling */
  padding: 15px;
  border-bottom: 1px solid #eee;
  color: black;
  display: flex;
  flex-direction: column;
  gap: 10px; /* Space between messages */
}

.message {
  padding: 8px 12px;
  border-radius: 15px;
  max-width: 75%;
  word-wrap: break-word; /* Wrap long words */
}

.message p {
    margin: 0; /* Remove default paragraph margin */
}

.user-message {
  background-color: #dcf8c6;
  align-self: flex-end; /* Align user messages to the right */
  border-bottom-right-radius: 5px; /* Slightly different corner */
}

.bot-message {
  background-color: #e5e5ea;
  align-self: flex-start; /* Align bot messages to the left */
  border-bottom-left-radius: 5px; /* Slightly different corner */
}

.input-area {
  display: flex;
  padding: 10px;
  border-top: 1px solid #eee;
  background-color: #f9f9f9;
}

.input-area input {
  flex-grow: 1;
  padding: 10px;
  border: 1px solid #ccc;
  border-radius: 20px;
  margin-right: 10px;
  font-size: 1em;
}

.input-area input:focus {
    outline: none;
    border-color: #4a90e2;
}

.input-area button {
  padding: 10px 20px;
  border: none;
  background-color: #4a90e2;
  color: white;
  border-radius: 20px;
  cursor: pointer;
  font-size: 1em;
  transition: background-color 0.2s ease;
}

.input-area button:hover {
  background-color: #357abd;
}

.input-area button:disabled {
  background-color: #a0c4e9;
  cursor: not-allowed;
}

.error-message {
    padding: 10px;
    color: #D8000C; /* Red */
    background-color: #FFD2D2; /* Light red */
    border: 1px solid #D8000C;
    border-radius: 5px;
    margin: 10px;
    text-align: center;
}
.sending-error {
    margin: 0 10px 10px 10px; /* Margin only below input area */
}

</style>