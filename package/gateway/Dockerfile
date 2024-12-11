FROM golang:1.22.5-alpine3.19 as builder

WORKDIR /usr/src/app

RUN apk add --no-cache make
Run apk add --no-cache git

COPY ./rag-gateway/go.mod ./rag-gateway/go.sum ./
RUN go mod download

COPY . .
RUN cd rag-gateway && make build

FROM alpine:3.19
RUN mkdir -p /var/opt/ragger
LABEL org.opencontainers.image.description="RAG gateway for enabling self hosted RAG retrieval pipelines"

COPY --from=builder /usr/src/app/rag-gateway/ragger /usr/local/bin/ragger
COPY --from=builder /usr/src/app/rag-gateway/config.yaml /var/opt/ragger/config.yaml
RUN export PATH=$PATH:/usr/local/bin/ragger

EXPOSE 5555
ENTRYPOINT ["ragger"]