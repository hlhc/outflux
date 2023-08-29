FROM alpine:3.18
WORKDIR /app
RUN apk install --no-cache bash
COPY outflux /app/outflux
ENTRYPOINT [ "" ]
