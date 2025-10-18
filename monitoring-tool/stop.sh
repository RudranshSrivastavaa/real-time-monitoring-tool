#!/bin/bash

# Simple stop script for Real-time Monitoring Tool

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🛑 Stopping Real-time Monitoring Tool...${NC}"

# Stop backend
if [ -f "backend.pid" ]; then
    BACKEND_PID=$(cat backend.pid)
    if kill -0 $BACKEND_PID 2>/dev/null; then
        echo -e "${BLUE}🔧 Stopping Backend (PID: $BACKEND_PID)...${NC}"
        kill $BACKEND_PID
        rm backend.pid
        echo -e "${GREEN}✅ Backend stopped${NC}"
    else
        echo -e "${GREEN}✅ Backend was not running${NC}"
        rm backend.pid
    fi
else
    echo -e "${GREEN}✅ Backend was not running${NC}"
fi

# Stop frontend
if [ -f "frontend.pid" ]; then
    FRONTEND_PID=$(cat frontend.pid)
    if kill -0 $FRONTEND_PID 2>/dev/null; then
        echo -e "${BLUE}⚛️  Stopping Frontend (PID: $FRONTEND_PID)...${NC}"
        kill $FRONTEND_PID
        rm frontend.pid
        echo -e "${GREEN}✅ Frontend stopped${NC}"
    else
        echo -e "${GREEN}✅ Frontend was not running${NC}"
        rm frontend.pid
    fi
else
    echo -e "${GREEN}✅ Frontend was not running${NC}"
fi

# Stop MongoDB (optional - comment out if you want to keep it running)
echo -e "${BLUE}📊 Stopping MongoDB...${NC}"
if pgrep -x "mongod" > /dev/null; then
    pkill mongod
    echo -e "${GREEN}✅ MongoDB stopped${NC}"
else
    echo -e "${GREEN}✅ MongoDB was not running${NC}"
fi

echo ""
echo -e "${GREEN}🎉 All services stopped successfully!${NC}"
