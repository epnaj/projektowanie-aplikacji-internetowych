# this is a devenv docker

FROM golang:1.26.3-bookworm

RUN apt install git

# swap to current user
ARG GetMyUsername
RUN echo ${GetMyUsername}
USER ${GetMyUsername}

ENTRYPOINT ["/bin/bash"]