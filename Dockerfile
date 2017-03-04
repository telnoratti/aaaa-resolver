FROM golang:onbuild

EXPOSE 8053

ENV ZONE=${ZONE:-ipv6-literal}

CMD app --zone ${ZONE}
