FROM ubuntu:20.04

ENV DEBIAN_FRONTEND=noninteractive

# install compilers
RUN \
    apt update && \
    apt install software-properties-common apt-transport-https dirmngr curl wget time iproute2 build-essential sudo unzip -y && \
    touch ~/.profile && \
    # C#(mono) install
    apt install gnupg ca-certificates -y && \
    yes | apt-key adv --keyserver hkp://keyserver.ubuntu.com:80 --recv-keys 3FA7E0328081BFF6A14DA29AA6A19B38D3D831EF && \
    echo "deb https://download.mono-project.com/repo/ubuntu stable-focal main" | tee /etc/apt/sources.list.d/mono-official-stable.list && \
    apt update && \
    apt install mono-devel -y && \
    # C#(.NET) install
    wget https://download.visualstudio.microsoft.com/download/pr/820db713-c9a5-466e-b72a-16f2f5ed00e2/628aa2a75f6aa270e77f4a83b3742fb8/dotnet-sdk-5.0.100-linux-x64.tar.gz && \
    mkdir -p $HOME/dotnet && tar zxf dotnet-sdk-5.0.100-linux-x64.tar.gz -C $HOME/dotnet && \
    echo 'export DOTNET_ROOT=$HOME/dotnet' >> ~/.profile && \
    echo 'export PATH=$PATH:$HOME/dotnet' >> ~/.profile && \
    # C/C++ install
    apt-get install g++-10 gcc-10 -y && \
    # Java11 install
    apt install default-jdk -y && \
    # Python3 install
    apt install python3 -y && \
    # go install
    wget https://golang.org/dl/go1.15.5.linux-amd64.tar.gz && \
    tar -C /usr/local -xzf go1.15.5.linux-amd64.tar.gz && \
    # Rust install
    curl https://sh.rustup.rs -sSf | sh -s -- -y && \
    # Nim install
    curl https://nim-lang.org/choosenim/init.sh -sSf | sh -s -- -y && \
    echo -e 'export PATH=/root/.nimble/bin:$PATH\n' >> ~/.profile && \
    # Raku install
    apt-key adv --keyserver hkp://keyserver.ubuntu.com:80 --recv 379CE192D401AB61 && \
    echo "deb https://dl.bintray.com/nxadm/rakudo-pkg-debs `lsb_release -cs` main" | tee -a /etc/apt/sources.list.d/rakudo-pkg.list && \
    apt-get update && apt-get install rakudo-pkg && \
    /opt/rakudo-pkg/bin/add-rakudo-to-path && \
    source /home/earlgray/.profile && \
    # Ruby install
    git clone https://github.com/sstephenson/rbenv.git ~/.rbenv && \
    echo 'export PATH="$HOME/.rbenv/bin:$PATH"' >> ~/.profile && \
    echo 'eval "$(rbenv init -)"' >> ~/.profile && \
    exec $SHELL -l && \
    git clone https://github.com/sstephenson/ruby-build.git ~/.rbenv/plugins/ruby-build && \
    rbenv install 2.7.2 && rbenv global 2.7.2 && \
    # Kotlin install
    curl -s https://get.sdkman.io | bash && \
    source "/home/earlgray/.sdkman/bin/sdkman-init.sh" && \
    sdk install kotlin && \
    # Fortran install
    apt install gfortran-10 -y && \
    # crystal
    curl -sSL https://dist.crystal-lang.org/apt/setup.sh | bash && \
    apt install crystal && \
    # Perl install
    wget https://www.cpan.org/src/5.0/perl-5.32.0.tar.gz && \
    tar -xzf perl-5.32.0.tar.gz && \
    cd perl-5.32.0 && \
    ./Configure -des -Dprefix=$HOME/localperl && \
    make && \
    make test && \
    make install



    
# install external libraries
RUN \
    wget https://raw.githubusercontent.com/MikeMirzayanov/testlib/master/testlib.h && \
    wget https://github.com/atcoder/ac-library/releases/download/v1.0/ac-library.zip && \
    unzip ac-library.zip

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
