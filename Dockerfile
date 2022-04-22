FROM golang:1.16-alpine
WORKDIR /app
COPY . ./
RUN go mod download
RUN go build -o /write-template-2-flux
ENTRYPOINT [ "/write-template-2-flux" ]