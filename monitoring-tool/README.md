# Real-time Monitoring Tool

A simple real-time monitoring solution for APIs and web services - Perfect for college projects!

## ✨ Features

- **Real-time Monitoring**: Monitor multiple endpoints with configurable intervals
- **WebSocket Updates**: Live updates via WebSocket connections
- **Response Time Tracking**: Track response times and HTTP status codes
- **Modern Dashboard**: Clean React-based dashboard
- **RESTful API**: Simple REST API for monitor management
- **MongoDB**: Persistent data storage

## 🏗️ Simple Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   React         │    │   Go Backend    │    │   MongoDB       │
│   Frontend      │◄──►│   API Server    │◄──►│   Database      │
│   (Port 3000)   │    │   (Port 8080)   │    │   (Port 27017)  │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │
         │              ┌─────────────────┐
         └──────────────►│   WebSocket     │
                        │   Hub           │
                        └─────────────────┘
```

## 🚀 Quick Start

### Prerequisites

- Go 1.21 or later
- Node.js 18 or later  
- MongoDB

### Simple Setup & Run

```bash
# 1. Clone the repository
git clone <your-repository-url>
cd monitoring-tool

# 2. Run setup script (installs dependencies)
./setup.sh

# 3. Start the application
./start.sh

# 4. Access the application
# Frontend: http://localhost:3000
# Backend API: http://localhost:8080/api/v1

# 5. Stop the application
./stop.sh
```

That's it! No Docker, no complex configurations - just simple scripts to get you started.

## 🌐 Deploy to Cloud (Recommended for College Projects)

For a professional deployment that's perfect for college presentations:

### Option 1: Render + Vercel (Free Hosting)
- **Backend**: Deploy to Render (free Go hosting)
- **Frontend**: Deploy to Vercel (free React hosting)  
- **Database**: Use MongoDB Atlas (free cloud database)

See [RENDER_VERCEL_DEPLOYMENT.md](./RENDER_VERCEL_DEPLOYMENT.md) for detailed steps.

### Option 2: Local Development
Use the simple scripts above for local development and testing.

## 📖 API Documentation

### Endpoints

#### Monitors
- `GET /api/v1/monitors` - List all monitors
- `POST /api/v1/monitors` - Create a new monitor
- `DELETE /api/v1/monitors/:id` - Delete a monitor
- `GET /api/v1/monitors/:id/metrics` - Get monitor metrics

#### Dashboard
- `GET /api/v1/dashboard/stats` - Get dashboard statistics
- `GET /api/v1/health` - Health check endpoint

#### WebSocket
- `WS /ws` - WebSocket connection for real-time updates

### Example API Usage

```bash
# Create a monitor
curl -X POST http://localhost:8080/api/v1/monitors \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Google API",
    "url": "https://www.google.com",
    "interval": 30
  }'

# Get all monitors
curl http://localhost:8080/api/v1/monitors

# Get dashboard stats
curl http://localhost:8080/api/v1/dashboard/stats
```

## 🔧 Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `MONGODB_URI` | MongoDB connection string | `mongodb://localhost:27017` |
| `DATABASE_NAME` | Database name | `realtime_monitor` |
| `PORT` | Backend port | `8080` |
| `ENVIRONMENT` | Environment mode | `debug` |
| `DEFAULT_INTERVAL` | Default monitoring interval (seconds) | `30` |
| `MAX_CONCURRENT_CHECKS` | Maximum concurrent checks | `100` |

### Monitor Configuration

When creating a monitor, you can specify:

- **Name**: Display name for the monitor
- **URL**: The endpoint to monitor
- **Interval**: Check interval in seconds (minimum: 5 seconds)

## 🛠️ Development

### Backend (Go)

```bash
cd backend

# Install dependencies
go mod download

# Run tests
go test ./...

# Run with hot reload (requires air)
air

# Build binary
go build -o monitoring-tool main.go
```

### Frontend (React)

```bash
cd frontend/monitoring-dashboard

# Install dependencies
npm install

# Start development server
npm start

# Build for production
npm run build

# Run tests
npm test
```

## 📊 Monitoring Features

### Real-time Metrics

- **Response Time**: Track average, min, max response times
- **Status Codes**: Monitor HTTP status codes
- **Uptime**: Calculate uptime percentages
- **Trend Analysis**: Historical data visualization

### Dashboard Views

- **Overview**: Summary statistics and quick status
- **Individual Monitors**: Detailed metrics per monitor
- **Response Time Charts**: Visual trend analysis
- **Real-time Updates**: Live status changes

## 🔒 Security

### Production Security

- Environment-based configuration
- CORS protection
- Input validation
- SQL injection protection (MongoDB)
- Rate limiting ready
- HTTPS support

### Best Practices

1. Use strong passwords for MongoDB
2. Enable HTTPS in production
3. Configure proper CORS origins
4. Regular security updates
5. Monitor resource usage

## 🐳 Docker

### Development

```bash
# Start only MongoDB
docker-compose up -d mongodb

# Start all services
docker-compose up -d
```

### Production

```bash
# Build and start production services
docker-compose -f docker-compose.prod.yml up -d --build
```

## 📈 Scaling

### Horizontal Scaling

The application supports horizontal scaling:

- **Backend**: Multiple instances behind load balancer
- **Frontend**: Multiple instances with sticky sessions
- **Database**: MongoDB replica sets
- **WebSocket**: Redis for WebSocket scaling

### Performance Tuning

- Adjust `MAX_CONCURRENT_CHECKS` based on server capacity
- Configure `METRICS_RETENTION_DAYS` for storage optimization
- Use connection pooling for database
- Implement caching for frequently accessed data

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.

## 🆘 Support

- **Documentation**: See [DEPLOYMENT.md](./DEPLOYMENT.md) for deployment help
- **Issues**: Report bugs and feature requests via GitHub issues
- **Discussions**: Use GitHub discussions for questions

## 🔄 Changelog

### Version 1.0.0
- Initial release
- Real-time monitoring dashboard
- RESTful API
- WebSocket support
- Docker deployment
- Production-ready configuration

## 🏆 Acknowledgments

- Built with Go and React
- Uses MongoDB for data persistence
- WebSocket implementation with Gorilla WebSocket
- UI components with Lucide React icons
- Charts with Recharts library
