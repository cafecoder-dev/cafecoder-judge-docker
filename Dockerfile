FROM ubuntu:20.04

ENV DEBIAN_FRONTEND=noninteractive

#install compilers
RUN \
    apt update && \
    apt install software-properties-common apt-transport-https dirmngr curl wget time iproute2 build-essential -y && \
    # C#(mono) install
    apt install gnupg ca-certificates -y && \
    yes | apt-key adv --keyserver hkp://keyserver.ubuntu.com:80 --recv-keys 3FA7E0328081BFF6A14DA29AA6A19B38D3D831EF && \
    echo "deb https://download.mono-project.com/repo/ubuntu stable-focal main" | tee /etc/apt/sources.list.d/mono-official-stable.list && \
    apt update && \
    apt install mono-devel -y && \
    # C#(.NET) install
    wget https://packages.microsoft.com/config/ubuntu/20.04/packages-microsoft-prod.deb -O packages-microsoft-prod.deb && \
    dpkg -i packages-microsoft-prod.deb && \
    apt update && \
    apt-get update && \
    apt-get install dotnet-sdk-3.1 -y && \
    apt-get install -y aspnetcore-runtime-3.1 && \
    # C/C++ install
    apt-get install g++-10 gcc-10 -y && \
    # Java11 install
    apt install default-jdk -y && \
    # Python3 install
    apt install python3 -y && \
    # go install
    wget https://golang.org/dl/go1.14.7.linux-amd64.tar.gz && \
    tar -C /usr/local -xzf go1.14.7.linux-amd64.tar.gz && \
    # Rust install
    curl https://sh.rustup.rs -sSf | sh -s -- -y && \
    # Nim install
    curl https://nim-lang.org/choosenim/init.sh -sSf | sh -s -- -y

COPY vendor .
COPY go.mod .
COPY go.sum .
COPY main.go .
RUN export PATH=$PATH:/usr/local/go/bin && go build -mod=mod -o .

# download ACL
RUN \
    apt install unzip && \
    wget earlgray283.github.io/download/atcoder.zip && \
    unzip atcoder.zip

WORKDIR / 

ENTRYPOINT ["./cafecoder-container-client"]
