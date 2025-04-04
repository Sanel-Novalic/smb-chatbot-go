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

## Example Setup

1.  **Environment Variables:** This application requires certain environment variables to be set. Create an environment file called .env

    Then, edit the `.env` file and add your own valid OpenAI API key and adjust any necessary PostgreSQL credentials (though the defaults should work with the provided `docker-compose.yml`).

2.  **Database & Migrations:** The required PostgreSQL database is included in the `docker-compose.yml` file. Migrations (including seeding) will run automatically when starting the services.

3.  **Running the Application:** Ensure you have Docker and Docker installed. From the project root directory, run:
    ```bash
    docker-compose up --build
    ```
    The application API will be available at `http://localhost:8080` (or the `PORT` you specified). The database will be accessible from your host machine on port `5433`.

4. **Using the Application:** You can ask for help from the bot and when you want to check the review process, make a sentence that implies a conversation ender like "I like the recommendation, thank you".
