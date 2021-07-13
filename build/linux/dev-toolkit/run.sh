#!/usr/bin/env bash

# Script used to init the Portainer development environment inside the dev-toolkit image

### COLOR OUTPUT ###

ESeq="\x1b["
RCol="$ESeq"'0m'    # Text Reset

# Regular               Bold                    Underline               High Intensity          BoldHigh Intens         Background              High Intensity Backgrounds
Bla="$ESeq"'0;30m';     BBla="$ESeq"'1;30m';    UBla="$ESeq"'4;30m';    IBla="$ESeq"'0;90m';    BIBla="$ESeq"'1;90m';   On_Bla="$ESeq"'40m';    On_IBla="$ESeq"'0;100m';
Red="$ESeq"'0;31m';     BRed="$ESeq"'1;31m';    URed="$ESeq"'4;31m';    IRed="$ESeq"'0;91m';    BIRed="$ESeq"'1;91m';   On_Red="$ESeq"'41m';    On_IRed="$ESeq"'0;101m';
Gre="$ESeq"'0;32m';     BGre="$ESeq"'1;32m';    UGre="$ESeq"'4;32m';    IGre="$ESeq"'0;92m';    BIGre="$ESeq"'1;92m';   On_Gre="$ESeq"'42m';    On_IGre="$ESeq"'0;102m';
Yel="$ESeq"'0;33m';     BYel="$ESeq"'1;33m';    UYel="$ESeq"'4;33m';    IYel="$ESeq"'0;93m';    BIYel="$ESeq"'1;93m';   On_Yel="$ESeq"'43m';    On_IYel="$ESeq"'0;103m';
Blu="$ESeq"'0;34m';     BBlu="$ESeq"'1;34m';    UBlu="$ESeq"'4;34m';    IBlu="$ESeq"'0;94m';    BIBlu="$ESeq"'1;94m';   On_Blu="$ESeq"'44m';    On_IBlu="$ESeq"'0;104m';
Pur="$ESeq"'0;35m';     BPur="$ESeq"'1;35m';    UPur="$ESeq"'4;35m';    IPur="$ESeq"'0;95m';    BIPur="$ESeq"'1;95m';   On_Pur="$ESeq"'45m';    On_IPur="$ESeq"'0;105m';
Cya="$ESeq"'0;36m';     BCya="$ESeq"'1;36m';    UCya="$ESeq"'4;36m';    ICya="$ESeq"'0;96m';    BICya="$ESeq"'1;96m';   On_Cya="$ESeq"'46m';    On_ICya="$ESeq"'0;106m';
Whi="$ESeq"'0;37m';     BWhi="$ESeq"'1;37m';    UWhi="$ESeq"'4;37m';    IWhi="$ESeq"'0;97m';    BIWhi="$ESeq"'1;97m';   On_Whi="$ESeq"'47m';    On_IWhi="$ESeq"'0;107m';

printSection() {
  echo -e "${BIYel}>>>> ${BIWhi}${1}${RCol}"
}

info() {
  echo -e "${BIWhi}${1}${RCol}"
}

success() {
  echo -e "${BIGre}${1}${RCol}"
}

error() {
  echo -e "${BIRed}${1}${RCol}"
}

errorAndExit() {
  echo -e "${BIRed}${1}${RCol}"
  exit 1
}

### !COLOR OUTPUT ###

SETUP_FILE=/setup-done

display_configuration() {
  info "Portainer dev-toolkit container configuration"
  info "Go version"
  /usr/local/go/bin/go version
  info "Node version"
  node -v
  info "Yarn version"
  yarn -v
  info "Docker version"
  docker version
}

main() {
    [[ -z $PUSER ]] && errorAndExit "Unable to find PUSER environment variable. Please ensure PUSER is set before running this script."
    [[ -z $PUID ]] && errorAndExit "Unable to find PUID environment variable. Please ensure PUID is set before running this script."
    [[ -z $PGID ]] && errorAndExit "Unable to find PGID environment variable. Please ensure PGID is set before running this script."
    [[ -z $DOCKERGID ]] && errorAndExit "Unable to find DOCKERGID environment variable. Please ensure DOCKERGID is set before running this script."

    if [[ -f "${SETUP_FILE}" ]]; then
        info "Portainer dev-toolkit container already configured."
        display_configuration
    else
        info "Creating user group..."
        groupadd -g $PGID $PUSER

        info "Creating user..."
        useradd -l -u $PUID -g $PUSER $PUSER

        info "Setting up home..."
        install -d -m 0755 -o $PUSER -g $PUSER /home/$PUSER

        info "Configuring Docker..."
        groupadd -g $DOCKERGID docker
        usermod -aG docker $PUSER

        info "Configuring Go..."
        echo "PATH=\"$PATH:/usr/local/go/bin\"" > /etc/environment

        info "Configuring Git..."
        su $PUSER -c "git config --global url.git@github.com:.insteadOf https://github.com/"

        info "Configuring SSH..."
        mkdir /home/$PUSER/.ssh
        cp /host-ssh/* /home/$PUSER/.ssh/
        chown -R $PUSER:$PUSER /home/$PUSER/.ssh

        touch "${SETUP_FILE}"
        success "Portainer dev-toolkit container successfully configured."

        display_configuration
    fi
}

main
su $PUSER -s "$@"