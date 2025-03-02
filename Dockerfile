FROM golang:1.23

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o /app/issue-scouter

COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh
ENTRYPOINT ["/entrypoint.sh"]
