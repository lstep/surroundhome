# SurroundHome

A modular home automation system written in Go that enables integration of various services and automation rules through a flexible architecture.

## Overview

SurroundHome is designed to be a flexible and extensible home automation platform that can handle various types of inputs and trigger corresponding actions based on configurable rules. The system uses NATS for communication between services, making it highly scalable and loosely coupled.

## Components

### REST Proxy Service
A service that provides a REST API interface for interacting with the system. It acts as an entry point for HTTP-based integrations and forwards requests to appropriate services through NATS.

### Obsidian New Discoveries Service
A service that integrates with Obsidian note-taking application to automatically add new discoveries and links to your daily notes through NATS messages.

## Features

- Modular architecture allowing easy addition of new services
- REST API interface for external integrations
- NATS-based communication between services
- Flexible rule-based automation system
- Support for various input types (REST, planned: CLI, web interface, clipboard)

## Getting Started

### Prerequisites

- NATS server
- Any specific requirements for individual services (e.g., Obsidian for obsidian-new-discoveries)

### Installation

1. Clone the repository
2. Install dependencies:
   ```bash
   go mod download
   ```
3. Build and run desired services

## Roadmap

- [ ] Enhanced rule engine with DSL support for defining automation rules
- [ ] Additional input interfaces (CLI, web interface, clipboard)
- [ ] Optional authentication system
- [ ] Package generation for various platforms (DMG, DEB, etc.)
- [ ] Custom NATS proxy implementation for enhanced response handling
- [ ] Service-specific documentation
- [ ] Systemd service files for each component

## Contributing

Contributions are welcome! Each service should include:
- A detailed README
- Systemd service file
- Dockerfile for containerized deployment

## License

See the [LICENSE](LICENSE) file for details.
