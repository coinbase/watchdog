FROM golang@sha256:ea29c32b677885c58e03a3b27ee2f5db5f65f8a57c4ab73676d5ef07ff9deca6

ADD . /watchdog

WORKDIR /watchdog

RUN make install

EXPOSE 3000

ENTRYPOINT ["watchdog"]
