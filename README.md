## What is this

IO-DATA製CO2センサー [UD-CO2S](https://www.iodata.jp/product/tsushin/iot/ud-co2s/) から測定データを読み取り、MQTTや標準出力へJSONデータを出力するプログラムです。([Amazon.co.jp](https://amzn.to/3DX78Hi))

## Install

### with go install

```console
$ go install github.com/northeye/chissoku@v2
```
### Download binary

[リリースページ](https://github.com/northeye/chissoku/releases)からダウンロー<br>ド。

## How to use

デバイスを接続してシリアルポートの確認をしておきます。<br>
コマンドライン引数にシリアルポートのデバイス名を指定して実行します。

シリアルデバイスが `/dev/ttyACM0` の場合 (Linux等)
```console
$ ./chissoku -q /dev/ttyACM0 --tags Living
{"co2":1242,"humidity":31.3,"temperature":29.4,"tags":["Living"],"timestamp":"2023-02-01T20:50:51.240+09:00"}
```

シリアルデバイスが `COM3` の場合(Windows)
```cmd.exe
C:\> chissoku.exe -q COM3 --tags Living
{"co2":1242,"humidity":31.3,"temperature":29.4,"tags":["Living"],"timestamp":"2023-02-01T20:50:51.240+09:00"}
```

※ センサーデータ(JSON)以外のプロセス情報は標準エラー(stderr)に出力されます。

### with Docker image

```console
$ docker run --rm -it --device /dev/ttyACM0:/dev/ttyACM0 ghcr.io/northeye/chissoku:latest /dev/ttyACM0 [<options>]
```
※ そもそもシングルバイナリなのでdockerで動かす意味はないかと思います。

### Outputter

`--output` オプションにより出力メソッドを指定することが可能です。<br>
現在用意されているメソッドは `stdout`, `mqtt` で、複数指定することも可能です。

```console
$ chissoku --output=stdout,mqtt --mqtt.address tcp://mosquitto:1883/ --mqtt.topic=sensors/co2 --mqtt.qos=2  /dev/ttyACM0
```

何も指定しなければデフォルトとして `stdout` が選択されます。

outputter にはそれぞれオプションが指定可能な場合があります。<br>
outputter のオプションは基本的に outputter の名前がプレフィックスになっています。

今後ファイルやクラウド出力等のメソッドが実装されるかもしれません。

### Stdout Outputter

コマンドラインオプションの `--output=stdout` により標準出力にデータを流せます。<br>
|オプション|意味|
|----|----|
|--stdout.interval=`INT`|データを出力する間隔(秒)(`default: 60`)|

### MQTT Outputter

コマンドラインオプションの `--output=mqtt` により MQTTブローカーへデータを流せます。<br>
必要な場合はSSLの証明書やUsername,Passwordを指定することができます。

|オプション|意味|
|----|----|
|--mqtt.interval=`INT`|データを出力する間隔(秒)(`default: 60`)|
|--mqtt.address=`STRING`|MQTTブローカーURL (例: `tcp://mosquitto:1883`, `ssl://mosquitto:8883`)|
|--mqtt.topic=`STRING`|Publish topic (例: `sensors/co2`)|
|--mqtt.client-id=`STRING`|MQTT Client ID `default: chissoku`|
|--mqtt.qos=`INT`|publish QoS `default: 0`|
|--mqtt.ssl-ca-file=`STRING`|SSL Root CA|
|--mqtt.ssl-cert=`STRING`|SSL Client Certificate|
|--mqtt.ssl-key=`STRING`|SSL Client Private Key|
|--mqtt.username=`STRING`|MQTT v3.1/3.1.1 Authenticate Username|
|--mqtt.password=`STRING`|MQTT v3.1/3.1.1 Authenticate Password|

**Tips**

MQTT メソッドがうまく動かなければ標準出力を [mosquitto_pub](https://mosquitto.org/man/mosquitto_pub-1.html) などに渡せばうまくいくかもしれません。

### Global options

|オプション|意味|
|----|----|
|-o, --output=`stdout,...`|出力メソッドの指定(`default: stdout`)|
|-q, --quiet|標準エラーの出力をしない|
|-t, --tags=`TAG,...`|出力するJSONに `tags` フィールドを追加する(コンマ区切り文字列)|
|-h, --help|オプションヘルプを表示する|
|-v, --version|バージョン情報を表示する|
|-d, --debug|デバッグログの出力を行う|

