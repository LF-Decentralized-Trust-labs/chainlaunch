# ChainLaunch

ChainLaunch is a blockchain node management platform that simplifies the deployment and management of blockchain nodes. Currently supports Hyperledger Fabric and Besu nodes.

## Features

- Deploy and manage Fabric peer and orderer nodes
- Deploy and manage Besu validator, bootnode and fullnodes 
- Service-based deployment with systemd/launchd support
- Node monitoring and log management
- Organization and identity management
- REST API for programmatic control

## Development

### Generate OpenAPI Documentation

To regenerate the Swagger/OpenAPI documentation:

```bash
swag init -g cmd/serve/serve.go -o docs --parseInternal --parseDependency --parseDepth 1 --generatedTime

```

### Generate Database Queries

To regenerate the SQL queries using sqlc:

```bash
sqlc generate
```


## Getting Started

[Add getting started section here with setup instructions]

## API Documentation

The API documentation is available at `/swagger/index.html` when running the server.
