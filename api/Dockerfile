FROM golang:tip-bookworm

WORKDIR /app

COPY . .
RUN go mod tidy
RUN go build -o api .

CMD ["./api"]