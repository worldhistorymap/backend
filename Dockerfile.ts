FROM golang:1.12

WORKDIR $GOPATH/src/github.com/worldhistorymap/backend
COPY cmd/tileserver/main.go cmd/tileserver/main.go
COPY pkg/tileserver/tileserver.go pkg/tileserver/tileserver.go
RUN go get -d ./...
RUN go install ./...
RUN mkdir /tiles/
EXPOSE 8000

VOLUME ["/tiles/"]

WORKDIR cmd/tileserver
RUN go build 
CMD ["./tileserver"]





