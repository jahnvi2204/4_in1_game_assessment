# 🐳 Docker Deployment Guide

## Quick Start

### 1. Build and Run Everything with Docker Compose

```bash
# Start PostgreSQL and Backend
docker-compose -f docker-compose.prod.yml up -d

# Check status
docker-compose -f docker-compose.prod.yml ps

# View logs
docker-compose -f docker-compose.prod.yml logs -f
```

### 2. Build Frontend Separately

The frontend needs to be built and served separately. You have two options:

**Option A: Build locally and serve with Nginx**
```bash
cd frontend
npm install
npm run build

# Serve with a simple HTTP server or Nginx
# Or deploy build folder to Netlify/Vercel
```

**Option B: Add Frontend to Docker Compose**

Create `frontend/Dockerfile`:
```dockerfile
FROM node:16-alpine as build
WORKDIR /app
COPY package*.json ./
RUN npm install
COPY . .
RUN npm run build

FROM nginx:alpine
COPY --from=build /app/build /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
```

## Environment Variables

Create a `.env` file in the root directory:

```env
# Database
DB_USER=postgres
DB_PASSWORD=your_secure_password
DB_NAME=connectfour

# Server
PORT=3001

# Kafka (Optional)
KAFKA_BROKERS=localhost:9092
```

## Production Deployment

### Using docker-compose.prod.yml

```bash
# 1. Set environment variables
export DB_PASSWORD=your_secure_password
export DB_USER=postgres

# 2. Start services
docker-compose -f docker-compose.prod.yml up -d

# 3. Check logs
docker-compose -f docker-compose.prod.yml logs backend

# 4. Stop services
docker-compose -f docker-compose.prod.yml down

# 5. Stop and remove volumes (clean slate)
docker-compose -f docker-compose.prod.yml down -v
```

### Build Backend Image Manually

```bash
cd backend
docker build -t connect-four-backend .
docker run -d \
  -p 3001:3001 \
  -e DB_HOST=your-db-host \
  -e DB_USER=postgres \
  -e DB_PASSWORD=your-password \
  -e DB_NAME=connectfour \
  --name connect-four-backend \
  connect-four-backend
```

## Docker Compose Services

- **postgres**: PostgreSQL database
- **backend**: Go backend server
- **zookeeper** & **kafka**: Optional analytics (in docker-compose.yml)

## Access Your Application

- Backend API: http://localhost:3001
- Backend Health: http://localhost:3001/api/health
- WebSocket: ws://localhost:3001/ws

## Troubleshooting

**View logs:**
```bash
docker-compose -f docker-compose.prod.yml logs backend
docker-compose -f docker-compose.prod.yml logs postgres
```

**Restart services:**
```bash
docker-compose -f docker-compose.prod.yml restart backend
```

**Rebuild after code changes:**
```bash
docker-compose -f docker-compose.prod.yml up -d --build
```

**Check if containers are running:**
```bash
docker-compose -f docker-compose.prod.yml ps
```

**Connect to database:**
```bash
docker-compose -f docker-compose.prod.yml exec postgres psql -U postgres -d connectfour
```

