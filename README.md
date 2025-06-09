# StreamLLM
A simple web service that allows you to chat with a large language model (LLM) using the [groq.dev](https://groq.dev) API.
The purpose of this app is to play around with how the LLMs streaming works and how to handle edge cases.


## Development
To run the app locally, make sure you clone the repository and install the dependencies:
```bash
go mod tidy
```

Then, you can run the app using:
```golang
go run main.go
```

If you prefer running the app in a Docker container:
```bash
docker compose -f compose-dev.yml up --build
```

## Todo
- [ ] Handle errors and edge cases that could happen from groq's side
- [X] Make groq remmeber the context of the conversation
- [X] Allow stopping the conversation while groq is thinking/processing
- [X] Add a way to save and load the conversation
- [X] Write tests for the code
- [X] Properly display the conversation in the UI: using markdown, etc.
- [X] Add endpoint to check if service is up and running
- [X] Gracefully propagate LLM‑side errors/time‑outs back to the client with meaningful HTTP status & JSON error body
