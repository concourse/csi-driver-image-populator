FROM concourse/golang-builder
LABEL maintainers="Kubernetes Authors"
LABEL description="Image Driver"

ADD . /src
WORKDIR /src
RUN go mod download
RUN \
    mkdir -p bin && \
    go build -o ./bin/imagepopulatorplugin ./cmd/imagepopulatorplugin && \
    cp ./bin/imagepopulatorplugin /imagepopulatorplugin

ENTRYPOINT ["/imagepopulatorplugin"]

