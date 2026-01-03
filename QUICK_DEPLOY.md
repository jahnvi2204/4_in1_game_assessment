# 🚀 Quick Docker Deployment

## One-Command Deployment

```bash
docker-compose -f docker-compose.prod.yml up -d
```

That's it! Your backend and database are now running.

## Step-by-Step

### 1. Create Environment File (Optional)

Create `.env` file:
```env
DB_USER=postgres
DB_PASSWORD=your_secure_password
DB_NAME=connectfour
PORT=3001
```

### 2. Deploy Backend + Database

```bash
# Build and start
docker-compose -f docker-compose.prod.yml up -d

# Check status
docker-compose -f docker-compose.prod.yml ps

# View logs
docker-compose -f docker-compose.prod.yml logs -f backend
```

### 3. Deploy Frontend

**Option A: Build and Deploy to Netlify/Vercel**
```bash
cd frontend
npm install
npm run build
# Upload build/ folder to Netlify or Vercel
```

**Option B: Add to Docker (see below)**

### 4. Access Your App

- Backend: http://localhost:3001
- API Health: http://localhost:3001/api/health
- WebSocket: ws://localhost:3001/ws

## Frontend in Docker

Add this to `docker-compose.prod.yml`:

```yaml
  frontend:
    build:
      context: ./frontend
      dockerfile: Dockerfile
    ports:
      - "3000:80"
    environment:
      - REACT_APP_API_URL=http://localhost:3001
      - REACT_APP_WS_URL=ws://localhost:3001/ws
    depends_on:
      - backend
```

Create `frontend/Dockerfile`:
```dockerfile
FROM node:16-alpine as build
WORKDIR /app
COPY package*.json ./
RUN npm install
COPY . .
ARG REACT_APP_API_URL
ARG REACT_APP_WS_URL
ENV REACT_APP_API_URL=$REACT_APP_API_URL
ENV REACT_APP_WS_URL=$REACT_APP_WS_URL
RUN npm run build

FROM nginx:alpine
COPY --from=build /app/build /usr/share/nginx/html
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
```

## Useful Commands

```bash
# Stop everything
docker-compose -f docker-compose.prod.yml down

# Stop and remove data
docker-compose -f docker-compose.prod.yml down -v

# Rebuild after code changes
docker-compose -f docker-compose.prod.yml up -d --build

# View logs
docker-compose -f docker-compose.prod.yml logs -f

# Restart a service
docker-compose -f docker-compose.prod.yml restart backend
```

## Production Tips

1. **Use environment variables** - Never hardcode passwords
2. **Use Docker secrets** - For sensitive data in production
3. **Set up reverse proxy** - Use Nginx/Traefik for SSL
4. **Enable logging** - Configure log aggregation
5. **Set resource limits** - In docker-compose.yml

## Troubleshooting

**Backend won't start:**
```bash
docker-compose -f docker-compose.prod.yml logs backend
```

**Database connection issues:**
```bash
docker-compose -f docker-compose.prod.yml exec postgres psql -U postgres -d connectfour
```

**Rebuild everything:**
```bash
docker-compose -f docker-compose.prod.yml down
docker-compose -f docker-compose.prod.yml build --no-cache
docker-compose -f docker-compose.prod.yml up -d
```


