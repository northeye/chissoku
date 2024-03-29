VERSION --try 0.7
FROM golang:1.21

ENV GOPATH=/go
ENV PATH=$GOPATH/bin:$PATH

deps:
    RUN apt-get update && apt-get install -y --no-install-recommends p7zip
    RUN curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.55.2
    SAVE IMAGE --cache-hint

lint:
    FROM +deps
    WORKDIR /workspace/lint
    COPY . .
    TRY
        RUN --no-cache golangci-lint run -c ./.golangci-lint.yml --out-format junit-xml > lint-report.xml
    FINALLY
        SAVE ARTIFACT lint-report.xml AS LOCAL lint-report.xml
    END

build:
    FROM +deps
    ARG TARGET_OS=$(go env GOOS)
    ARG TARGET_ARCH=$(go env GOARCH)
    WORKDIR /workspace/build
    COPY go.mod go.sum .
    RUN go mod download
    COPY . .
    RUN CGO_ENABLED=0 GOOS=$TARGET_OS GOARCH=$TARGET_ARCH go build ./
    RUN rm -rf release && mkdir -p release
    IF [ "$TARGET_OS" = "windows" ]
        RUN 7zr a release/chissoku-$(go run . -v)-windows-$TARGET_ARCH.7z chissoku.exe
    ELSE
        RUN tar -czf release/chissoku-$(go run . -v)-$TARGET_OS-$TARGET_ARCH.tar.gz chissoku
    END
    SAVE ARTIFACT release/* release/

release:
    # NOT IMPLEMENTED YET