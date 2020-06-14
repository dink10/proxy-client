## Proxy Client

Providing requests using free external proxy servers
(see http://free-proxy-list.appspot.com/proxy.json)

### Simple Example of usage

```go
import "github.com/dink10/proxy-client"

client := proxy_client.NewClient(proxy_client.Config{}, logger)
res, _ = client.DoRequest(
			"https://google.com",
			"GET",
			proxy_client.Options{},
		)
fmt.Printf("%s", res)
```

You could see an example in /example folder.