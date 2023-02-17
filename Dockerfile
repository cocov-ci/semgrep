FROM golang:alpine AS builder

ARG go_token

#RUN apt update && apt install git openssl
RUN apk --no-cache add git openssl
RUN git config --global url."https://oauth2:$go_token@github.com/".insteadOf "https://github.com/"

ENV GOPRIVATE="github.com/cocov-ci"
ENV CGO_ENABLED=0

RUN mkdir /app
WORKDIR /app
COPY . .
RUN go build cmd/main.go

FROM golang:alpine

COPY --from=builder /app/main /bin/plugin-semgrep

RUN apk --no-cache add git gcc musl-dev python3 py3-pip python3-dev
RUN apk --no-cache add --virtual deps curl

RUN python3 -m pip install semgrep

RUN apk del deps

RUN addgroup -g 1000 cocov && \
    adduser --shell /bin/ash --disabled-password \
   --uid 1000 --ingroup cocov cocov

USER cocov

CMD ["plugin-semgrep"]

