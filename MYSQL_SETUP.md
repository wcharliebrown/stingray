# MySQL Setup for Sting Ray

This application has been converted from SQLite to MySQL. Follow these steps to set up the MySQL database.

## Prerequisites

1. **MySQL Server**: Install MySQL Server on your system
   - **macOS**: `brew install mysql`
   - **Ubuntu/Debian**: `sudo apt-get install mysql-server`
   - **Windows**: Download from [MySQL Downloads](https://dev.mysql.com/downloads/mysql/)

2. **Start MySQL Service**:
   - **macOS**: `brew services start mysql`
   - **Ubuntu/Debian**: `sudo systemctl start mysql`
   - **Windows**: Start MySQL service from Services

## Docker Setup (Alternative)

If you prefer to run MySQL in a Docker container instead of installing it locally:

### Prerequisites
- **Docker**: Install Docker Desktop or Docker Engine
  - **macOS**: Download from [Docker Desktop](https://www.docker.com/products/docker-desktop/)
  - **Ubuntu/Debian**: `sudo apt-get install docker.io`
  - **Windows**: Download from [Docker Desktop](https://www.docker.com/products/docker-desktop/)

### Running MySQL with Docker

1. **Pull MySQL Image**:
   ```bash
   docker pull mysql:8.0
   ```

2. **Run MySQL Container**:
   ```bash
   docker run --name stingray-mysql \
     -e MYSQL_ROOT_PASSWORD=password \
     -e MYSQL_DATABASE=stingray \
     -p 3306:3306 \
     -d mysql:8.0
   ```

3. **Verify Container is Running**:
   ```bash
   docker ps
   ```

4. **Connect to MySQL Container** (Optional):
   ```bash
   docker exec -it stingray-mysql mysql -u root -p
   ```

### Docker Compose (Recommended)

For easier management, create a `docker-compose.yml` file:

```yaml
version: '3.8'
services:
  mysql:
    image: mysql:8.0
    container_name: stingray-mysql
    restart: unless-stopped
    environment:
      MYSQL_ROOT_PASSWORD: password
      MYSQL_DATABASE: stingray
    ports:
      - "3306:3306"
    volumes:
      - mysql_data:/var/lib/mysql
    command: --default-authentication-plugin=mysql_native_password

volumes:
  mysql_data:
```

Then run:
```bash
docker-compose up -d
```

### Docker Environment Variables

When using Docker, update your environment variables:

```bash
export MYSQL_HOST=localhost
export MYSQL_PORT=3306
export MYSQL_USER=root
export MYSQL_PASSWORD=password
export MYSQL_DATABASE=stingray
```

### Docker Management Commands

```bash
# Start the container
docker start stingray-mysql

# Stop the container
docker stop stingray-mysql

# Remove the container
docker rm stingray-mysql

# View logs
docker logs stingray-mysql

# Using docker-compose
docker-compose up -d    # Start
docker-compose down     # Stop and remove
docker-compose logs     # View logs
```

## Configuration

The application uses environment variables for MySQL configuration. You can set these or use the defaults:

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `MYSQL_HOST` | `localhost` | MySQL server hostname |
| `MYSQL_PORT` | `3306` | MySQL server port |
| `MYSQL_USER` | `root` | MySQL username |
| `MYSQL_PASSWORD` | `password` | MySQL password |
| `MYSQL_DATABASE` | `stingray` | Database name |

### Example Environment Setup

```bash
# Set environment variables
export MYSQL_HOST=localhost
export MYSQL_PORT=3306
export MYSQL_USER=root
export MYSQL_PASSWORD=your_password
export MYSQL_DATABASE=stingray
```

## Database Setup

### Option 1: Automatic Setup (Recommended)

The application will automatically create the database and tables if they don't exist. Just run:

```bash
go run .
```

### Option 2: Manual Setup

1. **Connect to MySQL**:
   ```bash
   mysql -u root -p
   ```

2. **Create Database**:
   ```sql
   CREATE DATABASE stingray CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
   ```

3. **Create User** (Optional):
   ```sql
   CREATE USER 'stingray_user'@'localhost' IDENTIFIED BY 'your_password';
   GRANT ALL PRIVILEGES ON stingray.* TO 'stingray_user'@'localhost';
   FLUSH PRIVILEGES;
   ```

## Running the Application

1. **Install Dependencies**:
   ```bash
   go mod tidy
   ```

2. **Run the Application**:
   ```bash
   go run .
   ```

3. **Access the Application**:
   - Open your browser to `http://localhost:8080`
   - The application will automatically create the database and tables on first run

## Database Schema

The application creates several tables with the following structure:

```sql
-- Pages table
CREATE TABLE _page (
    id INT AUTO_INCREMENT PRIMARY KEY,
    slug VARCHAR(255) UNIQUE NOT NULL,
    title VARCHAR(255) NOT NULL,
    meta_description TEXT,
    header TEXT,
    navigation TEXT,
    main_content TEXT,
    sidebar TEXT,
    footer TEXT,
    css_class VARCHAR(255),
    scripts TEXT,
    template VARCHAR(100) DEFAULT 'default',
    created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Users table
CREATE TABLE _user (
    id INT AUTO_INCREMENT PRIMARY KEY,
    username VARCHAR(255) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Groups table
CREATE TABLE _group (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) UNIQUE NOT NULL,
    description TEXT,
    created TIMESTAMP DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- User-Groups relationship table
CREATE TABLE _user_and_group (
    id INT AUTO_INCREMENT PRIMARY KEY,
    user_id INT NOT NULL,
    group_id INT NOT NULL,
    created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY unique_user_group (user_id, group_id),
    FOREIGN KEY (user_id) REFERENCES _user(id) ON DELETE CASCADE,
    FOREIGN KEY (group_id) REFERENCES _group(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Sessions table
CREATE TABLE _session (
    id INT AUTO_INCREMENT PRIMARY KEY,
    session_id VARCHAR(255) UNIQUE NOT NULL,
    user_id INT NOT NULL,
    username VARCHAR(255) NOT NULL,
    created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    FOREIGN KEY (user_id) REFERENCES _user(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

## Troubleshooting

### Connection Issues

1. **Check MySQL Service**:
   ```bash
   # macOS
   brew services list | grep mysql
   
   # Ubuntu/Debian
   sudo systemctl status mysql
   ```

2. **Test MySQL Connection**:
   ```bash
   mysql -u root -p -h localhost -P 3306
   ```

3. **Check Firewall**: Ensure port 3306 is not blocked

### Permission Issues

1. **Reset MySQL Root Password**:
   ```bash
   sudo mysql_secure_installation
   ```

2. **Create New User**:
   ```sql
   CREATE USER 'stingray_user'@'localhost' IDENTIFIED BY 'your_password';
   GRANT ALL PRIVILEGES ON stingray.* TO 'stingray_user'@'localhost';
   FLUSH PRIVILEGES;
   ```

### Migration from SQLite

If you have existing data in SQLite, you can migrate it:

1. **Export SQLite Data**:
   ```bash
   sqlite3 stingray.db ".dump" > stingray_dump.sql
   ```

2. **Convert and Import** (Manual process required due to syntax differences)

## Security Notes

- Change the default password in production
- Use environment variables for sensitive configuration
- Consider using a dedicated MySQL user instead of root
- Enable SSL/TLS for production deployments

## Dependencies

The application now uses:
- `github.com/go-sql-driver/mysql v1.9.3` - MySQL driver for Go
- `filippo.io/edwards25519 v1.1.0` - Required by MySQL driver 