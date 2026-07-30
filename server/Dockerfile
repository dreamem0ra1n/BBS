FROM golang:1.17 AS dependencies
WORKDIR /usr/src/app
COPY go.mod go.sum ./
RUN go env -w "GO111MODULE=on" \
	&& go env -w "GOPROXY=https://goproxy.cn,direct" \
	&& go mod download

FROM dependencies AS dev-builder
COPY . .
RUN GOOS=linux GOARCH=amd64 go build -tags passwordlogin -o /usr/src/app/bbs-go .

FROM busybox:latest AS runtime
WORKDIR /
EXPOSE 8082
CMD ["/bbs-go"]

FROM runtime AS dev
COPY --from=dev-builder /usr/src/app/bbs-go /bbs-go

FROM dependencies AS prod-builder
COPY . .
RUN GOOS=linux GOARCH=amd64 go build -o /usr/src/app/bbs-go .

FROM runtime AS prod
COPY --from=prod-builder /usr/src/app/bbs-go /bbs-go
