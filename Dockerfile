FROM golang:latest
WORKDIR $GOPATH/src/github.com/egegunes/LicenseBot
COPY . .
RUN go version && go get -u -v golang.org/x/vgo
RUN vgo install ./...
CMD /go/bin/LicenseBot
