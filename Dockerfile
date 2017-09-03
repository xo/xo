FROM sergeymakinen/oracle-instant-client:12.2

ENV GOLANG_VERSION 1.9
ENV GOPATH /go
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH

RUN mkdir -p $GOPATH/src $GOPATH/bin && chmod -R 777 $GOPATH
RUN apt-get update && apt-get install -y --no-install-recommends \
      wget \
      git \
      gcc \
      libc6-dev \
      pkg-config \
      ca-certificates \
    && rm -rf /var/lib/apt/lists/*
RUN wget -O go.tgz https://storage.googleapis.com/golang/go${GOLANG_VERSION}.linux-amd64.tar.gz && \
    tar -C /usr/local -xzf go.tgz && \
    rm go.tgz

ADD contrib /contrib
ENV PKG_CONFIG_PATH /contrib

RUN go get -u \
       golang.org/x/tools/cmd/goimports \
       github.com/denisenkom/go-mssqldb \
       github.com/go-sql-driver/mysql \
       gopkg.in/rana/ora.v4 \
       github.com/lib/pq \
       github.com/mattn/go-sqlite3

RUN go get -tags oracle -u github.com/knq/xo

WORKDIR /go/src/github.com/knq/xo
