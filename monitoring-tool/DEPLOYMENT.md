# Real-time Monitoring Tool - Deployment Guide

This guide provides comprehensive instructions for deploying the Real-time Monitoring Tool to production environments.

## 📋 Prerequisites

- Docker Engine 20.10+ and Docker Compose 2.0+
- Git
- At least 2GB RAM and 10GB disk space
- Domain name (for production deployment)

## 🚀 Quick Start (Development)

### 1. Clone and Setup

```bash
git clone <your-repository-url>
cd monitoring-tool
```

### 2. Start Development Environment

```bash
# Start only MongoDB for local development
cd backend
docker-compose up -d mongodb

# Run backend locally
go run main.go

# In another terminal, run frontend locally
cd frontend/monitoring-dashboard
npm install
npm start
```

Access the application at:
- Frontend: http://localhost:3000
- Backend API: http://localhost:8080/api/v1
- Health Check: http://localhost:8080/api/v1/health

## 🏭 Production Deployment

### Option 1: Docker Compose (Recommended)

#### 1. Configure Environment

```bash
# Copy the environment template
cp env.prod.example .env

# Edit the environment file
nano .env
```

**Important Environment Variables to Update:**

```bash
# Database
MONGO_ROOT_PASSWORD=your_secure_password_here

# Security
ALLOWED_ORIGINS=https://yourdomain.com
FRONTEND_URL=https://yourdomain.com
REACT_APP_API_URL=https://yourdomain.com/api
REACT_APP_WS_URL=wss://yourdomain.com/ws

# Optional: Enable HTTPS
ENABLE_HTTPS=true
CERT_FILE=/path/to/cert.pem
KEY_FILE=/path/to/key.pem
```

#### 2. Deploy with Script

```bash
# Make the deployment script executable
chmod +x deploy.sh

# Deploy the application
./deploy.sh

# Check deployment status
./deploy.sh status

# View logs
./deploy.sh logs
```

#### 3. Manual Deployment

```bash
# Build and start services
docker-compose -f docker-compose.prod.yml up -d --build

# Check service health
docker-compose -f docker-compose.prod.yml ps

# View logs
docker-compose -f docker-compose.prod.yml logs -f
```

### Option 2: Cloud Deployment

#### AWS ECS/Fargate

1. **Create ECR Repository:**
```bash
aws ecr create-repository --repository-name monitoring-tool
```

2. **Build and Push Images:**
```bash
# Backend
docker build -t monitoring-tool-backend ./backend
docker tag monitoring-tool-backend:latest <account>.dkr.ecr.<region>.amazonaws.com/monitoring-tool:backend
docker push <account>.dkr.ecr.<region>.amazonaws.com/monitoring-tool:backend

# Frontend
docker build -t monitoring-tool-frontend ./frontend/monitoring-dashboard
docker tag monitoring-tool-frontend:latest <account>.dkr.ecr.<region>.amazonaws.com/monitoring-tool:frontend
docker push <account>.dkr.ecr.<region>.amazonaws.com/monitoring-tool:frontend
```

3. **Deploy with ECS Task Definition**

#### Google Cloud Run

1. **Enable APIs:**
```bash
gcloud services enable run.googleapis.com
gcloud services enable cloudbuild.googleapis.com
```

2. **Deploy Backend:**
```bash
cd backend
gcloud run deploy monitoring-tool-backend \
  --source . \
  --platform managed \
  --region us-central1 \
  --allow-unauthenticated \
  --set-env-vars="MONGODB_URI=mongodb+srv://..."
```

3. **Deploy Frontend:**
```bash
cd frontend/monitoring-dashboard
gcloud run deploy monitoring-tool-frontend \
  --source . \
  --platform managed \
  --region us-central1 \
  --allow-unauthenticated
```

#### DigitalOcean App Platform

