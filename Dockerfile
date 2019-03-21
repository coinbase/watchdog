FROM golang:1.12

ADD . /watchdog

WORKDIR /watchdog

RUN make install

EXPOSE 3000

ENTRYPOINT ["watchdog"]
