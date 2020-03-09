FROM umputun/baseimage:buildgo-v1.3.1 as build

WORKDIR /build/sat-parser
COPY . /build/sat-parser

# run tests
RUN cd app && go test -mod=vendor ./...

RUN \
    version=$(/script/git-rev.sh) && \
    echo "version=$version" && \
    go build -mod=vendor -o sat-parser -ldflags "-X main.revision=${version} -s -w" ./app


FROM umputun/baseimage:app-v1.3.1

LABEL maintainer="Artem Kolin <artemkaxboy@gmail.com>"

ENV TIME_ZONE=Asia/Novosibirsk

COPY --from=build /build/sat-parser/sat-parser /srv/sat-parser
COPY sat-parser.conf /srv/

RUN \
    chown -R app:app /srv && \
    chmod +x /srv/sat-parser

WORKDIR /srv

CMD ["/srv/sat-parser"]
ENTRYPOINT ["/init.sh"]
