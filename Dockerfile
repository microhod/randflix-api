ARG BASE_PATH=/go/src/github.com/microhod/randflix-api

FROM golang:1.15 AS build

ARG BASE_PATH

COPY . ${BASE_PATH}
WORKDIR ${BASE_PATH}

RUN GO111MODULE=on CGO_ENABLED=0 go build -o randflix-api .

FROM scratch AS final

ARG BASE_PATH

# Copy ca certs from build container
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
# Copy application
COPY --from=build ${BASE_PATH}/randflix-api /randflix-api

EXPOSE 8080
ENTRYPOINT [ "/randflix-api" ]
