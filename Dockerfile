FROM golang:1.12

WORKDIR /dockersh
ADD . /dockersh/
RUN cd /dockersh && go build -mod vendor && chmod +x ./installer.sh

CMD ["./installer.sh"]

