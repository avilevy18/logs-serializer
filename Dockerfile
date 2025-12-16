FROM golang:2.24

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o /logs-serializer

EXPOSE 18888 18889

CMD [ "/logs-serializer" ]
