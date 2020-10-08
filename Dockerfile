FROM concourse/golang-builder
LABEL maintainers="Kubernetes Authors"
LABEL description="Image Driver"

ADD . /src
WORKDIR /src
RUN \
    mkdir -p bin && \
    go build -o ./bin ./cmd/imagepopulatorplugin

COPY ./bin/imagepopulatorplugin /imagepopulatorplugin
ENTRYPOINT ["/imagepopulatorplugin"]

