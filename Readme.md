# go-sclang
go-sclang is a library for go which enables to execute and comunicate with SuperCollider programming language (sclang).

## How to Install

    $ go get github.com/kn1kn1/go-sclang/sclang

## How to use
```go
import (
	"github.com/kn1kn1/go-sclang/sclang"
	"os"
)

const PathToSclang = "/Applications/SuperCollider/SuperCollider.app/Contents/Resources/"
var stdoutWriter io.Writer = os.Stdout
sclangObj, err = sclang.Start(PathToSclang, &stdoutWriter)
```
You may find more usage in ./sclang/example/sclang_example.go

## License 
go-sclang is released under the GNU General Public License (GPL) version 3, 
see the file 'COPYING' for more information.
