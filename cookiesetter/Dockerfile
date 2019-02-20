# build stage
FROM golang:latest AS build-env
RUN apt-get update
RUN apt-get upgrade -y
ADD . /src
RUN cd /src && go get github.com/gorilla/mux
RUN cd /src && CGO_ENABLED=0 GOOS=linux go build -o goapp

# final stage
FROM alpine
COPY --from=build-env /src/goapp /goapp
ENTRYPOINT /goapp
