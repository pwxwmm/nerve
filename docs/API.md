# Nerve API Documentation (English)

> **Note**: This is the English version of the API documentation. For a more detailed and up-to-date Chinese version, please see [API_REFERENCE.md](API_REFERENCE.md).

## Base URL

```
http://nerve-center:8090/api
```

## Authentication

All endpoints require Bearer token authentication:

```
Authorization: Bearer <token>
```

## Quick Reference

For detailed API documentation, please refer to the [API Reference Guide](API_REFERENCE.md) which includes:

- Complete endpoint descriptions
- Request/response examples
- Error handling
- Authentication details
- Usage examples

## Available Endpoints

### Agent Management
- `POST /api/agents/register` - Register a new agent
- `GET /api/agents` - List all agents
- `GET /api/agents/{id}` - Get agent details
- `PUT /api/agents/{id}/status` - Update agent status
- `POST /api/agents/{id}/heartbeat` - Send heartbeat
- `DELETE /api/agents/{id}` - Delete agent

### System
- `GET /api/health` - Health check
- `GET /api/v1/system/stats` - System statistics

### Installation
- `GET /api/install?token=<token>` - Get installation script
- `GET /api/download?token=<token>` - Download agent binary

## See Also

- [API Reference (中文)](API_REFERENCE.md) - Detailed Chinese API documentation
- [Architecture Documentation](ARCHITECTURE.md)
- [Deployment Guide](DEPLOYMENT.md)
