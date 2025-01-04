FROM golang:alpine as base 

WORKDIR /builder 

RUN apk add upx 

ENV GO111MODULE=on CGO_ENABLED=0 GOOS=linux GOARCH=amd64 

COPY go.mod go.sum /builder/ 

RUN go mod download 

COPY . . 

RUN go build -o /builder/main /builder/main.go 

RUN upx -9 /builder/main 

# runner image 

FROM alpine 

WORKDIR /app 

COPY --from=base /builder/main main 

CMD ["/app/main"]

EXPOSE 8000