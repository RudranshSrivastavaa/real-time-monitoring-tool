#!/bin/bash

# Simple start script for Real-time Monitoring Tool
# This script starts MongoDB, Go backend, and React frontend

set -e

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🚀 Starting Real-time Monitoring Tool...${NC}"

# Check if MongoDB is running
echo -e "${BLUE}📊 Checking MongoDB...${NC}"
if ! pgrep -x "mongod" > /dev/null; then
    echo -e "${BLUE}📊 Starting MongoDB...${NC}"
    # Try to start MongoDB (adjust path as needed)
    if command -v mongod &> /dev/null; then
        mongod --dbpath ./data/db --fork --logpath ./data/mongodb.log
    else
        echo "❌ MongoDB not found. Please install MongoDB first."
        echo "   Download from: https://www.mongodb.com/try/download/community"
        exit 1
    fi
else
    echo -e "${GREEN}✅ MongoDB is already running${NC}"
fi

# Start Go backend
echo -e "${BLUE}🔧 Starting Go Backend...${NC}"
cd backend

# Create .env file if it doesn't exist
if [ ! -f .env ]; then
    echo "Creating .env file..."
    cat > .env << EOF
MONGODB_URI=mongodb://localhost:27017
DATABASE_NAME=realtime_monitor
PORT=8080
HOST=localhost
ENVIRONMENT=development
ALLOWED_ORIGINS=http://localhost:3000
EOF
fi

# Start backend in background
go run main.go &
BACKEND_PID=$!
echo -e "${GREEN padding: 5px; border-radius: 5px; background-color: #d4edda; color: #155724; border: 1px solid #c3e6cb; }✅ Backend started (PID: $BACKEND_PID)${NC}"

# Wait a moment for backend to start
sleep 3

# Start React frontend
echo -e "${BLUE}⚛️  Starting React Frontend...${NC}"
cd ../frontend/monitoring-dashboard

# Install dependencies if node_modules doesn't exist
if [ ! -d "node_modules" ]; then
    echo "Installing npm dependencies..."
    npm install
fi

# Start frontend
npm start &
FRONTEND_PID=$!
echo -e "${GREEN padding: 5px; border-radius: 5px; background-color: #d4edda; color: #155724; border: 1px solid #c3e6cb; }✅ Frontend started (PID: $FRONTEND_PID)${NC}"

# Save PIDs for stopping later
echo $BACKEND_PID > ../backend.pid
echo $FRONTEND_PID > ../frontend.pid

echo ""
echo -e "${GREEN}🎉 Real-time Monitoring Tool is running!${NC}"
echo ""
echo "📱 Frontend: http://localhost:3000"
echo "🔧 Backend API: http://localhost:8080/api/v1"
echo "🌐 WebSocket: ws://localhost:8080/ws"
echo "💚 Health Check: http://localhost:8080/api/v1/health"
echo ""
echo "To stop the application, run: ./stop.sh"
echo ""

# Wait for user to stop
wait
