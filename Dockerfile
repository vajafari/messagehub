
FROM ARG_FROM

# MAINTAINER Tim Hockin <thockin@google.com>

ADD bin/ARG_ARCH/ARG_BIN /ARG_BIN

USER nobody:nobody
ENTRYPOINT ["/ARG_BIN"]