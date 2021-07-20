ARG OSVERSION
FROM --platform=linux/amd64 gcr.io/k8s-staging-e2e-test-images/windows-servercore-cache:1.0-linux-amd64-${OSVERSION} as core
FROM --platform=linux/amd64 alpine:3.13.3 as downloader
ENV GIT_VERSION 2.30.0
ENV GIT_PATCH_VERSION 2

RUN mkdir mingit/ \
&& wget https://github.com/git-for-windows/git/releases/download/v$GIT_VERSION.windows.$GIT_PATCH_VERSION/MinGit-$GIT_VERSION.$GIT_PATCH_VERSION-busybox-64-bit.zip \
&& unzip MinGit-$GIT_VERSION.$GIT_PATCH_VERSION-busybox-64-bit.zip -d mingit/

FROM mcr.microsoft.com/windows/nanoserver:${OSVERSION}
ENV PATH "C:\mingit\cmd;C:\Windows\system32;C:\Windows;"

COPY --from=downloader /mingit mingit/
COPY --from=core /Windows/System32/netapi32.dll /Windows/System32/netapi32.dll

USER ContainerAdministrator

COPY dist /

EXPOSE 9000
EXPOSE 8000

ENTRYPOINT ["/portainer.exe"]
