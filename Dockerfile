FROM quay.io/vektorcloud/go:dep AS source

COPY . /go/src/github.com/mesanine/mbt

RUN cd /go/src/github.com/mesanine/mbt \
  && make 

FROM quay.io/vektorcloud/base:3.6 

COPY --from=source /go/src/github.com/mesanine/mbt/bin/mbt /usr/bin/

ENTRYPOINT ["/usr/bin/mbt"]