1. Create `app.yaml`:
```yaml
name: monitoring-tool
services:
- name: backend
  source_dir: /backend
  github:
    repo: your-username/monitoring-tool
    branch: main
  run_command: ./main
  environment_slug: go
  instance_count: 1
  instance_size_slug: basic-xxs
  envs:
  - key: MONGODB_URI
    value: ${db.CONNECTIONSTRING}
  - key: ENVIRONMENT
    value: production

- name: frontend
  source_dir: /frontend/monitoring-dashboard
  github:
    repo: your-username/monitoring-tool
    branch: main
  run_command: npm start
  environment_slug: node-js
  instance_count: 1
  instance_size_slug: basic-xxs
  envs:
  - key: REACT_APP_API_URL
    value: ${backend.PUBLIC_URL}/api
  - key: REACT_APP_WS_URL
    value: wss://${backend.PUBLIC_URL}/ws

databases:
- name: db
  engine: MONGODB
  version: "5"
```

### Option 3: Kubernetes

#### 1. Create Namespace
```bash
kubectl create namespace monitoring-tool
```

#### 2. Create ConfigMap
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: monitoring-tool-config
  namespace: monitoring-tool
data:
  MONGODB_URI: "mongodb://mongodb:27017"
  DATABASE_NAME: "realtime_monitor"
  ENVIRONMENT: "production"
  ALLOWED_ORIGINS: "https://yourdomain.com"
```

#### 3. Deploy MongoDB
```bash
kubectl apply -f - <<EOF
apiVersion: apps/v1
kind: Deployment
metadata:
  name: mongodb
  namespace: monitoring-tool
spec:
  replicas: 1
  selector:
    matchLabels:
      app: mongodb
  template:
    metadata:
      labels:
        app: mongodb
    spec:
      containers:
      - name: mongodb
        image: mongo:7.0
        ports:
        - containerPort: 27017
        env:
        - name: MONGO_INITDB_DATABASE
          value: "realtime_monitor"
---
apiVersion: v1
kind: Service
metadata:
  name: mongodb
  namespace: monitoring-tool
spec:
  selector:
    app: mongodb
  ports:
  - port: 27017
    targetPort: 27017
EOF
```

#### 4. Deploy Backend
```bash
kubectl apply -f - <<EOF
apiVersion: apps/v1
kind: Deployment
metadata:
  name: backend
  namespace: monitoring-tool
spec:
  replicas: 2
  selector:
    matchLabels:
      app: backend
  template:
    metadata:
      labels:
        app: backend
    spec:
      containers:
      - name: backend
        image: your-registry/monitoring-tool:backend
        ports:
        - containerPort: 8080
        envFrom:
        - configMapRef:
            name: monitoring-tool-config
---
apiVersion: v1
kind: Service
metadata:
  name: backend
  namespace: monitoring-tool
spec:
  selector:
    app: backend
  ports:
  - port: 8080
    targetPort: 8080
EOF
```

#### 5. Deploy Frontend
```bash
kubectl apply -f - <<EOF
apiVersion: apps/v1
kind: Deployment
metadata:
  name: frontend
  namespace: monitoring-tool
spec:
  replicas: 2
  selector:
    matchLabels:
      app: frontend
  template:
    metadata:
      labels:
        app: frontend
    spec:
      containers:
      - name: frontend
        image: your-registry/monitoring-tool:frontend
        ports:
        - containerPort: 3000
        env:
        - name: REACT_APP_API_URL
          value: "http://backend:8080/api"
        - name: REACT_APP_WS_URL
          value: "ws://backend:8080/ws"
---
apiVersion: v1
kind: Service
metadata:
  name: frontend
  namespace: monitoring-tool
spec:
  selector:
    app: frontend
  ports:
  - port: 3000
    targetPort: 3000
