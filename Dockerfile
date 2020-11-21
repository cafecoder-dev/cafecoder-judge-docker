FROM ubuntu:20.04

ENV DEBIAN_FRONTEND=noninteractive
SHELL ["/bin/bash", "-c"]

# install compilers
RUN \
    apt update && \
    apt install software-properties-common apt-transport-https dirmngr curl wget time iproute2 build-essential sudo unzip git -y && \
    touch ~/.profile

# Raku install
RUN apt-key adv --keyserver hkp://keyserver.ubuntu.com:80 --recv 379CE192D401AB61 && \
    echo "deb https://dl.bintray.com/nxadm/rakudo-pkg-debs `lsb_release -cs` main" | tee -a /etc/apt/sources.list.d/rakudo-pkg.list && \
    apt-get update && apt-get install apt-utils && apt-get install rakudo-pkg && \
    /opt/rakudo-pkg/bin/add-rakudo-to-path
    #source ~/.profile
   
# C#(mono) install
RUN apt install gnupg ca-certificates -y && \
    yes | apt-key adv --keyserver hkp://keyserver.ubuntu.com:80 --recv-keys 3FA7E0328081BFF6A14DA29AA6A19B38D3D831EF && \
    echo "deb https://download.mono-project.com/repo/ubuntu stable-focal main" | tee /etc/apt/sources.list.d/mono-official-stable.list && \
    apt update && \
    apt install mono-devel -y

# C#(.NET) install
RUN wget https://download.visualstudio.microsoft.com/download/pr/820db713-c9a5-466e-b72a-16f2f5ed00e2/628aa2a75f6aa270e77f4a83b3742fb8/dotnet-sdk-5.0.100-linux-x64.tar.gz && \
    mkdir -p $HOME/dotnet && tar zxf dotnet-sdk-5.0.100-linux-x64.tar.gz -C $HOME/dotnet && \
    echo 'export DOTNET_ROOT=$HOME/dotnet' >> ~/.profile && \
    echo 'export PATH=$PATH:$HOME/dotnet' >> ~/.profile

# C/C++ install
RUN apt-get install g++-10 gcc-10 -y

# Java11 install
RUN apt install default-jdk -y

# Python3 install
RUN apt install python3 -y

# Pypy3 install
RUN apt install pypy3 -y

# go install
RUN wget https://golang.org/dl/go1.15.5.linux-amd64.tar.gz && \
    tar -C /usr/local -xzf go1.15.5.linux-amd64.tar.gz && \
    echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.profile

ENV USER=$USER
    
# Rust install
RUN curl https://sh.rustup.rs -sSf | sh -s -- -y && \
    source $HOME/.cargo/env && \
    cargo new rust_workspace && \
    cd rust_workspace &&\
    wget https://raw.githubusercontent.com/cafecoder-dev/language-update/20.10/Rust/Cargo.toml -O Cargo.toml && \
    cargo build --release && \
    cd /

# Nim install
RUN curl https://nim-lang.org/choosenim/init.sh -sSf | sh -s -- -y && \
    echo 'export PATH=/root/.nimble/bin:$PATH' >> ~/.profile
    
# Ruby install
RUN apt install make libffi-dev openssl libssl-dev zlib1g-dev -y && \
    git clone https://github.com/sstephenson/rbenv.git ~/.rbenv && \
    echo 'export PATH="$HOME/.rbenv/bin:$PATH"' >> ~/.profile && \
    echo 'eval "$(rbenv init -)"' >> ~/.profile && \
    bash -c exec $SHELL -l && \
    git clone https://github.com/sstephenson/ruby-build.git ~/.rbenv/plugins/ruby-build && \
    export PATH="$HOME/.rbenv/bin:$PATH" && rbenv install 2.7.2 && rbenv global 2.7.2

# Kotlin install
RUN apt install zip unzip -y && \
    curl -s https://get.sdkman.io | bash && \
    bash && \
    echo 'source "/root/.sdkman/bin/sdkman-init.sh"' >> ~/.profile && \
    source ~/.profile && \
    sdk install kotlin

# Fortran install
RUN apt install gfortran-10 -y
    
# crystal
RUN curl -sSL https://dist.crystal-lang.org/apt/setup.sh | bash && \
    apt install crystal -y
    
# Perl install
RUN wget https://www.cpan.org/src/5.0/perl-5.32.0.tar.gz && \
    tar -xzf perl-5.32.0.tar.gz && \
    cd perl-5.32.0 && \
    ./Configure -Dprefix=$HOME/perl -Dscriptdir=$HOME/perl/bin -des -Dman1dir=none -Dman3dir=none -DDEBUGGING=-g && \
    make --jobs=8 install

# install external libraries
RUN \
    wget https://raw.githubusercontent.com/MikeMirzayanov/testlib/master/testlib.h && \
    wget https://github.com/atcoder/ac-library/releases/download/v1.0/ac-library.zip && unzip ac-library.zip

# system
RUN \
    useradd --create-home cafecoder && \
    echo 'cafecoder hard nproc 4096' >> /etc/security/limits.conf && \
    chmod -R 777 /home && \
    mkdir Main -m 777

ENV TZ Asia/Tokyo

COPY vendor /vendor
COPY key.json .
COPY gcplib /gcplib
COPY util /util
COPY go.mod .
COPY go.sum .
COPY main.go .
RUN export PATH=$PATH:/usr/local/go/bin && go build -mod=vendor -o .

WORKDIR / 

ENTRYPOINT ["./cafecoder-container-client"]
