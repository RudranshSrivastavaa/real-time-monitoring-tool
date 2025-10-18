# Simple Deployment Guide

This is a simple deployment guide for your college project - no Docker, no complex configurations!

## 📋 Prerequisites

Before you start, make sure you have these installed:

1. **Go 1.21 or later**
   - Download from: https://golang.org/dl/
   - Verify: `go version`

2. **Node.js 18 or later**
   - Download from: https://nodejs.org/
   - Verify: `node --version` and `npm --version`

3. **MongoDB**
   - Download from: https://www.mongodb.com/try/download/community
   - Verify: `mongod --version`

## 🚀 Step-by-Step Setup

### 1. Clone and Setup

```bash
# Clone your repository
git clone <your-repository-url>
cd monitoring-tool

# Run the setup script
./setup.sh
```

The setup script will:
- Check if all prerequisites are installed
- Create necessary directories
- Install Go dependencies
- Install npm dependencies
- Create a `.env` file with default settings

### 2. Start the Application

```bash
# Start everything (MongoDB, Backend, Frontend)
./start.sh
```

This will:
- Start MongoDB on port 27017
- Start Go backend on port 8080
- Start React frontend on port 3000

### 3. Access Your Application

- **Frontend Dashboard**: http://localhost:3000
- **Backend API**: http://localhost:8080/api/v1
- **Health Check**: http://localhost:8080/api/v1/health
- **WebSocket**: ws://localhost:8080/ws

### 4. Stop the Application

```bash
# Stop everything
./stop.sh
```

## 🔧 Manual Setup (Alternative)

If you prefer to set up manually:

### Backend Setup

```bash
cd backend

# Install Go dependencies
go mod download

# Create .env file
cat > .env << EOF
MONGODB_URI=mongodb://localhost:27017
DATABASE_NAME=realtime_monitor
PORT=8080
HOST=localhost
ENVIRONMENT=development
ALLOWED_ORIGINS=http://localhost:3000
EOF

# Start backend
go run main.go
```

### Frontend Setup

```bash
cd frontend/monitoring-dashboard

# Install dependencies
npm install

# Start frontend
npm start
```

### MongoDB Setup

```bash
# Create data directory
mkdir -p data/db

# Start MongoDB
mongod --dbpath ./data/db
```

## 📊 Using the Application

1. **Open the Dashboard**: Go to http://localhost:3000
2. **Add a Monitor**: Click "Add Monitor" and enter:
   - Name: Any name you want
   - URL: The website/API you want to monitor (e.g., https://google.com)
   - Interval: How often to check (in seconds)
3. **View Results**: See real-time status updates, response times, and charts

## 🛠️ Troubleshooting

### MongoDB Connection Issues

If you see MongoDB connection errors:

1. Make sure MongoDB is running:
   ```bash
   # Check if MongoDB is running
   pgrep mongod
   ```

2. If not running, start it:
   ```bash
   mongod --dbpath ./data/db
   ```

### Backend Issues

If the backend won't start:

1. Check if port 8080 is free:
   ```bash
   lsof -i :8080
   ```

2. Try a different port by editing `.env`:
   ```
   PORT=8081
   ```

### Frontend Issues

If the frontend won't start:

1. Check if port 3000 is free:
   ```bash
   lsof -i :3000
   ```

2. Clear npm cache:
   ```bash
   npm cache clean --force
   rm -rf node_modules
   npm install
   ```

## 📝 Project Structure

```
monitoring-tool/
├── backend/                 # Go backend
│   ├── main.go             # Main application file
│   ├── handlers/           # API handlers
│   ├── services/           # Business logic
│   ├── models/             # Data models
│   ├── database/           # Database connection
│   └── config/             # Configuration
├── frontend/monitoring-dashboard/  # React frontend
│   ├── src/
│   │   ├── components/     # React components
│   │   └── App.js          # Main app component
│   └── package.json        # Dependencies
├── start.sh                # Start script
├── stop.sh                 # Stop script
├── setup.sh                # Setup script
└── README.md               # This file
```

## 🎯 College Project Tips

1. **Documentation**: Update the README with your specific details
2. **Customization**: Modify the UI colors, add your college logo
3. **Features**: Add more monitoring features like email alerts
4. **Database**: Show different types of data visualization
5. **Testing**: Add some unit tests for your Go code

## 📞 Need Help?

- Check the logs in the terminal where you ran `./start.sh`
- Make sure all ports (3000, 8080, 27017) are free
- Verify all prerequisites are installed correctly
- Check your internet connection for downloading dependencies

Good luck with your college project! 🎓
