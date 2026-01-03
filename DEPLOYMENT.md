# 🚀 Deployment Guide

This guide covers deploying the 4-in-a-Row game to production.

## 📋 Prerequisites

- Go binary (for backend)
- PostgreSQL database (hosted or managed service)
- Static hosting (for frontend)
- Optional: Kafka cluster (for analytics)

## 🎯 Deployment Options

### Option 1: Traditional VPS/Server (Recommended for Full Control)

#### Backend Deployment

1. **Build the Go binary:**
   ```bash
   cd backend
   go build -o connect-four-server main.go
   ```

2. **Transfer files to server:**
   ```bash
   # Using SCP
   scp connect-four-server user@your-server.com:/opt/connect-four/
   scp -r .env user@your-server.com:/opt/connect-four/
   ```

3. **Set up systemd service** (create `/etc/systemd/system/connect-four.service`):
   ```ini
   [Unit]
   Description=Connect Four Game Server
   After=network.target postgresql.service

   [Service]
   Type=simple
   User=your-user
   WorkingDirectory=/opt/connect-four
   EnvironmentFile=/opt/connect-four/.env
   ExecStart=/opt/connect-four/connect-four-server
   Restart=always
   RestartSec=5

   [Install]
   WantedBy=multi-user.target
   ```

4. **Start the service:**
   ```bash
   sudo systemctl daemon-reload
   sudo systemctl enable connect-four
   sudo systemctl start connect-four
   sudo systemctl status connect-four
   ```

#### Frontend Deployment

1. **Build the React app:**
   ```bash
   cd frontend
   npm install
   npm run build
   ```

2. **Deploy to static hosting:**
   - **Netlify**: Drag and drop the `build` folder
   - **Vercel**: `vercel --prod`
   - **AWS S3 + CloudFront**: Upload `build` folder to S3 bucket
   - **Nginx**: Copy `build` folder to `/var/www/connect-four/`

3. **Update API URLs:**
   Create `.env.production` in frontend:
   ```env
   REACT_APP_API_URL=https://api.yourdomain.com
   REACT_APP_WS_URL=wss://api.yourdomain.com
   ```

#### Nginx Configuration (if hosting frontend on same server)

```nginx
# Frontend
server {
    listen 80;
    server_name yourdomain.com;
    
    root /var/www/connect-four/build;
    index index.html;
    
    location / {
        try_files $uri $uri/ /index.html;
    }
}

# Backend API
server {
    listen 80;
    server_name api.yourdomain.com;
    
    location / {
        proxy_pass http://localhost:3001;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

### Option 2: Docker Deployment

#### Create Dockerfile for Backend

```dockerfile
# backend/Dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o server main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/server .
COPY --from=builder /app/.env .env

EXPOSE 3001
CMD ["./server"]
```

#### Create docker-compose.prod.yml

```yaml
version: '3.8'

services:
  postgres:
    image: postgres:15
    environment:
      POSTGRES_DB: connectfour
      POSTGRES_USER: ${DB_USER}
      POSTGRES_PASSWORD: ${DB_PASSWORD}
    volumes:
      - postgres_data:/var/lib/postgresql/data
    restart: always

  backend:
    build:
      context: ./backend
      dockerfile: Dockerfile
    ports:
      - "3001:3001"
    environment:
      - PORT=3001
      - DB_HOST=postgres
      - DB_USER=${DB_USER}
      - DB_PASSWORD=${DB_PASSWORD}
      - DB_NAME=connectfour
      - DB_PORT=5432
    depends_on:
      - postgres
    restart: always

volumes:
  postgres_data:
