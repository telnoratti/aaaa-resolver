FROM golang:onbuild

EXPOSE 8053

ENV ZONE=${ZONE:-ipv6-literal}
ENV NS=${NS:-ns1.ipv6-literal}
ENV MBOX=${MBOX:-hostmaster.ipv6.literal}

CMD app --zone ${ZONE} --ns ${NS} --mbox ${MBOX}
