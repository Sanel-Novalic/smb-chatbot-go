# Project Setup and Usage

## 1. Run API (and Database)

You have two options to run the backend API and its database:

### Using Docker:

Execute the following command in your terminal:

```bash
docker-compose up --build
```

### Without Docker

If you wish to use it in an environment, run:
```bash
go run main.go
```


## 2. Run frontend UI:
Navigate to the my-chat-app directory:
```bash
cd my-chat-app
```
And then run:
```bash
npm run dev
```

After running, visit localhost:5173

## How to Trigger a Review

The chat bot (powered by ChatGPT) is designed to request a review when it detects the end of a helpful conversation. It looks for sentences that typically conclude an interaction. If a review is requested, the bot will persist in asking until one is provided by the user.
## Frontend Note

The frontend interface, built with Vue, serves as a simple web UI for interaction. Please note that this frontend component was entirely generated by AI.

## Environment variables for local development / docker
DATABASE_URL="<LOCAL_POSTGRESQL_URL" -> Not needed if you run it via docker
OPENAI_API_KEY="<YOUR_OPENAI_API_KEY"
PORT="8080" -> It also defaults to 8080 if left empty
POSTGRES_USER="POSTGRES_USER"
POSTGRES_PASSWORD="POSTGRES_PASSWORD"
