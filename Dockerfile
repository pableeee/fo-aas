# change latest
FROM golang:latest

WORKDIR /app

ENV GO111MODULE=on

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

COPY . .

# Enable cgo for kafka wrapper lib
RUN CGO_ENABLED=0 GOOS=linux go build

ENTRYPOINT [ "./fo-aas" ]

EXPOSE 8080