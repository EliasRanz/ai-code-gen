# Database Schema Documentation

This directory contains the database schema and migration scripts for the AI UI Generator application.

## Overview

The database schema supports a comprehensive AI-powered UI generation platform with the following core entities:

- **Users**: User accounts and authentication
- **Projects**: User projects containing generated UI components
- **Chat Sessions**: Conversation sessions with the AI
- **Chat Messages**: Individual messages within chat sessions
- **UI Generations**: Tracked AI-generated UI components
- **User Settings**: User preferences and configuration
- **API Keys**: User API key management

## Database Tables

### Core Tables

#### `users`
- User account information
- Authentication and profile data
- Role-based access control
- Email verification status

#### `projects`
- User projects containing UI components
- Project metadata and configuration
- Status tracking (draft, active, completed, archived)
- Public/private visibility settings

#### `chat_sessions`
- AI conversation sessions
- Session context and metadata
- Message count tracking
- Session expiry management

#### `chat_messages`
- Individual messages in chat sessions
- Support for different message types (text, code, images)
- Message threading with parent-child relationships
- Token usage and model tracking

### Extended Tables

#### `ui_generations`
- Tracks AI-generated UI components
- Links to originating chat messages
- Generated code and assets storage
- Generation status and error tracking

#### `user_settings`
- User preferences and configuration
- Theme, language, and timezone settings
- AI model preferences
- Notification settings

#### `user_api_keys`
- API key management for programmatic access
- Permission-based access control
- Rate limiting configuration
- Usage tracking

## Migration Files

| File | Description |
|------|-------------|
| `001_create_users_table.sql` | Core user accounts table |
| `002_create_projects_table.sql` | Project management table |
| `003_create_chat_sessions_table.sql` | Chat session tracking |
| `004_create_chat_messages_table.sql` | Message storage and threading |
| `005_create_ui_generations_table.sql` | AI generation tracking |
| `006_create_user_settings_and_api_keys.sql` | User preferences and API management |

## Setup Instructions

### Prerequisites
- PostgreSQL 13+ 
- `psql` command-line tool
- Database user with CREATE privileges

### Quick Setup

1. **Run the initialization script:**
   ```bash
   ./migrations/init_db.sh
   ```

2. **Or manual setup:**
   ```bash
   # Create database
   createdb ai_ui_generator
   
   # Run migrations in order
   psql -d ai_ui_generator -f migrations/001_create_users_table.sql
   psql -d ai_ui_generator -f migrations/002_create_projects_table.sql
   psql -d ai_ui_generator -f migrations/003_create_chat_sessions_table.sql
   psql -d ai_ui_generator -f migrations/004_create_chat_messages_table.sql
   psql -d ai_ui_generator -f migrations/005_create_ui_generations_table.sql
   psql -d ai_ui_generator -f migrations/006_create_user_settings_and_api_keys.sql
   ```

### Environment Variables

Configure the following environment variables for database connection:

```bash
export DB_HOST=localhost
export DB_PORT=5432
export DB_NAME=ai_ui_generator
export DB_USER=postgres
export DB_PASSWORD=your_password
```

## Schema Features

### Auto-updating Timestamps
All tables include automatic `updated_at` timestamp updates via PostgreSQL triggers.

### UUID Primary Keys
All tables use UUID primary keys for better scalability and security.

### JSONB Storage
Flexible metadata storage using PostgreSQL's JSONB type for:
- Project configuration
- Chat session context
- UI generation metadata
- User preferences

### Indexes
Comprehensive indexing strategy for:
- Foreign key relationships
- Query performance optimization
- JSONB field searches
- Time-based queries

### Constraints and Relationships
- Foreign key constraints with appropriate cascade behaviors
- Unique constraints for email addresses and API keys
- Check constraints for data validation

## Performance Considerations

- **Partitioning**: Consider partitioning large tables like `chat_messages` by date
- **Archiving**: Implement archiving strategy for old chat sessions
- **Monitoring**: Monitor query performance and index usage
- **Connection Pooling**: Use connection pooling in production

## Security

- **API Keys**: Stored as hashes, never plain text
- **Permissions**: Role-based access control via user roles array
- **Rate Limiting**: Built-in rate limiting for API key usage
- **Audit Trail**: Comprehensive timestamp tracking

## Rollback

Each migration includes rollback instructions in the `-- +migrate Down` section. To rollback:

```sql
-- Example rollback for last migration
-- Run the Down section of the migration file
```

## Next Steps

After setting up the database schema:

1. Implement repository pattern in Go
2. Add database connection pooling
3. Implement data access layer
4. Add database migration management
5. Set up monitoring and backup strategies
