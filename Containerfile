FROM scratch
ARG BUILT_ARTIFACT
ADD "$BUILT_ARTIFACT" /
ENTRYPOINT [ "/kci" ]
CMD [ ]
