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

## Setup & Run

Set your API key

```bash
export API_KEY=SECRET123
```

Start the server

```bash
./imgcast
```

The server runs at `http://localhost:8080`.

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
