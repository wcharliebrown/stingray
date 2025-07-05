# Stingray API

![Banana Seat](Banana_Seat.png)

Stingray is a simple, fun web API written in Go. It is designed to be open source, dependency-free, and to have zero supply-chain attack surface. Perfect for learning, hacking, or deploying as a minimal, trustworthy service.

## Features

- 🚀 **Simple**: Minimalist codebase, easy to read and extend.
- 🦀 **No Dependencies**: Pure Go, no third-party packages.
- 🔒 **Secure**: No supply-chain attack surface.
- 🌐 **Web API**: Exposes a RESTful API for easy integration.
- 👐 **Open Source**: MIT licensed, contributions welcome!

## Current Status

- Stability 10/10
- Usefulness 0/10

## Getting Started

### Prerequisites
- [Go](https://golang.org/dl/) (any recent version)

### Build & Run

```bash
git clone https://github.com/yourusername/stingray.git
cd stingray
go run .
```

The API will start on `http://localhost:8080` by default.

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

- [ ] Add more API endpoints
- [ ] Write unit tests
- [ ] Improve documentation
- [ ] Add database for saving user and route data
- [ ] Add authentication/authorization
- [ ] Implement rate limiting
- [ ] Add usage examples
- [ ] Create a demo frontend