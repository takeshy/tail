# tail
tail -F by go lang

## Usage

```go
package main
import "github.com/takeshy/tail"
import "fmt"

func main(){
  c := tail.Watch("/path/to/file")
  for {
    select {
    case s := <-c:
      fmt.Println(s)
    }
  }
}

```

## Installation

```
go get github.com/takeshy/tail
```


## License

MIT

## Author

Takeshi Morita
