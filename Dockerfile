FROM alpine:3
COPY tavern /usr/local/bin/tavern

WORKDIR /data
VOLUME /data

EXPOSE 8000/tcp

# Set the default command
ENTRYPOINT [ "/usr/local/bin/tavern" ]
CMD [ "serve", "--path", "/data" ]
