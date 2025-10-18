#!/bin/bash

# Prepare your project for cloud deployment (Render + Vercel)
# Run this script before deploying to update URLs

set -e

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${BLUE}🌐 Preparing project for cloud deployment...${NC}"

# Get backend URL from user
echo -e "${YELLOW}📝 Please provide your Render backend URL:${NC}"
echo "Example: https://monitoring-tool-backend.onrender.com"
read -p "Backend URL: " BACKEND_URL

# Get frontend URL from user
echo -e "${YELLOW}📝 Please provide your Vercel frontend URL:${NC}"
echo "Example: https://monitoring-tool-frontend.vercel.app"
read -p "Frontend URL: " FRONTEND_URL

# Validate URLs
if [[ ! $BACKEND_URL =~ ^https?:// ]]; then
    echo "❌ Invalid backend URL format"
    exit 1
fi

if [[ ! $FRONTEND_URL =~ ^https?:// ]]; then
    echo "❌ Invalid frontend URL format"
    exit 1
fi

echo -e "${BLUE}🔧 Updating configuration files...${NC}"

# Update Vercel configuration
if [ -f "frontend/monitoring-dashboard/vercel.json" ]; then
    # Extract domain from URLs
    BACKEND_DOMAIN=$(echo $BACKEND_URL | sed 's|https\?://||')
    FRONTEND_DOMAIN=$(echo $FRONTEND_URL | sed 's|https\?://||')
    
    # Update vercel.json
    sed -i.bak "s/your-backend-app.onrender.com/$BACKEND_DOMAIN/g" frontend/monitoring-dashboard/vercel.json
    echo -e "${GREEN}✅ Updated vercel.json${NC}"
fi

# Update environment example file
if [ -f "frontend/monitoring-dashboard/env.production.example" ]; then
    BACKEND_DOMAIN=$(echo $BACKEND_URL | sed 's|https\?://||')
    sed -i.bak "s/your-backend-app.onrender.com/$BACKEND_DOMAIN/g" frontend/monitoring-dashboard/env.production.example
    echo -e "${GREEN}✅ Updated env.production.example${NC}"
fi

# Update Render configuration
if [ -f "backend/render.yaml" ]; then
    FRONTEND_DOMAIN=$(echo $FRONTEND_URL | sed 's|https\?://||')
    sed -i.bak "s/your-frontend-app.vercel.app/$FRONTEND_DOMAIN/g" backend/render.yaml
    echo -e "${GREEN}✅ Updated render.yaml${NC}"
fi

echo ""
echo -e "${GREEN}🎉 Configuration updated successfully!${NC}"
echo ""
echo -e "${BLUE}📋 Next steps:${NC}"
echo "1. Push your code to GitHub"
echo "2. Deploy backend to Render: https://render.com"
echo "3. Deploy frontend to Vercel: https://vercel.com"
echo "4. Set up MongoDB Atlas: https://cloud.mongodb.com"
echo ""
echo -e "${BLUE}📖 For detailed instructions, see: RENDER_VERCEL_DEPLOYMENT.md${NC}"
echo ""
echo -e "${BLUE}🔗 Your URLs:${NC}"
echo "Backend: $BACKEND_URL"
echo "Frontend: $FRONTEND_URL"
echo ""

# Clean up backup files
find . -name "*.bak" -delete

echo -e "${GREEN}✨ Ready for cloud deployment!${NC}"
