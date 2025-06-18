## UniAI-Client

### Overview
This is a golang client for the UniAI API, which provides a simple interface to interact with UniDoc AI models and services.
This will read PDF and render pages as images, then send them to the UniDoc AI model for processing based on the provided prompt.
This intends to be used as testing and quality checks of results for UniDoc AI models.

### Requirements
- Go 1.20 or later
- UniCLOUD API Key

### Running The Client
```bash
go run main.go <pdf_file_path> <output_directory> <prompt>
```