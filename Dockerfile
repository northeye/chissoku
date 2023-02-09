FROM golang:1.20-alpine AS build

RUN mkdir -p /tmp/build
COPY go.* *.go /tmp/build/
RUN cd /tmp/build && go install .

FROM gcr.io/distroless/static-debian11:latest

COPY --from=build /go/bin/chissoku /usr/local/bin/

ENTRYPOINT [ "/usr/local/bin/chissoku" ]
CMD [ "--help" ]
