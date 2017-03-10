FROM alpine:latest
ARG VER

COPY bin/linux64/cf-copy-plugin /usr/local/bin/ 

RUN apk --no-cache add ca-certificates && update-ca-certificates
RUN apk --no-cache add curl
RUN curl -L "https://cli.run.pivotal.io/stable?release=linux64-binary&version=6.25.0&source=github-rel" | tar xz -C /usr/local/bin/ cf
RUN cf install-plugin -f -r CF-Community "Targets"

RUN curl -L "https://github.com/mevansam/cf-copy-plugin/releases/download/$VER/cf-copy-plugin_linux64.tar.gz" | tar xz -C /usr/local/bin/ cf-copy-plugin
RUN cf install-plugin -f /usr/local/bin/cf-copy-plugin && rm /usr/local/bin/cf-copy-plugin
