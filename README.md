# cafecoder-judge-docker
ホストからのリクエストを受け付ける。主に `docker build` のみ。

# Requirements
+ go(1.15)
+ docker(19.03)
+ key.json ... gcp のキーファイル

# Usage
```console
$ go mod vendor
$ docker build -t cafecoder:latest .
```