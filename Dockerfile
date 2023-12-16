FROM golang:1.21-alpine AS build

WORKDIR /build
COPY go.* .
RUN go mod download
COPY *.go .
RUN go build .

FROM gcr.io/distroless/static-debian11:latest

COPY --from=build /build/chissoku /usr/local/bin/

ENTRYPOINT [ "/usr/local/bin/chissoku" ]
CMD [ "--help" ]
