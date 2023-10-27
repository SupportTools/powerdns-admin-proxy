# PowerDNS Admin Proxy

This application serves as a reverse proxy for the PowerDNS Admin service. It forwards incoming HTTP requests to a specified backend and also provides metrics and logging capabilities.

## Prerequisites

- Go 1.21 or higher
- Git

## Installation

1. Clone the repository:

    ```bash
    git clone https://github.com/supporttools/powerdns-admin-proxy.git
    ```

2. Change to the directory:

    ```bash
    cd powerdns-admin-proxy
    ```

3. Edit Helm values
   
   ```bash
   vi ./charts/powerdns-admin-proxy/values.yaml
   ```

NOTE: You will need to change the `backendUrl` to the URL of your PowerDNS Admin service and the ingress host to the hostname of your PowerDNS Admin service.

4. Install Helm chart:

    ```bash
    helm upgrade --install pdns-proxy -n powerdns --create-namespace ./charts/powerdns-admin-proxy
    ```

## Configuration

The application can be configured using environment variables.

- `PORT`: The port where the proxy listens for incoming HTTP requests. Default is `8080`.
- `BACKEND_URL`: The URL of the backend service where requests will be forwarded. Default is `http://powerdns-server:8081`.
- `DEBUG`: Enable debug logging. Set this to `true` to enable.

For example:

```bash
export PORT=8080
export BACKEND_URL=http://powerdns-server:8081
export DEBUG=true
```

## Build

To build the application:

```bash
go build -o powerdns-admin-proxy
```

This will compile the code and generate an executable named `powerdns-admin-proxy`.

## Run

After building, you can run the application as follows:

```bash
./powerdns-admin-proxy
```

## Metrics

Prometheus metrics are available at `http://localhost:9000/metrics`.

## Health Check

A simple health check endpoint is available at `http://localhost:8080/healthz`.

## Debugging

Debugging logs are enabled by setting the `DEBUG` environment variable to `true`.

## License

This project is licensed under the Apache License, Version 2.0. See [LICENSE](LICENSE) for the full license text.