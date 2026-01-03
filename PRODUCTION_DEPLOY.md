# 🌐 Production Deployment Guide

This guide will help you deploy your Connect Four game to production and get a live URL.

## 🚀 Quick Deploy Options

### Option 1: Railway (Recommended - Easiest)

Railway supports Docker and provides free tier with PostgreSQL.

#### Steps:

1. **Sign up at [railway.app](https://railway.app)**

2. **Create a new project:**
   - Click "New Project"
   - Select "Deploy from GitHub repo" (or upload code)

3. **Deploy Backend:**
   - Click "New Service" → "GitHub Repo"
   - Select your repository
   - Railway will auto-detect the Go backend
   - **Root Directory:** Set to `backend`
   - **Build Command:** `go build -o server main.go`
   - **Start Command:** `./server`

4. **Add PostgreSQL:**
   - Click "New" → "Database" → "PostgreSQL"
   - Railway will automatically set `DATABASE_URL` environment variable

5. **Set Environment Variables:**
   - Go to backend service → Variables
   - Add:
     ```
     PORT=3001
     DB_HOST=<from-postgres-service>
     DB_USER=<from-postgres-service>
     DB_PASSWORD=<from-postgres-service>
     DB_NAME=<from-postgres-service>
     DB_PORT=5432
     ```
   - Railway provides these automatically via `${{Postgres.DATABASE_URL}}`

6. **Get your backend URL:**
   - Railway provides: `https://your-backend.railway.app`
   - Note this URL!

7. **Deploy Frontend:**
   - Click "New Service" → "GitHub Repo"
   - **Root Directory:** Set to `frontend`
   - **Build Command:** `npm install && npm run build`
   - **Start Command:** `npx serve -s build -l 3000`
   - **Environment Variables:**
     ```
     REACT_APP_API_URL=https://your-backend.railway.app
     REACT_APP_WS_URL=wss://your-backend.railway.app/ws
     ```

8. **Get your frontend URL:**
   - Railway provides: `https://your-frontend.railway.app`
   - **This is your production URL!** 🎉

---

### Option 2: Render (Also Easy)

#### Backend Deployment:

1. **Sign up at [render.com](https://render.com)**

2. **Create Web Service:**
   - Click "New" → "Web Service"
   - Connect your GitHub repo
   - **Settings:**
     - **Name:** `connect-four-backend`
     - **Environment:** `Docker`
     - **Root Directory:** `backend`
     - **Build Command:** `go build -o server main.go`
     - **Start Command:** `./server`

3. **Add PostgreSQL Database:**
   - Click "New" → "PostgreSQL"
   - Render will provide connection details

4. **Set Environment Variables:**
   - In your Web Service → Environment
   - Add:
     ```
     PORT=3001
     DB_HOST=<from-postgres-service>
     DB_USER=<from-postgres-service>
     DB_PASSWORD=<from-postgres-service>
     DB_NAME=<from-postgres-service>
     DB_PORT=5432
     ```

5. **Get backend URL:**
   - Render provides: `https://connect-four-backend.onrender.com`

#### Frontend Deployment:

1. **Create Static Site:**
   - Click "New" → "Static Site"
   - Connect your GitHub repo
   - **Settings:**
     - **Name:** `connect-four-frontend`
     - **Root Directory:** `frontend`
     - **Build Command:** `npm install && npm run build`
     - **Publish Directory:** `frontend/build`

2. **Set Environment Variables:**
   ```
   REACT_APP_API_URL=https://connect-four-backend.onrender.com
   REACT_APP_WS_URL=wss://connect-four-backend.onrender.com/ws
   ```

3. **Get frontend URL:**
   - Render provides: `https://connect-four-frontend.onrender.com`
   - **This is your production URL!** 🎉

---

### Option 3: Fly.io (Docker-based)

1. **Install Fly CLI:**
   ```bash
   # Windows (PowerShell)
   powershell -Command "iwr https://fly.io/install.ps1 -useb | iex"
   ```

2. **Login:**
   ```bash
   fly auth login
   ```

3. **Deploy Backend:**
   ```bash
   cd backend
   fly launch
   # Follow prompts
   # Add PostgreSQL: fly postgres create
   # Attach: fly postgres attach -a <postgres-app-name>
   ```

4. **Deploy Frontend:**
   ```bash
   cd frontend
   fly launch
   ```

---

## 🔧 Configuration Files Needed

### For Railway/Render Backend:

Create `backend/railway.json` (optional):
```json
{
  "$schema": "https://railway.app/railway.schema.json",
  "build": {
    "builder": "NIXPACKS"
  },
  "deploy": {
    "startCommand": "./server",
    "restartPolicyType": "ON_FAILURE",
    "restartPolicyMaxRetries": 10
  }
}
```

### For Frontend:

Update `frontend/package.json` to include serve:
```json
{
  "scripts": {
    "start": "react-scripts start",
    "build": "react-scripts build",
    "serve": "serve -s build -l 3000"
  },
  "devDependencies": {
    "serve": "^14.2.0"
  }
}
```

---

## 🌍 Your Production URLs

After deployment, you'll have:

- **Frontend:** `https://your-app.railway.app` or `https://your-app.onrender.com`
- **Backend API:** `https://your-backend.railway.app/api/leaderboard`
- **WebSocket:** `wss://your-backend.railway.app/ws`

---

## ✅ Post-Deployment Checklist

- [ ] Test frontend URL loads
- [ ] Test game functionality
- [ ] Test WebSocket connection
- [ ] Test leaderboard API
- [ ] Verify HTTPS is working (wss:// for WebSocket)
- [ ] Check CORS settings allow your frontend domain

---

## 🐛 Troubleshooting

**Backend won't start:**
- Check database connection string
- Verify all environment variables are set
- Check logs in Railway/Render dashboard

**Frontend can't connect:**
- Verify `REACT_APP_API_URL` points to your backend
- Use `wss://` (not `ws://`) for WebSocket in production
- Check CORS settings in backend

**WebSocket connection fails:**
- Ensure backend supports WebSocket upgrades
- Use `wss://` protocol for HTTPS sites
- Check firewall/proxy settings

---

## 📝 Quick Commands

**Railway CLI:**
```bash
npm i -g @railway/cli
railway login
railway link
railway up
```

**Render CLI:**
```bash
npm i -g render-cli
render login
render deploy
```

---

## 🎯 Recommended: Railway

Railway is the easiest option because:
- ✅ Free tier available
- ✅ Auto-detects Go and Node.js
- ✅ Built-in PostgreSQL
- ✅ Automatic HTTPS
- ✅ Easy environment variable management
- ✅ GitHub integration

**Get started:** https://railway.app/new

