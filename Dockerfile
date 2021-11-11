FROM golang:1.17.3
WORKDIR /app
COPY . .
RUN go build -race -o a2
EXPOSE 8080
CMD ["./a2"]