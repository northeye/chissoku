FROM golang:1.27 AS build

WORKDIR /build

ADD . .

RUN CGO_ENABLED=0 go build -o chissoku ./cmd/chissoku

FROM gcr.io/distroless/static-debian13:latest

COPY --from=build /build/chissoku /usr/local/bin/

ENTRYPOINT [ "/usr/local/bin/chissoku" ]
CMD [ "--help" ]
