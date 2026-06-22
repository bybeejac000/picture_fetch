# 🖼️ picture_fetch

> A smart digital photo frame backend in Go — powered by [Immich](https://immich.app), supercharged with face recognition and a UniFi Protect doorbell.

`picture_fetch` curates a live slideshow from your Immich photo library and reacts to the world in real time: when someone walks up to your doorbell, it snaps a frame, figures out *who* it is, labels them, and drops the picture straight into your running slideshow.

---

## ✨ Features

| | Feature | What it does |
|---|---|---|
| 🎞️ | **Curated slideshow** | Pulls a random batch of images from your Immich database and serves them as a ready-to-display playlist. |
| 🎂 | **Birthday mode** | If it's someone's birthday today, the slideshow automatically fills with photos of *that* person. |
| 🔔 | **Doorbell integration** | Subscribes to UniFi Protect events and captures a snapshot the moment a **person** is detected. |
| 🧠 | **Face recognition** | Runs detected faces through Immich's ML server and matches them against known people using **pgvector** similarity search. |
| 🏷️ | **Live annotation** | Draws bounding boxes + names on the captured frame, and flags any unrecognized visitor with a "New Face Detected!" banner. |
| ⚡ | **Real-time injection** | Pushes the freshly-labeled doorbell photo into the live slideshow over a WebSocket. |
| 🚀 | **Redis-backed** | Slideshow playlists are cached in Redis for fast, refreshable delivery. |

---

## 🧩 How it works

```
                       ┌─────────────┐
   Immich (Postgres) ──┤  slideshow  ├──► Redis playlist ──► your frame
                       └─────────────┘
                                              ▲
                                              │ inject (WebSocket)
   UniFi Protect ──► person detected ─► capture snapshot
        (events ws)                         │
                                            ▼
                                   Immich ML server (faces)
                                            │
                                            ▼
                                  pgvector match → annotate
```

---

## 🛠️ Tech Stack

- **Go** `1.26`
- **PostgreSQL** + [`pgvector`](https://github.com/pgvector/pgvector) — Immich's database & face embeddings
- **Redis** — slideshow playlist cache
- **Gorilla WebSocket** — UniFi event stream + slideshow injection
- **fogleman/gg** — image annotation
- **Immich** — photo library & machine-learning face models
- **UniFi Protect** — doorbell snapshots & smart-detection events

---

## 🌐 Endpoints

| Method | Route | Description |
|--------|-------|-------------|
| `GET` | `/refresh` | Rebuilds the slideshow playlist in Redis. |
| `GET` | `/photo?file=<path>` | Serves a captured/annotated image by path. |
| `WS` | `/injectPictures` | WebSocket the display client connects to so new photos can be pushed live. |

---

## ⚙️ Configuration

All configuration is loaded from a `.env` file in the project root.

```env
# Immich
IMMICH_URL=
IMMICH_API_KEY=
IMMICH_RO_API_KEY=

# Database (Immich Postgres)
DB_HOST=
DB_PORT=
DB_USER=
DB_PASSWORD=
DB_NAME=

# Redis
REDIS_URL=
REDIS_PORT=
PHOTOS_LIST_KEY=

# Server
GO_LISTEN_PORT=
SLIDESHOW_BATCH_SIZE=

# UniFi Protect doorbell
DOORBELL_HOST=
DOORBELL_ID=
UNIFI_API_KEY=

# Machine learning (face detection/recognition)
ML_SERVER=
ML_FACE_MODEL=
```

---

## ▶️ Running

```bash
# Install dependencies
go mod download

# Build & run
go run ./cmd/photo_fetch
```

On startup the server will:
1. Connect to Postgres and Redis.
2. Build the initial slideshow playlist.
3. Open the HTTP server on `GO_LISTEN_PORT`.
4. Subscribe to the UniFi Protect event stream (auto-reconnecting).

---

## 📁 Project Layout

```
cmd/photo_fetch/      # entrypoint — wires everything together
internal/config/      # .env loading
internal/database/    # Postgres connection
internal/store/       # Redis playlist cache
internal/slideshow/   # playlist building + birthday logic
internal/faces/       # snapshot capture, detection, matching, annotation
internal/realtime/    # UniFi event stream + WebSocket injection
internal/routes/      # HTTP handlers
```
