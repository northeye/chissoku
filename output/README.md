# How to implement custom Outputter

独自のOutputter(出力メソッド)を実装する方法です

## ソースファイルの準備

たとえば AWS Kinesis に出力するような Outputter として `Kinesis` を実装してみます。

`output/kinesis.go` にソースファイルを用意します。
```go
package output

import (
	"log/slog"
	"reflect"
	"strings"

	"github.com/northeye/chissoku/options"
)

// Kinesis
// コマンドラインオプションは Kong のAPIに準じます https://github.com/alecthomas/kong
type Kinesis struct {
	Base // Base を埋め込むことで最低限の構成を用意できます

	AccessKeyID     string `help:"AWS Access Key ID" required:"" env:"AWS_ACCESS_KEY_ID"`
	SecretAccessKey string `help:"AWS Secret Secret Key" required:"" env:"AWS_SECRET_ACCESS_KEY"`
}

// Name
func (k *Kinesis) Name() string {
	return strings.ToLower(reflect.TypeOf(k).Elem().Name())
}

// Initialize initialize outputter
func (k *Kinesis) Initialize(ctx context.Context) (_ error) {
	k.Base.Initialize(ctx)

	// データレシーバの初期化ルーチンを書く

	go func() {
		for d := range k.r {
			// データ(d)処理ルーチンを書く
			slog.Debug("Output", "outputter", k.Name(), "data", d)
		}
	}()
	return
}
```

`Name()` は埋め込みのままだと `base` になってしまうので実装する必要があります。

`Output()` メソッドは `Base` に最低限で実装されているのでチャンネルで受け取る形で十分であれば実装する必要はありませんが、 `Interval` オプションが不要な場合など `Base` を埋め込まない場合は実装する必要があります。

### プログラム本体に追加する

`main.go` 内の `Chissoku` 構造体メンバに追加します。

```go
type Chissoku struct {
	// Options
	Options options.Options `embed:""`

	// Stdout output
	output.Stdout `prefix:"stdout." group:"Stdout Output:"`
	// MQTT output
	output.Mqtt `prefix:"mqtt." group:"MQTT Output:"`
    // Kinesis output
    output.Kinesis `prefix:"kinesis." group:"Kinesis Output:"`

    // ...
}
```

### help を確認してみる

```console
$ go run . --help
...
Kinesis Output:
  --kinesis.interval=60                 interval (second) for output. default: '60'
  --kinesis.access-key-id=STRING        AWS Access Key ID ($AWS_ACCESS_KEY_ID)
  --kinesis.secret-access-key=STRING    AWS Secret Secret Key ($AWS_SECRET_ACCESS_KEY)
```


## context

`ctx context.Context` には 以下のValueが埋め込まれています。

| Key | Value | 説明 |
|-----|----|----|
|`options.ContextKeyOptions{}`|`*options.Options{}`|グローバルオプション構造体のポインタ|

### メインループの終了を受け取る

`ctx.Done()` を受信することによりメインループの終了を知ることができます。

### outputter 側から自身を無効化する

`ctx` と自身のポインタを引数として `output.deactivate()` に渡すことで、メインループから自身を対象外にすることができます。<br>
その際、メインループ側から `Close()` は呼ばれません。<br>
ただし、メインループが終了する場合は全ての `Outputter` に `ctx.Done()` を通じて通知し、`outputter.Close()` を呼びます。<br>
すなわち、 `Close()` メソッドはアトミック且つ冪等に実装する必要があります。<br>
`sync` パッケージの `sync.OnceFunc` や `sync.Once` を使うと便利です。

```go
func (o *foo) Initialize(ctx context.Context) error {
	o.close = sync.OnceFunc(func () {
		// deactivate foo
		deactivate(ctx, o)
		// close and release something...
		close(o.r)
	})

	// receiver loop
	go func () {
		for {
			select {
			case <-ctx.Done():
				o.Close()
			case d := <-o.r
				// ...
			}
		}
	}()
}

func (o *foo) Close() {
	o.close()
}
```


以上です。

いいものができましたらPRを送ってください。