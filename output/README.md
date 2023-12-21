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
func (k *Kinesis) Initialize(_ *options.Options) (_ error) {
	k.Base.Initialize(nil)

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

以上です。

いいものができましたらPRを送ってください。