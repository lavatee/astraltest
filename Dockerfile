FROM golang:1.24-alpine
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o main ./cmd/main.go
CMD [ "./main" ]
EXPOSE 8080:8080