# 🎮 4 in a Row - Connect Four Game

A real-time multiplayer 4-in-a-Row game built with **Go backend** and React frontend, featuring competitive bot AI, WebSocket communication, PostgreSQL persistence, and Kafka analytics.

## 🚀 Features

- **Real-time Multiplayer**: Play against other players or a competitive bot
- **Smart Matchmaking**: Automatic pairing with 10-second timeout before bot fallback
- **Competitive Bot AI**: Strategic bot that blocks wins and creates winning opportunities
- **Reconnection Support**: 30-second window to reconnect if disconnected
- **Leaderboard**: Track wins, losses, and draws for all players
- **Kafka Analytics**: Decoupled analytics service for game metrics
- **PostgreSQL Persistence**: Store completed games and leaderboard data

## 📋 Prerequisites

- Go (v1.21 or higher)
- Node.js (v16 or higher) - for frontend only
- PostgreSQL (v12 or higher)
- Kafka (optional, for analytics)

## 🛠️ Local Setup Instructions

### Prerequisites

- **Go** (v1.21 or higher) - [Download](https://go.dev/dl/)
- **Node.js** (v16 or higher) - [Download](https://nodejs.org/)
- **PostgreSQL** (v12 or higher) - [Download](https://www.postgresql.org/download/)
- **Docker** (optional, for PostgreSQL/Kafka) - [Download](https://www.docker.com/)

### Step 1: Clone the Repository

```bash
git clone <your-repo-url>
cd assignment
```

### Step 2: Install Go Dependencies

```bash
cd backend
go mod download
cd ..
```

### Step 3: Set Up PostgreSQL Database

#### Option A: Using Docker (Easiest)

```bash
# Start PostgreSQL with Docker
docker-compose up -d postgres

# Wait a few seconds for PostgreSQL to start
```

#### Option B: Local PostgreSQL Installation

```bash
# Connect to PostgreSQL
psql -U postgres

# Create database
CREATE DATABASE connectfour;

# Exit psql
\q
```

### Step 4: Configure Environment Variables

Create a `.env` file in the `backend` directory (optional, defaults work for local):

```bash
# backend/.env (optional)
PORT=3001
DB_USER=postgres
DB_HOST=localhost
DB_NAME=connectfour
DB_PASSWORD=postgres
DB_PORT=5432
KAFKA_BROKERS=localhost:9092
```

Or set environment variables:

```bash
# Windows PowerShell
$env:PORT=3001
$env:DB_USER=postgres
$env:DB_HOST=localhost
$env:DB_NAME=connectfour
$env:DB_PASSWORD=postgres
$env:DB_PORT=5432

# Linux/Mac
export PORT=3001
export DB_USER=postgres
export DB_HOST=localhost
export DB_NAME=connectfour
export DB_PASSWORD=postgres
export DB_PORT=5432
```

### Step 5: Set Up Kafka (Optional)

Kafka is optional - the app works without it. If you want analytics:

```bash
# Start Kafka using Docker
docker-compose up -d kafka zookeeper

# Or install Kafka locally and start it
```

The application will continue to work without Kafka, but analytics events won't be processed.

### Step 6: Start the Backend

```bash
cd backend
go run main.go
```

Or build and run:

```bash
cd backend
go build -o server main.go
./server
```

You should see:
```
Server starting on port 3001
```

### Step 7: Start the Frontend

Open a **new terminal window**:

```bash
cd frontend
npm install
npm start
```

The frontend will open automatically at `http://localhost:3000`

### Step 8: Access the Application

- **Frontend:** http://localhost:3000
- **Backend API:** http://localhost:3001
- **Health Check:** http://localhost:3001/api/health
- **Leaderboard API:** http://localhost:3001/api/leaderboard
- **WebSocket:** ws://localhost:3001/ws

### Quick Test

1. Open http://localhost:3000 in your browser
2. Enter a username and click "Join Game"
3. Wait 10 seconds - a bot will join automatically
4. Start playing!

## 🎮 How to Play

1. **Enter Username**: Enter your username and click "Join Game"
2. **Wait for Opponent**: The system will try to match you with another player
3. **Bot Fallback**: If no opponent joins within 10 seconds, a bot will start the game
4. **Make Moves**: Click the column buttons (↓) to drop your disc
5. **Win Condition**: Connect 4 discs vertically, horizontally, or diagonally
6. **Reconnection**: If you disconnect, you have 30 seconds to reconnect using your username

## 🏗️ Project Structure

```
assignment/
├── backend/
│   ├── main.go                # Main server file
│   ├── go.mod                 # Go dependencies
│   ├── game/
│   │   ├── game.go           # Game state management
│   │   └── logic.go          # Game rules and win detection
│   ├── matchmaking/
│   │   └── matchmaking.go    # Player matching logic
│   ├── bot/
│   │   └── bot.go            # Competitive bot AI
│   └── analytics/
│       └── analytics.go       # Kafka analytics integration
├── frontend/
│   ├── public/
│   ├── src/
│   │   ├── App.js             # Main React component
│   │   ├── index.js           # React entry point
│   │   └── index.css          # Styles
│   └── package.json
├── docker-compose.yml         # PostgreSQL and Kafka setup
└── README.md
```

## 🔌 API Endpoints

### REST API

- `GET /api/leaderboard` - Get leaderboard data
- `GET /api/health` - Health check

### WebSocket Messages

**Client → Server:**
- `{ type: 'join', username: 'player1' }` - Join matchmaking
- `{ type: 'rejoin', username: 'player1', gameId: 'uuid' }` - Rejoin game
- `{ type: 'makeMove', gameId: 'uuid', column: 3 }` - Make a move

**Server → Client:**
- `{ type: 'waiting', message: '...' }` - Waiting for opponent
- `{ type: 'gameState', game: {...} }` - Game state update
- `{ type: 'playerDisconnected', message: '...' }` - Player disconnected
- `{ type: 'playerReconnected', username: '...' }` - Player reconnected
- `{ type: 'error', message: '...' }` - Error message

## 🤖 Bot AI Strategy

The competitive bot uses a strategic approach:

1. **Win Detection**: Checks if it can win in the next move
2. **Block Opponent**: Blocks opponent's immediate win threat
3. **Strategic Positioning**: Evaluates board positions and prefers center columns
4. **Position Scoring**: Uses heuristic evaluation for optimal moves

## 📊 Analytics

The analytics service tracks:
- Game start/end events
- Move events
- Game duration
- Winner statistics
- Games per day/hour

Events are sent to Kafka topic `game-events` and consumed by the analytics service for processing.

## 🚢 Production Deployment

### Option 1: Deploy to Render (Recommended)

#### Backend Deployment:

1. **Go to [render.com](https://render.com)** and sign in
2. **Click "New +"** → **"Web Service"**
3. **Connect your GitHub repository**
4. **Configure settings:**
   - **Name:** `connect-four-backend`
   - **Region:** Choose closest to you
   - **Branch:** `main` (or your default branch)
   - **Root Directory:** `backend`
   - **Runtime:** `Go` (NOT Docker)
   - **Build Command:** `go build -o server main.go`
   - **Start Command:** `./server`
5. **Add PostgreSQL Database:**
   - Click **"New +"** → **"PostgreSQL"**
   - Name: `connectfour-postgres`
   - Database: `connectfour`
   - Plan: **Free**
   - Click **"Create Database"**
6. **Set Environment Variables:**
   - In your Web Service → **"Environment"** tab
   - Add these variables (copy from PostgreSQL service):
     ```
     PORT=3001
     DB_HOST=<from PostgreSQL service>
     DB_USER=<from PostgreSQL service>
     DB_PASSWORD=<from PostgreSQL service>
     DB_NAME=connectfour
     DB_PORT=5432
     ```
7. **Click "Create Web Service"**
8. **Wait for deployment** (2-3 minutes)
9. **Copy your backend URL:** `https://connect-four-backend.onrender.com`

#### Frontend Deployment:

1. **In Render dashboard, click "New +"** → **"Static Site"**
2. **Connect the same GitHub repository**
3. **Configure settings:**
   - **Name:** `connect-four-frontend`
   - **Branch:** `main`
   - **Root Directory:** `frontend`
   - **Build Command:** `npm install && npm run build`
   - **Publish Directory:** `frontend/build`
4. **Set Environment Variables:**
   - Click **"Environment"** tab
   - Add:
     ```
     REACT_APP_API_URL=https://connect-four-backend.onrender.com
     REACT_APP_WS_URL=wss://connect-four-backend.onrender.com/ws
     ```
   - (Replace with your actual backend URL from step 9 above)
5. **Click "Create Static Site"**
6. **Wait for deployment** (2-3 minutes)
7. **Your production URL:** `https://connect-four-frontend.onrender.com` 🎉

**Note:** Use `wss://` (not `ws://`) for WebSocket in production!

---

### Option 2: Deploy to Railway

1. **Go to [railway.app](https://railway.app)** and sign up with GitHub
2. **Click "New Project"** → **"Deploy from GitHub repo"**
3. **Select your repository**
4. **Deploy Backend:**
   - Railway auto-detects Go
   - Set **Root Directory:** `backend`
   - Add **PostgreSQL** database (Railway auto-configures env vars)
5. **Deploy Frontend:**
   - Add new service → **GitHub Repo**
   - Set **Root Directory:** `frontend`
   - **Build Command:** `npm install && npm run build`
   - **Start Command:** `npx serve -s build -l 3000`
   - Set environment variables:
     ```
     REACT_APP_API_URL=https://your-backend.up.railway.app
     REACT_APP_WS_URL=wss://your-backend.up.railway.app/ws
     ```

---

### Option 3: Alternative PostgreSQL (If Render doesn't show PostgreSQL option)

If you can't find PostgreSQL in Render, use a free external database:

#### Using Supabase (Free):

1. **Go to [supabase.com](https://supabase.com)** and sign up
2. **Create a new project**
3. **Go to Settings** → **Database**
4. **Copy connection details:**
   - Host, User, Password, Database name
5. **In Render backend environment variables, add:**
   ```
   DB_HOST=db.xxxxx.supabase.co
   DB_USER=postgres
   DB_PASSWORD=<your-supabase-password>
   DB_NAME=postgres
   DB_PORT=5432
   ```

#### Using Neon (Free):

1. **Go to [neon.tech](https://neon.tech)** and sign up
2. **Create a new project**
3. **Copy connection string**
4. **Add to Render environment variables** (same format as above)

---

### Other Deployment Options

- **Docker:** See [DOCKER_DEPLOY.md](DOCKER_DEPLOY.md) for Docker deployment
- **VPS/Server:** See [DEPLOYMENT.md](DEPLOYMENT.md) for traditional server setup
- **Netlify/Vercel:** Deploy frontend only, use separate backend hosting

For detailed instructions, see:
- [DEPLOY_NOW.md](DEPLOY_NOW.md) - Quick deployment guide
- [DEPLOYMENT.md](DEPLOYMENT.md) - Complete deployment options
- [PRODUCTION_DEPLOY.md](PRODUCTION_DEPLOY.md) - Detailed production guide

## 🧪 Testing

Test the game by:
1. Opening multiple browser tabs/windows
2. Joining with different usernames
3. Testing bot gameplay by waiting 10 seconds
4. Testing reconnection by closing and reopening a tab

## 📝 Notes

- Games are stored in-memory while active
- Completed games are persisted to PostgreSQL
- Leaderboard updates automatically after each game
- Kafka integration is optional - the app works without it
- Bot makes moves automatically after player moves

## 🐛 Troubleshooting

**Go Build Errors:**
- Ensure Go is installed: `go version`
- Run `go mod download` in the backend directory
- Check that all dependencies are available

**Database Connection Error:**
- Check PostgreSQL is running
- Verify database credentials
- Ensure database exists
- Check connection string in `backend/game/game.go`

**WebSocket Connection Failed:**
- Check backend server is running
- Verify port 3001 is not blocked
- Check firewall settings

**Kafka Errors:**
- Kafka is optional - app works without it
- Check Kafka is running if using analytics
- Verify broker addresses

## 📄 License

This project is created for assignment purposes.

## 👤 Author

Created as part of Backend Engineering Intern Assignment

#