EOF
```

## 🔧 Configuration

### Environment Variables

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `MONGODB_URI` | MongoDB connection string | `mongodb://localhost:27017` | Yes |
| `DATABASE_NAME` | Database name | `realtime_monitor` | Yes |
| `PORT` | Backend port | `8080` | No |
| `ENVIRONMENT` | Environment mode | `debug` | No |
| `ALLOWED_ORIGINS` | CORS allowed origins | `http://localhost:3000` | No |
| `DEFAULT_INTERVAL` | Default monitoring interval (seconds) | `30` | No |
| `MAX_CONCURRENT_CHECKS` | Maximum concurrent checks | `100` | No |
| `METRICS_RETENTION_DAYS` | Metrics retention period | `30` | No |

### Security Considerations

1. **Change Default Passwords:**
   - Update `MONGO_ROOT_PASSWORD` in production
   - Use strong, unique passwords

2. **Enable HTTPS:**
   - Set `ENABLE_HTTPS=true`
   - Provide SSL certificates
   - Update CORS origins to use HTTPS

3. **Network Security:**
   - Use private networks for database
   - Configure firewall rules
   - Enable TLS for MongoDB connections

4. **Resource Limits:**
   - Set appropriate CPU/memory limits
   - Monitor resource usage
   - Configure auto-scaling

## 📊 Monitoring and Maintenance

### Health Checks

- **Backend Health:** `GET /api/v1/health`
- **Frontend Health:** `GET /` (should return 200)
- **Database Health:** MongoDB connection test

### Logs

```bash
# Docker Compose
docker-compose -f docker-compose.prod.yml logs -f

# Kubernetes
kubectl logs -f deployment/backend -n monitoring-tool
kubectl logs -f deployment/frontend -n monitoring-tool
```

### Backup

```bash
# MongoDB Backup
docker exec realtime_monitor_db_prod mongodump --out /backup
docker cp realtime_monitor_db_prod:/backup ./mongodb-backup

# Restore
docker exec -i realtime_monitor_db_prod mongorestore /backup
```

### Updates

```bash
# Update with zero downtime
docker-compose -f docker-compose.prod.yml pull
docker-compose -f docker-compose.prod.yml up -d --no-deps backend
docker-compose -f docker-compose.prod.yml up -d --no-deps frontend
```

## 🐛 Troubleshooting

### Common Issues

1. **Database Connection Failed:**
   - Check MongoDB is running
   - Verify connection string
   - Check network connectivity

2. **CORS Errors:**
   - Update `ALLOWED_ORIGINS` in environment
   - Check frontend URL configuration

3. **WebSocket Connection Failed:**
   - Verify WebSocket URL configuration
   - Check firewall/load balancer settings
   - Ensure WebSocket support in proxy

4. **High Memory Usage:**
   - Adjust `MAX_CONCURRENT_CHECKS`
   - Reduce `METRICS_RETENTION_DAYS`
   - Monitor and limit concurrent monitors

### Debug Mode

```bash
# Enable debug logging
export ENVIRONMENT=debug
docker-compose -f docker-compose.prod.yml up -d
```

## 📞 Support

For deployment issues:
1. Check logs: `./deploy.sh logs`
2. Verify environment configuration
3. Test individual components
4. Check resource usage and limits

## 🔄 Scaling

### Horizontal Scaling

```bash
# Scale backend
docker-compose -f docker-compose.prod.yml up -d --scale backend=3

# Scale frontend
docker-compose -f docker-compose.prod.yml up -d --scale frontend=2
```

### Load Balancer Configuration

Use a reverse proxy (nginx, Traefik) to distribute load:

```nginx
upstream backend {
    server backend1:8080;
    server backend2:8080;
    server backend3:8080;
}

upstream frontend {
    server frontend1:3000;
    server frontend2:3000;
}

server {
    listen 80;
    server_name yourdomain.com;
    
    location /api {
        proxy_pass http://backend;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
    
    location /ws {
        proxy_pass http://backend;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
    }
    
    location / {
        proxy_pass http://frontend;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```
