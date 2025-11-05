# Yukti FinOps - React Frontend

Modern React.js dashboard for AWS cost optimization and resource management.

## Features

🚀 **Dashboard**
- Real-time cost metrics and resource overview
- Interactive charts showing cost trends and breakdowns
- Resource status monitoring with live updates

🖥️ **Resource Management**
- Complete AWS EC2 instance inventory
- Filter by status (running, stopped, terminated)
- Emergency stop functionality for cost control

💰 **Cost Analysis**
- Historical cost trends and projections
- AI-powered optimization recommendations
- Potential savings calculations with actionable insights

## Quick Start

```bash
# Start the React UI
./start-ui.sh

# Or manually:
cd frontend
npm install
npm start
```

## Architecture

- **Framework**: React 18 with functional components and hooks
- **Routing**: React Router v6 for SPA navigation
- **Charts**: Recharts for interactive data visualization
- **Styling**: Tailwind CSS for responsive design
- **API**: Axios for backend communication with CORS support

## API Integration

The frontend connects to the Go backend API running on port 8085:

- `GET /api/v1/resources` - Fetch AWS resources
- `GET /api/v1/health` - System health check
- `POST /api/v1/emergency-stop` - Emergency instance termination

## Development

```bash
# Install dependencies
npm install

# Start development server
npm start

# Build for production
npm run build
```

## Environment Variables

- `REACT_APP_API_URL` - Backend API URL (default: http://localhost:8085)

## Browser Support

- Chrome 90+
- Firefox 88+
- Safari 14+
- Edge 90+