```

#### Deploy with Docker:

```bash
docker-compose -f docker-compose.prod.yml up -d
```

### Option 3: Cloud Platforms

#### Heroku

**Backend:**
1. Create `Procfile` in backend:
   ```
   web: ./connect-four-server
   ```
2. Deploy:
   ```bash
   heroku create your-app-name
   heroku addons:create heroku-postgresql:hobby-dev
   git push heroku main
   ```

**Frontend:**
- Use Netlify or Vercel (see below)

#### Railway

1. Connect your GitHub repo
2. Railway auto-detects Go and Node.js
3. Set environment variables in Railway dashboard
4. Deploy automatically on push

#### Render

**Backend:**
1. Create new Web Service
2. Build command: `cd backend && go build -o server main.go`
3. Start command: `./server`
4. Add PostgreSQL database
5. Set environment variables

**Frontend:**
1. Create new Static Site
2. Build command: `cd frontend && npm install && npm run build`
3. Publish directory: `frontend/build`

#### Netlify (Frontend)

1. Connect GitHub repo
2. Build settings:
   - Base directory: `frontend`
   - Build command: `npm run build`
   - Publish directory: `frontend/build`
3. Environment variables:
   ```
   REACT_APP_API_URL=https://api.yourdomain.com
   REACT_APP_WS_URL=wss://api.yourdomain.com
   ```

#### Vercel (Frontend)

1. Install Vercel CLI: `npm i -g vercel`
2. In frontend directory:
   ```bash
   vercel
   ```
3. Set environment variables in Vercel dashboard

### Option 4: AWS/GCP/Azure

#### AWS (EC2 + S3)

**Backend on EC2:**
- Follow VPS instructions above
- Use AWS RDS for PostgreSQL
- Use Application Load Balancer for scaling

**Frontend on S3:**
```bash
aws s3 sync frontend/build s3://your-bucket-name --delete
aws cloudfront create-invalidation --distribution-id YOUR_ID --paths "/*"
```

#### Google Cloud Platform

**Backend:**
- Use Cloud Run (containerized)
- Use Cloud SQL for PostgreSQL

**Frontend:**
- Use Firebase Hosting or Cloud Storage + CDN

## 🔐 Environment Variables

Create `.env` file on your server:

```env
# Server
PORT=3001

# Database
DB_USER=postgres
DB_HOST=localhost
DB_NAME=connectfour
DB_PASSWORD=your_secure_password
DB_PORT=5432

# Kafka (Optional)
KAFKA_BROKERS=kafka1:9092,kafka2:9092

# CORS (if needed)
ALLOWED_ORIGINS=https://yourdomain.com,https://www.yourdomain.com
```

## 🔒 Security Checklist

- [ ] Use HTTPS (SSL/TLS certificates)
- [ ] Set secure database passwords
- [ ] Enable CORS only for your frontend domain
- [ ] Use environment variables for secrets
- [ ] Set up firewall rules (only allow necessary ports)
- [ ] Enable rate limiting
- [ ] Set up monitoring and logging
- [ ] Regular database backups
- [ ] Keep dependencies updated

## 📊 Monitoring

### Health Check Endpoint
- `GET /api/health` - Returns server status

### Logging
- Backend logs to stdout/stderr
- Use log aggregation service (e.g., Loggly, Papertrail)
- Or configure systemd journal: `journalctl -u connect-four -f`

## 🔄 CI/CD Pipeline (GitHub Actions Example)

Create `.github/workflows/deploy.yml`:

```yaml
name: Deploy

on:
  push:
    branches: [main]

jobs:
  deploy-backend:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - uses: actions/setup-go@v2
        with:
          go-version: '1.21'
      - name: Build
        run: |
          cd backend
          go build -o server main.go
      - name: Deploy
        run: |
          # Add your deployment commands here
          # e.g., scp, rsync, or cloud CLI commands

  deploy-frontend:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - uses: actions/setup-node@v2
        with:
          node-version: '16'
      - name: Build
        run: |
          cd frontend
          npm install
          npm run build
      - name: Deploy
        run: |
          # Add your deployment commands here
```

## 🐳 Quick Docker Deployment

If you want the simplest deployment:

```bash
# 1. Build and start everything
docker-compose up -d

# 2. Your app is running!
# Backend: http://localhost:3001
# Frontend: http://localhost:3000
# PostgreSQL: localhost:5432
```

## 📝 Post-Deployment

1. **Test the deployment:**
   - Visit your frontend URL
   - Test game functionality
   - Check WebSocket connections
   - Verify leaderboard API

2. **Set up domain name:**
   - Point DNS to your server
   - Configure SSL certificate (Let's Encrypt)
   - Update CORS settings

3. **Monitor:**
   - Check server logs
   - Monitor database connections
   - Set up alerts for errors

## 🆘 Troubleshooting

**Backend won't start:**
- Check database connection
- Verify environment variables
- Check port availability
- Review logs: `journalctl -u connect-four -n 50`

**Frontend can't connect:**
- Verify API_URL in environment
- Check CORS settings
- Ensure WebSocket URL uses `wss://` for HTTPS

**Database errors:**
- Verify PostgreSQL is running
- Check connection string
- Ensure database exists
- Check firewall rules

## 📚 Additional Resources

- [Go Deployment Best Practices](https://go.dev/doc/deployment)
- [React Production Build](https://create-react-app.dev/docs/deployment)
- [PostgreSQL Production Tuning](https://www.postgresql.org/docs/current/runtime-config.html)


