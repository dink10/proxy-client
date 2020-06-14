## Proxy Client

Providing requests using free external proxy servers
(see http://free-proxy-list.appspot.com/proxy.json)

### Usage example

```go
package main

import (
    "fmt"
    "log"

	"github.com/dink10/proxy-client"
)

func main()  {
     client := proxy_client.NewClient(proxy_client.Config{}, nil)
     res, _ := client.DoRequest(
                "https://google.com",
                "GET",
                proxy_client.Options{},
            )
     fmt.Printf("%s", res)
}

```

You could see an example in /example folder.