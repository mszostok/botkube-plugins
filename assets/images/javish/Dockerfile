FROM polinux/stress
COPY run.sh /scripts/run.sh
RUN ls /scripts
RUN chmod 755 /scripts/run.sh
WORKDIR /scripts
CMD ["/scripts/run.sh"]
