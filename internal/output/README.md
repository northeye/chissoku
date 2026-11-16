# How to implement custom Outputter

独自のOutputter(出力メソッド)を実装する方法です

## ソースファイルの準備

たとえば AWS Kinesis に出力するような Outputter として `Kinesis` を実装してみます。

`internal/output/kinesis.go` にソースファイルを用意します。
```go
package output

import (
	"context"
	"log/slog"
)

// Kinesis
// コマンドラインオプションは Kong のAPIに準じます https://github.com/alecthomas/kong
type Kinesis struct {
	Base // Base を埋め込むことで Interval オプション・Output()・Close() が用意されます

	AccessKeyID     string `help:"AWS Access Key ID" required:"" env:"AWS_ACCESS_KEY_ID"`
	SecretAccessKey string `help:"AWS Secret Secret Key" required:"" env:"AWS_SECRET_ACCESS_KEY"`
}

// Name
func (*Kinesis) Name() string {
	return "kinesis"
}

// Initialize initialize outputter
func (k *Kinesis) Initialize(ctx context.Context) error {
	// 接続などの初期化処理を書く
	// エラーを返すとこの Outputter は無効になります

	// 出力ループを開始する
	k.start(ctx, k.put)
	return nil
}

// put データ(b)の出力処理
func (k *Kinesis) put(b []byte) {
	slog.Debug("Output", "outputter", k.Name(), "data", string(b))
}
```

`Name()` は必ず実装してください。

`Base.start(ctx, emit)` は出力ループを goroutine として起動します。
`emit` は最初のデータ受信時に即座に、以降は `Interval` 秒ごとに最新のデータで呼ばれます（`Interval <= 0` の場合は受信のたびに呼ばれます）。
`emit` は出力ループの goroutine 上で呼ばれるので、ネットワークI/Oなどでブロックしても構いません。

`Output()` は `Base` に実装されており、最新のデータを保持するだけでブロックしません。
`Base` を埋め込まない場合は `Output()` と `Close()` も実装する必要があります。`Output()` は決してブロックしてはいけません。

後片付け(コネクションの切断等)が必要な場合は `Close()` を実装し、その中で `Base.Close()` を呼んでください。
`Base.Close()` は出力ループの終了を待つので、その後であれば安全にリソースを解放できます。

```go
func (k *Kinesis) Close() {
	k.Base.Close()
	// release something...
}
```

### プログラム本体に追加する

`internal/chissoku/chissoku.go` の `Chissoku` 構造体メンバに追加します。

```go
type Chissoku struct {
	// Options
	Options Options `embed:""`

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
$ go run ./cmd/chissoku --help
...
Kinesis Output:
  --kinesis.interval=60                 interval (second) for output. default: '60'
  --kinesis.access-key-id=STRING        AWS Access Key ID ($AWS_ACCESS_KEY_ID)
  --kinesis.secret-access-key=STRING    AWS Secret Secret Key ($AWS_SECRET_ACCESS_KEY)
```

## ライフサイクル

- `Initialize(ctx)` に渡される `ctx` はメインループの終了時にキャンセルされます。
- メインループの終了時、稼働中の全ての Outputter の `Close()` が呼ばれます。
- `Close()` は複数回呼ばれても安全(冪等)に実装してください。

### outputter 側から自身を無効化する

`Initialize()` に渡された `ctx` と自身のポインタを `deactivate()` に渡すことで、メインループの出力対象から除外できます。
除外後、メインループ側から `Close()` が呼ばれます。
全ての Outputter が無効になるとプログラムは終了します。

```go
func (k *Kinesis) put(ctx context.Context, b []byte) {
	if err := send(b); err != nil {
		deactivate(ctx, k)
	}
}
```

以上です。

いいものができましたらPRを送ってください。
