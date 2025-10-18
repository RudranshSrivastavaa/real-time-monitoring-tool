# Deploy to Render (Backend) + Vercel (Frontend)

This guide will help you deploy your Real-time Monitoring Tool with:
- **Backend**: Deployed on Render (free tier)
- **Frontend**: Deployed on Vercel (free tier)
- **Database**: MongoDB Atlas (free tier)

Perfect for your college project! 🎓

## 📋 Prerequisites

1. **GitHub Account** - Your code should be on GitHub
2. **Render Account** - Sign up at https://render.com (free)
3. **Vercel Account** - Sign up at https://vercel.com (free)
4. **MongoDB Atlas Account** - Sign up at https://cloud.mongodb.com (free)

## 🗄️ Step 1: Setup MongoDB Atlas (Free Database)

### 1.1 Create MongoDB Atlas Account
1. Go to https://cloud.mongodb.com
2. Sign up for free
3. Create a new project (e.g., "Monitoring Tool")

### 1.2 Create Database Cluster
1. Click "Build a Database"
2. Choose "FREE" tier (M0)
3. Select a cloud provider and region
4. Name your cluster (e.g., "monitoring-cluster")
5. Click "Create Cluster"

### 1.3 Create Database User
1. Go to "Database Access" in the left menu
2. Click "Add New Database User"
3. Choose "Password" authentication
4. Create username and password (save these!)
5. Set privileges to "Read and write to any database"
6. Click "Add User"

### 1.4 Whitelist IP Addresses
1. Go to "Network Access" in the left menu
2. Click "Add IP Address"
3. Click "Allow Access from Anywhere" (0.0.0.0/0)
4. Click "Confirm"

### 1.5 Get Connection String
1. Go to "Clusters" and click "Connect"
2. Choose "Connect your application"
3. Copy the connection string (looks like: `mongodb+srv://username:password@cluster.mongodb.net/`)
4. **Save this connection string - you'll need it for Render!**

## 🚀 Step 2: Deploy Backend to Render

### 2.1 Prepare Your Repository
Make sure your code is pushed to GitHub with the backend files in the `backend/` folder.

### 2.2 Create Render Service
1. Go to https://render.com and sign in
2. Click "New +" → "Web Service"
3. Connect your GitHub repository
4. Choose your repository

### 2.3 Configure Render Service
Fill in these details:

**Basic Settings:**
- **Name**: `monitoring-tool-backend`
- **Environment**: `Go`
- **Region**: Choose closest to you
- **Branch**: `main` (or your main branch)
- **Root Directory**: `backend`
- **Build Command**: `go build -o monitoring-tool main.go`
- **Start Command**: `./monitoring-tool`

**Environment Variables:**
Click "Add Environment Variable" and add these:

```
MONGODB_URI = mongodb+srv://username:password@cluster.mongodb.net/realtime_monitor?retryWrites=true&w=majority
DATABASE_NAME = realtime_monitor
PORT = 8080
HOST = 0.0.0.0
ENVIRONMENT = production
ALLOWED_ORIGINS = https://your-frontend-app.vercel.app
DEFAULT_INTERVAL = 30
DEFAULT_TIMEOUT = 10
MAX_CONCURRENT_CHECKS = 100
METRICS_RETENTION_DAYS = 30
```

**Important**: Replace `username:password@cluster.mongodb.net` with your actual MongoDB Atlas credentials!

### 2.4 Deploy
1. Click "Create Web Service"
2. Wait for deployment (5-10 minutes)
3. Your backend will be available at: `https://your-app-name.onrender.com`

**Save your backend URL!** You'll need it for the frontend.

## ⚛️ Step 3: Deploy Frontend to Vercel

### 3.1 Prepare Frontend
1. Update the environment variables in your frontend
2. Edit `frontend/monitoring-dashboard/vercel.json`
3. Replace `your-backend-app.onrender.com` with your actual Render backend URL

### 3.2 Create Vercel Project
1. Go to https://vercel.com and sign in
2. Click "New Project"
3. Import your GitHub repository
4. Choose your repository

### 3.3 Configure Vercel Project
**Project Settings:**
- **Framework Preset**: `Create React App`
- **Root Directory**: `frontend/monitoring-dashboard`
- **Build Command**: `npm run build`
- **Output Directory**: `build`

**Environment Variables:**
Add these environment variables:

```
REACT_APP_API_URL = https://your-backend-app.onrender.com/api
REACT_APP_WS_URL = wss://your-backend-app.onrender.com/ws
```

Replace `your-backend-app.onrender.com` with your actual Render backend URL!

### 3.4 Deploy
1. Click "Deploy"
2. Wait for deployment (2-3 minutes)
3. Your frontend will be available at: `https://your-app-name.vercel.app`

## 🔧 Step 4: Update CORS Settings

After both deployments are complete:

1. Go to your Render dashboard
2. Find your backend service
3. Go to "Environment" tab
4. Update `ALLOWED_ORIGINS` to include your Vercel frontend URL:
   ```
   https://your-frontend-app.vercel.app
   ```
5. Redeploy the backend

## 🎉 Step 5: Test Your Deployment

1. **Frontend**: Visit your Vercel URL
2. **Backend API**: Test `https://your-backend.onrender.com/api/v1/health`
3. **Add a Monitor**: Try adding a monitor in your frontend
4. **Check Real-time Updates**: WebSocket should work automatically

## 📊 Your Live Application

- **Frontend**: `https://your-app-name.vercel.app`
- **Backend API**: `https://your-app-name.onrender.com/api/v1`
- **Health Check**: `https://your-app-name.onrender.com/api/v1/health`
- **WebSocket**: `wss://your-app-name.onrender.com/ws`

## 🛠️ Troubleshooting

### Backend Issues
1. **MongoDB Connection Failed**: Check your MongoDB Atlas connection string
2. **Build Failed**: Make sure your Go code compiles locally
3. **Environment Variables**: Double-check all environment variables in Render

### Frontend Issues
1. **API Connection Failed**: Check if backend URL is correct
2. **WebSocket Not Working**: Make sure backend allows your frontend domain
3. **Build Failed**: Make sure your React app builds locally

### Common Solutions
1. **CORS Errors**: Update `ALLOWED_ORIGINS` in Render
2. **Environment Variables**: Make sure they're set correctly in both platforms
3. **Database Issues**: Check MongoDB Atlas connection and user permissions

## 💡 Pro Tips for College Project

1. **Custom Domain**: You can add a custom domain to both Render and Vercel
2. **SSL Certificates**: Both platforms provide free SSL certificates
3. **Monitoring**: Render provides basic monitoring in the free tier
4. **Logs**: Check logs in both Render and Vercel dashboards for debugging
5. **Automatic Deploys**: Both platforms auto-deploy when you push to GitHub

## 📱 Demo Your Project

Perfect for college presentations:
- Show the live website
- Demonstrate real-time monitoring
- Explain the architecture
- Show the code on GitHub
- Mention the free hosting setup

## 🔄 Making Updates

1. **Code Changes**: Push to GitHub
2. **Backend**: Render auto-deploys
3. **Frontend**: Vercel auto-deploys
4. **Environment Variables**: Update in respective dashboards if needed

Your college project is now live on the internet! 🌐

## 📞 Need Help?

- **Render Support**: https://render.com/docs
- **Vercel Support**: https://vercel.com/docs
- **MongoDB Atlas**: https://docs.atlas.mongodb.com

Good luck with your project! 🎓✨
