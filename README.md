# Stingray API

![Banana Seat](Banana_Seat.png)

Stingray is a simple, fun web API written in Go. It is designed to be open source, dependency-free, and to have zero supply-chain attack surface. Perfect for learning, hacking, or deploying as a minimal, trustworthy service.

## Features

- 🚀 **Simple**: Minimalist codebase, easy to read and extend.
- 🦀 **Go-based**: Built with Go and MySQL for reliable data storage.
- 🔒 **Secure**: Configurable database connections with environment variables.
- 🌐 **Web API**: Exposes a RESTful API for easy integration.
- 🗄️ **Database**: MySQL backend with automatic schema creation.
- 👐 **Open Source**: MIT licensed, contributions welcome!

## Current Status

- Stability 10/10
- Usefulness 0/10

## Getting Started

### Prerequisites
- [Go](https://golang.org/dl/) (any recent version)
- [MySQL](https://dev.mysql.com/downloads/) (5.7 or later)

### Build & Run

```bash
git clone https://github.com/yourusername/stingray.git
cd stingray
go mod tidy
go run .
```

The API will start on `http://localhost:8080` by default.

### Database Setup

The application uses MySQL for data storage. See [MYSQL_SETUP.md](MYSQL_SETUP.md) for detailed setup instructions.

**Quick Start**:
1. Install and start MySQL server
2. Set environment variables (optional, defaults provided)
3. Run the application - it will automatically create the database and tables

## Usage

You can interact with the API using `curl`, Postman, or any HTTP client. Example:

```bash
curl http://localhost:8080/your-endpoint
```

_Replace `/your-endpoint` with the actual endpoints provided by the API._

## Why Stingray?
- **Educational**: Great for learning Go and web APIs.
- **Trustworthy**: No hidden dependencies, no risk of supply-chain attacks.
- **Fun**: Tinker, extend, and make it your own!

## Contributing

Contributions are welcome! Please open issues or pull requests.

## License

MIT License. See [LICENSE](LICENSE) for details.

## TODOs

- [x] Add database for saving user and route data (MySQL)
- [x] Allow templates to contain other templates
- [ ] Add more API endpoints
- [ ] Write unit tests
- [ ] Improve documentation
- [ ] Add authentication/authorization
- [ ] Implement rate limiting
- [ ] Add usage examples
- [ ] Create a demo frontend