FROM scratch
COPY tfplan-check /
ENTRYPOINT [ "./tfplan-check" ]
