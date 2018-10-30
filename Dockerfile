FROM golang:1.11 as build-env
ENV GO111MODULE=on
WORKDIR /go/src/github.com/ddtmachado/lb-dxp/
ADD . /go/src/github.com/ddtmachado/lb-dxp/

RUN go get -d -v ./...
RUN go install

FROM traefik:1.7.3 as traefik

FROM gcr.io/distroless/base
COPY --from=build-env /go/bin/lb-dxp /lb-manager
COPY --from=traefik /traefik /
ENV PATH=/
EXPOSE 80
ENTRYPOINT ["/lb-manager"]