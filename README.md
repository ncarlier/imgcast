# Real-Time Shared Image Server

This is a lightweight server written in **Go** that allows real-time sharing of a single image with all connected web clients. When a new image is uploaded, it's instantly pushed to all viewers **without requiring a page reload**.

## Features

- Shared image display for all visitors
- **Real-time updates** using **WebSocket**
- Image upload via HTTP `POST` request
- **API key protection** via environment variable
- Temporary image storage in the system's `/tmp` directory

## Prerequisites

- [Go](https://golang.org/dl/) ≥ 1.21
- `Make` for building
- `curl` or any HTTP client to test uploads

## Environment Variables

| Variable | Description | Default | Example |
|----------|-------------|---------|---------|
| `API_KEY` | API key for upload authentication | `secret` | `SECRET123` |
| `PORT` | Server port | `8080` | `3000` |
| `BASE_PATH` | Base path for hosting under subdirectory | `/` | `/imgcast/` |

## Setup & Run

Set your API key and optionally configure the port and base path

```bash
export API_KEY=SECRET123
export PORT=8080        # Optional, defaults to 8080
export BASE_PATH=/      # Optional, defaults to / (root path)
```

Start the server

```bash
./imgcast
```

The server runs at `http://localhost:8080` by default, or on the port specified by the `PORT` environment variable. Use `BASE_PATH` to host the application under a subdirectory (e.g., `BASE_PATH=/imgcast/` will make it available at `http://localhost:8080/imgcast/`).

## Usage

### Viewing the image

Visit:

```
http://localhost:8080/
```

The currently shared image will be displayed and updated in real time for all viewers when a new one is uploaded.

### Uploading an image

```bash
curl -F "image=@A.jpg" -H "X-API-Key: $API_KEY" http://localhost:8080/upload
```

* Replace `A.jpg` with your image file
* All connected users will see the new image instantly


## License

MIT — Free for personal and commercial use.
