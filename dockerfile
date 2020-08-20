FROM loads/alpine:3.8

#LABEL maintainer="john@johng.cn"

ENV WORKDIR /var/app/main

ADD ./main   $WORKDIR/main
RUN chmod +x $WORKDIR/main

ADD ./config.yaml   $WORKDIR/config.yaml

WORKDIR $WORKDIR

CMD ./main