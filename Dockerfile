FROM golang:1.21-alpine AS build

WORKDIR /build

COPY . .
RUN go build .

FROM gcr.io/distroless/static-debian12:latest

COPY --from=build /build/chissoku /usr/local/bin/

ENTRYPOINT [ "/usr/local/bin/chissoku" ]
CMD [ "--help" ]
