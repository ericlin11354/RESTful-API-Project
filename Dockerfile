FROM golang:latest
WORKDIR /app
COPY . .
RUN go build -race -o a2
EXPOSE 8080