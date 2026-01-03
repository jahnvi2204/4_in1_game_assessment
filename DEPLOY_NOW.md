# 🚀 Deploy to Production NOW

## Fastest Way: Railway (5 minutes)

### Step 1: Sign Up
1. Go to https://railway.app
2. Sign up with GitHub

### Step 2: Deploy Backend
1. Click **"New Project"** → **"Deploy from GitHub repo"**
2. Select your repository
3. Click **"Add Service"** → **"GitHub Repo"** (if not auto-added)
4. In service settings:
   - **Root Directory:** `backend`
   - **Build Command:** `go build -o server main.go`
   - **Start Command:** `./server`
5. Click **"Add Database"** → **"PostgreSQL"**
6. Railway auto-sets database env vars!

### Step 3: Get Backend URL
- Railway gives you: `https://your-backend.up.railway.app`
- **Copy this URL!**

### Step 4: Deploy Frontend
1. Click **"New Service"** → **"GitHub Repo"**
2. Select same repository
3. In service settings:
   - **Root Directory:** `frontend`
   - **Build Command:** `npm install && npm run build`
   - **Start Command:** `npx serve -s build -l 3000`
4. Go to **"Variables"** tab, add:
   ```
   REACT_APP_API_URL=https://your-backend.up.railway.app
   REACT_APP_WS_URL=wss://your-backend.up.railway.app/ws
   ```
   (Replace with your actual backend URL)

### Step 5: Get Your Production URL! 🎉
- Railway gives you: `https://your-frontend.up.railway.app`
- **This is your live production URL!**

---

## Alternative: Render (Also Easy)

### Backend:
1. Go to https://render.com
2. **New** → **Web Service**
3. Connect GitHub repo
4. Settings:
   - **Root Directory:** `backend`
   - **Build Command:** `go build -o server main.go`
   - **Start Command:** `./server`
5. **New** → **PostgreSQL** (add database)
6. Set env vars from database

### Frontend:
1. **New** → **Static Site**
2. Connect GitHub repo
3. Settings:
   - **Root Directory:** `frontend`
   - **Build Command:** `npm install && npm run build`
   - **Publish Directory:** `frontend/build`
4. Set env vars:
   ```
   REACT_APP_API_URL=https://your-backend.onrender.com
   REACT_APP_WS_URL=wss://your-backend.onrender.com/ws
   ```

---

## ⚡ Quick Checklist

- [ ] Backend deployed and running
- [ ] PostgreSQL database added
- [ ] Environment variables set
- [ ] Frontend deployed with correct API URLs
- [ ] Test your production URL!

---

## 🎯 Your URLs Will Look Like:

**Railway:**
- Frontend: `https://connect-four-frontend-production.up.railway.app`
- Backend: `https://connect-four-backend-production.up.railway.app`

**Render:**
- Frontend: `https://connect-four-frontend.onrender.com`
- Backend: `https://connect-four-backend.onrender.com`

---

## 💡 Pro Tips

1. **Use Railway** - It's the easiest and has a good free tier
2. **Set environment variables** before first deploy
3. **Use `wss://`** (not `ws://`) for WebSocket in production
4. **Test immediately** after deployment
5. **Check logs** if something doesn't work

---

## 🆘 Need Help?

If deployment fails:
1. Check service logs in Railway/Render dashboard
2. Verify all environment variables are set
3. Make sure database is running
4. Check that build commands are correct

**Ready? Go to https://railway.app and deploy!** 🚀

