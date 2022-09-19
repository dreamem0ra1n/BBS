FROM golang:1.17 AS builder
WORKDIR /usr/src/app
COPY go.mod go.sum ./
RUN go env -w "GO111MODULE=on" \
	&& go env -w "GOPROXY=https://proxy.golang.com.cn,direct" \
	&& go mod download && go mod verify
COPY . .
RUN GOOS=linux GOARCH=amd64 go build -v -o /usr/src/app/ ./... \
	&& ls /usr/src/app

FROM busybox:latest as prod
WORKDIR /
COPY --from=builder /usr/src/app/bbs-go /bbs-go

CMD ["/bbs-go"]