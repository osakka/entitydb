# EntityDB Admin Setup Guide

EntityDB has been configured to work with default admin credentials.

## Default Credentials
- **Username**: `admin`
- **Password**: `admin`

## Setup Process

1. **Start the server**:
   ```bash
   /opt/entitydb/bin/entitydbd.sh start
   ```

2. **Initialize admin user** (if needed):
   ```bash
   /opt/entitydb/bin/init.sh
   ```

## Testing Login

```bash
# Test login with curl
curl -X POST http://localhost:8085/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin"}'
```

## Access Points
- **Dashboard**: http://localhost:8085/
- **API Documentation**: http://localhost:8085/swagger/
- **API Base URL**: http://localhost:8085/api/v1/

## Key Scripts
- `/opt/entitydb/bin/entitydbd.sh` - Server control script (start/stop/restart/status)
- `/opt/entitydb/bin/init.sh` - Database initialization script (creates admin/admin user)
- `/opt/entitydb/init_working_admin.sh` - Manual admin creation script

The system is now ready with admin/admin credentials!