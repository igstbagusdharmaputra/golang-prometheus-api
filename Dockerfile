FROM golang:1.18.3 as BUILD

ENV GO111MODULE=on  \
    CGO_ENABLED=0   \
    GOOS=linux  \
    GOARCH=amd64

WORKDIR /apps

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

RUN go build -o main .

EXPOSE 9000

CMD ["./main"]