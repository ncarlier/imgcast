# Real-Time Shared Image Server

This is a lightweight server written in **Go** that allows real-time sharing of images with multiple isolated rooms. When a new image is uploaded to a room, it's instantly pushed to all viewers in that room **without requiring a page reload**.

## Features

- **Multi-room support** with isolated storage and authentication
- **Real-time updates** using **Server-Sent Events (SSE)**
- Image upload via HTTP `POST` request with **Basic Authentication**
- **Per-room authentication** using htpasswd files
- **Admin-controlled room creation**

## Prerequisites

- [Go](https://golang.org/dl/) ≥ 1.21
- `htpasswd` (from Apache) for creating credentials, or use Go's bcrypt
- `curl` or any HTTP client to test uploads

## Quick Start

### 1. Create Admin Credentials

```bash
mkdir -p var
htpasswd -nbB admin yourpassword > var/.htpasswd
```

### 2. Build and Run

```bash
make build
./release/imgcast
```

### 3. Create a Room

```bash
curl -F "image=@myimage.jpg" \
  -u admin:yourpassword \
  http://localhost:8080/myroom/upload
```

### 4. View the Room

Open `http://localhost:8080/myroom` in your browser

## Multi-Room Usage

### Creating Rooms

Rooms are created automatically on first upload using **admin credentials**:

```bash
curl -F "image=@image.jpg" -u admin:password http://localhost:8080/demo/upload
curl -F "image=@image.jpg" -u admin:password http://localhost:8080/team-alpha/upload
```

### Viewing Rooms

Each room has its own viewer URL:
- `http://localhost:8080/demo`
- `http://localhost:8080/team-alpha`

### Adding Users to Rooms

```bash
# Add another user to a room
htpasswd -nbB alice alicepass >> var/rooms/team-alpha/.htpasswd

# Alice can now upload
curl -F "image=@image.jpg" -u alice:alicepass http://localhost:8080/team-alpha/upload
```

## API Endpoints

### Room Endpoints

- `POST /{roomname}/upload` - Upload image (requires Basic Auth)
- `GET /{roomname}/live` - Get current room image
- `GET /{roomname}/events` - SSE stream for room updates
- `GET /{roomname}` - Room viewer page

## Environment Variables

| Variable | Description | Default | Example |
|----------|-------------|---------|---------|
| `ROOM_BASE_DIR` | Directory for room storage | `var` | `/data/rooms` |
| `PORT` | Server port | `8080` | `3000` |
| `BASE_PATH` | Base path for hosting | `/` | `/imgcast/` |

## Authentication Model

### Admin Level (var/.htpasswd)
- Controls who can **create new rooms**
- Created manually before starting the server

### Room Level (var/rooms/{room}/.htpasswd)
- Controls who can **upload to a specific room**
- Initialized with the admin user who created the room
- Can be extended by adding more users

## Room Name Validation

Room names must be:
- Alphanumeric characters only
- Can include dash (`-`) and underscore (`_`)
- Examples: `team-alpha`, `room_123`, `demo`
- Invalid: `room!`, `my room`, `special@room`

## Docker usage

A Dockerfile is provided for easy deployment.
Build the Docker image:

```bash
make image
```

Run the container:

```bash
docker run -p 8080:8080 -e ROOMS_BASE_DIR=/home/nonroot ncarlier/imgcast
```

## License

MIT — Free for personal and commercial use.
