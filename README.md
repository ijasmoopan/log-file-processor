# Log Processing System

A distributed system for processing and analyzing large log files in real-time, built with Go and Next.js.

## Project Overview

This project consists of multiple microservices working together to handle log file processing, analysis, and visualization:

- **Backend Service**: REST API and WebSocket server for file management and real-time updates
- **Frontend Service**: Next.js web application for file upload and result visualization
- **Log Generator Service**: Utility service for generating test log files
- **Log Processor Service**: Service for processing and analyzing log files
- **Redis**: For real-time communication via Pub/Sub and caching
- **PostgreSQL**: For storing processing results and metadata

## Prerequisites

- Docker and Docker Compose
- Go 1.21 or later
- Node.js 18 or later
- pnpm (for frontend development)
- Redis Pub/Sub

## Project Structure

```
.
├── backend-service/            # Go backend API service
├── frontend-service/           # Next.js frontend application
├── log-generator-service/      # Log file generator utility
├── log-processor-service/      # Log processing service
├── uploads/                    # Shared volume for log files
└── docker-compose.yml          # Docker Compose configuration
```

## Services

### Backend Service

The backend service provides:
- REST API endpoints for file management
- WebSocket server for real-time updates
- Authentication using JWT Token via Supabase
- File upload handling
- Integration with Redis and PostgreSQL
- Communication between services via Redis Pub/Sub
- CORS configuration for frontend communication

### Frontend Service

The frontend service features:
- Modern UI built with Next.js
- Real-time updates via WebSocket
- File upload interface
- Results visualization
- Supabase Authentication:
  - Email/Password authentication
  - GitHub OAuth
  - Protected routes and API endpoints
  - User session management

### Log Generator Service

A utility service that generates test log files with:
- Random log levels (DEBUG, INFO, WARN, ERROR, FATAL)
- Timestamped entries
- Optional JSON payloads
- Configurable file sizes

### Log Processor Service

Processes log files and provides:
- Log analysis
- Pattern detection
- Real-time processing
- Result storage in PostgreSQL

## Getting Started

1. Clone the repository:
```bash
git clone https://github.com/ijasmoopan/log-file-processor.git
cd intucloud-task
```

2. Set up environment variables:
```bash
cp .env.example .env
cp backend-service/.env.local.example backend-service/.env.local
cp frontend-service/.env.local.example frontend-service/.env.local
cp log-processor-service/.env.local.example log-processor-service/.env.local
```

3. Start the services using Docker Compose:
```bash
docker-compose up -d
```

4. Access the application:
- Frontend: http://localhost:3000
- Backend API: http://localhost:8080
- PostgreSQL: localhost:5432
- Redis: localhost:6379

## Development

### Running Services Locally

1. Backend Service:
```bash
cd backend-service
go run main.go
```

2. Frontend Service:
```bash
cd frontend-service
pnpm install
pnpm dev
```

3. Log Processor Service:
```bash
cd log-processor-service
go run cmd/processor/main.go
```

4. Log Generator Service:
```bash
cd log-generator-service
go run main.go
```

### Environment Variables

Key environment variables needed:

#### Backend Service
- `DB_HOST`: PostgreSQL host
- `DB_PORT`: PostgreSQL port
- `DB_USER`: Database user
- `DB_PASSWORD`: Database password
- `DB_NAME`: Database name
- `REDIS_HOST`: Redis host
- `REDIS_PORT`: Redis port
- `SUPABASE_URL`: Supabase URL
- `SUPABASE_KEY`: Supabase API key

#### Frontend Service
- `NEXT_PUBLIC_BACKEND_URL`: Backend service URL
- `NEXT_PUBLIC_SUPABASE_URL`: Supabase URL
- `NEXT_PUBLIC_SUPABASE_ANON_KEY`: Supabase anonymous key

#### Log Processor Service
- `REDIS_HOST`: Redis host
- `REDIS_PORT`: Redis port

## API Endpoints

### Backend API

- `POST /api/v1/upload`: Upload log files
- `GET /api/v1/files`: List uploaded files
- `POST /api/v1/process`: Process uploaded files
- `GET /api/v1/ws`: WebSocket endpoint for real-time updates
- `GET /api/v1/results`: Get processing results
- `GET /api/v1/results/:id`: Get result by ID
- `GET /api/v1/results/filename/:filename`: Get result by filename

## Docker Support

The project includes Docker configurations for all services:

- Each service has its own `Dockerfile`
- `docker-compose.yml` orchestrates all services
- Shared volumes for file storage and database persistence
- Network isolation using Docker networks

## Contributing

1. Fork the repository
2. Create your feature branch
3. Commit your changes
4. Push to the branch
5. Create a new Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details. 