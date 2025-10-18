#!/bin/bash

# Simple setup script for Real-time Monitoring Tool

set -e

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${BLUE}🔧 Setting up Real-time Monitoring Tool...${NC}"

# Check if Go is installed
echo -e "${BLUE}📋 Checking Go installation...${NC}"
if ! command -v go &> /dev/null; then
    echo -e "${YELLOW}⚠️  Go is not installed. Please install Go 1.21 or later.${NC}"
    echo "   Download from: https://golang.org/dl/"
    exit 1
else
    GO_VERSION=$(go version | awk '{print $3}')
    echo -e "${GREEN}✅ Go is installed: $GO_VERSION${NC}"
fi

# Check if Node.js is installed
echo -e "${BLUE}📋 Checking Node.js installation...${NC}"
if ! command -v node &> /dev/null; then
    echo -e "${YELLOW}⚠️  Node.js is not installed. Please install Node.js 18 or later.${NC}"
    echo "   Download from: https://nodejs.org/"
    exit 1
else
    NODE_VERSION=$(node --version)
    echo -e "${GREEN}✅ Node.js is installed: $NODE_VERSION${NC}"
fi

# Check if MongoDB is installed
echo -e "${BLUE}📋 Checking MongoDB installation...${NC}"
if ! command -v mongod &> /dev/null; then
    echo -e "${YELLOW}⚠️  MongoDB is not installed. Please install MongoDB.${NC}"
    echo "   Download from: https://www.mongodb.com/try/download/community"
    exit 1
else
    MONGO_VERSION=$(mongod --version | head -n1 | awk '{print $3}')
    echo -e "${GREEN}✅ MongoDB is installed: $MONGO_VERSION${NC}"
fi

# Create data directory for MongoDB
echo -e "${BLUE}📁 Creating data directory...${NC}"
mkdir -p data/db
echo -e "${GREEN}✅ Data directory created${NC}"

# Setup backend
echo -e "${BLUE}🔧 Setting up backend...${NC}"
cd backend

# Download Go dependencies
echo "Downloading Go dependencies..."
go mod download
go mod tidy
echo -e "${GREEN}✅ Backend dependencies installed${NC}"

# Create .env file
if [ ! -f .env ]; then
    echo "Creating .env file..."
    cat > .env << EOF
MONGODB_URI=mongodb://localhost:27017
DATABASE_NAME=realtime_monitor
PORT=8080
HOST=localhost
ENVIRONMENT=development
ALLOWED_ORIGINS=http://localhost:3000
DEFAULT_INTERVAL=30
DEFAULT_TIMEOUT=10
MAX_CONCURRENT_CHECKS=100
METRICS_RETENTION_DAYS=30
EOF
    echo -e "${GREEN}✅ .env file created${NC}"
fi

cd ..

# Setup frontend
echo -e "${BLUE}⚛️  Setting up frontend...${NC}"
cd frontend/monitoring-dashboard

# Install npm dependencies
echo "Installing npm dependencies..."
npm install
echo -e "${GREEN}✅ Frontend dependencies installed${NC}"

cd ../..

echo ""
echo -e "${GREEN}🎉 Setup completed successfully!${NC}"
echo ""
echo "To start the application:"
echo "  ./start.sh"
echo ""
echo "To stop the application:"
echo "  ./stop.sh"
echo ""
echo "Application URLs:"
echo "  Frontend: http://localhost:3000"
echo "  Backend API: http://localhost:8080/api/v1"
echo "  WebSocket: ws://localhost:8080/ws"
