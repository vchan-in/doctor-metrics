# Doctor Metrics

A lightweight application for real-time Docker container metrics (CPU, memory, network, and filesystem). It offers REST APIs to retrieve metrics for all running containers or specific ones by name or ID.

## Table of Contents

- [Features](#features)
- [Installation](#installation)
- [Usage](#usage)
- [API Endpoints](#api-endpoints)
- [Development](#development)
- [Deployment](#deployment)
- [Environment Variables](#environment-variables)
- [License](#license)

## Features

- Retrieve metrics for all running Docker containers.
- Retrieve metrics for a specific container by name or ID.
- Basic authentication middleware.
- CORS support.
- Rate limiting.

## Installation

### From Release

1. Download the latest release for your operating system from the [GitHub release section](https://github.com/vchan-in/doctor-metrics/releases).

2. Extract the downloaded file.

3. Copy the example environment file and modify it as needed:
    ```sh
    cp env.example .env
    ```

## Usage

1. Run the application:
    ```sh
    ./dh
    ```

2. The server will start on the port specified in the `.env` file (default is `9095`).

## API Endpoints

- `GET /` - Root endpoint to check the API status.
- `GET /api/metrics` - Retrieve metrics for all running containers.
- `GET /api/metrics/:containerName` - Retrieve metrics for a specific container by name.
- `GET /api/metrics/:containerID` - Retrieve metrics for a specific container by ID.

## Authentication

The application uses basic authentication to secure the API endpoints. You need to set the `DM_USERNAME` and `DM_PASSWORD` environment variables to enable authentication.

### Example

1. Set the environment variables in the `.env` file:
    ```sh
    DM_USERNAME=yourusername
    DM_PASSWORD=yourpassword
    ```

2. When making requests to the API, include the `Authorization` header with the base64-encoded username and password:
    ```sh
    curl -u yourusername:yourpassword http://localhost:9095/api/metrics
    ```

## Development

1. Clone the repository:
    ```sh
    git clone https://github.com/vchan-in/doctor-metrics.git
    cd doctor-metrics
    ```

2. Install dependencies:
    ```sh
    go mod tidy
    ```

3. Copy the example environment file and modify it as needed:
    ```sh
    cp env.example .env
    ```

4. Run the tests:
    ```sh
    make test
    ```

5. Start the server for development:
    ```sh
    make run
    ```

6. You can use tools like [Postman](https://www.postman.com/) or [curl](https://curl.se/) to test the API endpoints.

## Deployment

1. Build the application:
    ```sh
    make build
    ```

2. Deploy the binary `dh` to your server.

3. Ensure Docker is installed and running on the server.

4. Set up the environment variables on the server. You can use a `.env` file or set them directly in the environment.

5. Start the application:
    ```sh
    ./bin/dh
    ```

## Environment Variables

- `DM_USERNAME` - Username for basic authentication.
- `DM_PASSWORD` - Password for basic authentication.
- `DM_CORS_ORIGIN` - Allowed origins for CORS.
- `DM_SERVER_PORT` - Port for the server to listen on.

## License

This project is licensed under the MIT License. See the LICENSE file for details.
