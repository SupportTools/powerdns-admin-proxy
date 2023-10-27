FROM golang:1.21.3-alpine3.18 AS builder

# Install git.
# Git is required for fetching the dependencies.
RUN apk update && apk add --no-cache git

# Set the Current Working Directory inside the container
WORKDIR /src

# Copy everything from the current directory to the PWD(Present Working Directory) inside the container
COPY . .

# Download all the dependencies
RUN go get -d -v ./...

# Install the dependencies
RUN go install -v ./...

# Build the Go app
RUN GOOS=linux GOARCH=amd64 go build -o /go/bin/powerdns-admin-proxy .
RUN chmod +x /go/bin/powerdns-admin-proxy

FROM scratch

# Set the Current Working Directory inside the container

WORKDIR /go/bin

# Copy our static executable.
COPY --from=builder /go/bin/powerdns-admin-proxy /go/bin/powerdns-admin-proxy

# Run the hello binary.
ENTRYPOINT ["/go/bin/powerdns-admin-proxy"]
