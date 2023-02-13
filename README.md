## What is this

IO-DATA製CO2センサー [UD-CO2S](https://www.iodata.jp/product/tsushin/iot/ud-co2s/) から測定データを読み取り、MQTTや標準出力へJSONデータを出力するプログラムです。([Amazon.co.jp](https://amzn.to/3DX78Hi))

## Install

### with go install

```sh
$ go install github.com/northeye/chissoku@latest
```
### Download binary

[リリースページ](https://github.com/northeye/chissoku/releases)からダウンロー<br>ド。

## How to use

デバイスを接続してシリアルポートの確認をしておきます。<br>
コマンドライン引数にシリアルポートのデバイス名を指定して実行します。

シリアルデバイスが `/dev/ttyACM0` の場合 (Linux等)
```sh
$ ./chissoku /dev/ttyACM0 --tags Living
I: Prepare device... STP ID? STA OK.
{"co2":1242,"humidity":31.3,"temperature":29.4,"tags":["Living"],"timestamp":"2023-02-01T20:50:51.240+09:00"}
```

シリアルデバイスが `COM3` の場合(Windows)
```cmd.exe
C:\> chissoku.exe COM3 --tags Living
I: Prepare device... STP ID? STA OK.
{"co2":1242,"humidity":31.3,"temperature":29.4,"tags":["Living"],"timestamp":"2023-02-01T20:50:51.240+09:00"}
```

※ センサーデータ(JSON)以外のプロセス情報は標準エラー(stderr)に出力されます。

### with Docker image

```sh
$ docker run --rm -it --device /dev/ttyACM0:/dev/ttyACM0 ghcr.io/northeye/chissoku:latest /dev/ttyACM0 [<options>]
```
※ そもそもシングルバイナリなのでdockerで動かす意味はないかと思います。

### with MQTT broker

下記のコマンドラインオプションによりMQTTブローカーへデータを流せます。
MQTTアドレスに何も指定しなければ送信しません。

必要な場合はSSLの証明書やUsername,Passwordを指定することができます。

|オプション|意味|
|----|----|
|-m,--mqtt-address=`STRING`|MQTTブローカーURL (例: `tcp://mosquitto:1883`, `ssl://mosquitto:8883`)|
|-t, --topic=`STRING`|Publish topic (例: `sensors/co2`)|
|-c, --client-id=`STRING`|MQTT Client ID `default: chissoku`|
|-q, --qos=`INT`|publish QoS `default: 0`|
|--cafile=`STRING`|SSL Root CA|
|--cert=`STRING`|SSL Client Certificate|
|--key=`STRING`|SSL Client Private Key|
|-u, --username=`STRING`|MQTT v3.1/3.1.1 Authenticate Username|
|-p, --password=`STRING`|MQTT v3.1/3.1.1 Authenticate Password|

### Other options

|オプション|意味|
|----|----|
|-n, --no-stdout|標準出力に出力しない|
|-i, --interval=`INT`|出力間隔(n秒) `default: 60`|
|--quiet|標準エラーの出力をしない|
|--tags=`TAG,...`|出力するJSONに `tags` フィールドを追加する(コンマ区切り文字列)|
|-h, --help|オプションヘルプを表示する|
|-v, --version|バージョン情報を表示する|

## Tips

MQTTがうまく動かなければ標準出力を [mosquitto_pub](https://mosquitto.org/man/mosquitto_pub-1.html) に渡せばうまくいくかもしれません。