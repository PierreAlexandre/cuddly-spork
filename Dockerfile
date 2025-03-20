FROM debian:latest

# Install necessary dependencies
RUN apt-get update && apt-get install -y \
    curl \
    wget \
    net-tools \
    python3 \
    && apt-get clean && rm -rf /var/lib/apt/lists/*

# Install Go 1.24.1
RUN wget https://go.dev/dl/go1.24.1.linux-amd64.tar.gz && \
    tar -C /usr/local -xzf go1.24.1.linux-amd64.tar.gz && \
    rm go1.24.1.linux-amd64.tar.gz

# Set environment variables for Go
ENV PATH="/usr/local/go/bin:${PATH}"
WORKDIR /workspaces

# Ensure metrics directory exists
RUN mkdir -p /workspaces/metrics && chmod -R 777 /workspaces/metrics

# Download and install Node Exporter v1.9.0
RUN wget https://github.com/prometheus/node_exporter/releases/download/v1.9.0/node_exporter-1.9.0.linux-amd64.tar.gz && \
    tar xvf node_exporter-1.9.0.linux-amd64.tar.gz && \
    mv node_exporter-1.9.0.linux-amd64/node_exporter /usr/local/bin/ && \
    rm -rf node_exporter-1.9.0.linux-amd64.tar.gz node_exporter-1.9.0.linux-amd64

# Expose Node Exporter metrics port
EXPOSE 9100

# Copy Go application
COPY . .
