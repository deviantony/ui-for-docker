FROM ubuntu:20.04

# Expose port for the Portainer UI and Edge server
EXPOSE 9000
EXPOSE 8000

WORKDIR /src/portainer

# Set TERM as noninteractive to suppress debconf errors
RUN echo 'debconf debconf/frontend select Noninteractive' | debconf-set-selections

# Set default go version
ARG GO_VERSION=go1.16.linux-amd64

# Install packages
RUN apt-get update --fix-missing && apt-get install -qq \
    dialog \
    apt-utils \
    curl \
    build-essential \
    git \
    wget \
    apt-transport-https \
    ca-certificates \
    gnupg-agent \
    software-properties-common

# Install Docker CLI
RUN curl -fsSL https://download.docker.com/linux/ubuntu/gpg | apt-key add - \
    && add-apt-repository \
   "deb [arch=amd64] https://download.docker.com/linux/ubuntu \
   $(lsb_release -cs) \
   stable" \
   && apt-get update \
   && apt-get install -y docker-ce-cli


# Install NodeJS
RUN curl -fsSL https://deb.nodesource.com/setup_14.x | bash - \
    && apt-get install -y nodejs

# Install Yarn
RUN curl -sS https://dl.yarnpkg.com/debian/pubkey.gpg | apt-key add - \
    && echo "deb https://dl.yarnpkg.com/debian/ stable main" | tee /etc/apt/sources.list.d/yarn.list \
    && apt-get update && apt-get -y install yarn

# Install Golang
RUN cd /tmp \
    && wget -q https://dl.google.com/go/${GO_VERSION}.tar.gz \
    && tar -xf ${GO_VERSION}.tar.gz \
    && mv go /usr/local

# Copy run script
COPY run.sh /
RUN chmod +x /run.sh

ENTRYPOINT ["/run.sh"]
