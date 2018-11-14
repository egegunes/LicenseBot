FROM golang:1.11 as builder
RUN mkdir /src
WORKDIR /src
ADD go.mod .
RUN go mod download
ADD . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /src/licensebot .

FROM alpine
RUN apk add ca-certificates
COPY --from=builder /src/licensebot .
CMD ["./licensebot"]
