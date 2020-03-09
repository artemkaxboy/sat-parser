# go-hocon [![Codacy Badge](https://api.codacy.com/project/badge/Grade/87d316c786f2459ca6eb8429e29d9b09)](https://www.codacy.com/manual/artemkaxboy/go-hocon?utm_source=github.com&amp;utm_medium=referral&amp;utm_content=artemkaxboy/go-hocon&amp;utm_campaign=Badge_Grade) [![Coverage Status](https://coveralls.io/repos/github/artemkaxboy/go-hocon/badge.svg?branch=master)](https://coveralls.io/github/artemkaxboy/go-hocon?branch=master) [![Build Status](https://travis-ci.com/artemkaxboy/go-hocon.svg?branch=master)](https://travis-ci.com/artemkaxboy/go-hocon)

The library allows to parse [HOCON](https://github.com/typesafehub/config/blob/master/HOCON.md) configurations to
structures for Golang, using [go-akka/configuration](https://github.com/go-akka/configuration/blob/master/README.md).

## How to use
### 1. Make configuration file
Write your properties in a file and make it readable for your application. It might look like this:
```hocon
{
  Greeting: HELLO WORLD

  advert {
    url: "https://github.com/artemkaxboy/go-hocon"
//    enabled: no
  }

  numbers: {
    first: 3
    second: 2
    sum: 5
    quotient: 1.5
    product: 1
  }
}
```

### 2. Make struct to accept configuration
In your application create a struct with special tags `hocon` that describe how exactly to map values.
It may contain the the following keys:
* `path` is a full path to the struct or field
* `node` is a name of struct or field which does not include parent path
* `default` is a default value of field which will be used if it is not found in conf file
```go
type properties struct {
	Greeting string
	LogLevel string `hocon:"path=logLevel,default=debug"`

	AdvertURL     string `hocon:"path=advert.url"`
	AdvertEnabled bool   `hocon:"path=advert.enabled,default=true"`

	Add struct {
		Add1 int32 `hocon:"node=first"`
		Add2 int32 `hocon:"node=second"`
		Sum  int32 `hocon:"node=sum"`
	} `hocon:"node=numbers"`

	Multi struct {
		Multi1  int64 `hocon:"path=numbers.first"`
		Multi2  int64 `hocon:"path=numbers.second"`
		Product int64 `hocon:"path=numbers.product"`
	}

	Div struct {
		Div1 float32 `hocon:"node=first"`
		Div2 float32 `hocon:"path=numbers.second"`
		Quot float32 `hocon:"node=quotient"`
	} `hocon:"path=numbers"`
}
```
> **_NOTE:_** In case if no path or node tags are provided the name of struct/field will be used to find the value.

> **_NOTE:_** In case if no value or default value are provided the configuration won't be parsed.

### 3. Parse configuration file
Pass file path and struct pointer to LoadConfigFile function
```go
import "github.com/artemkaxboy/go-hocon"
...
    var props properties
    hocon.LoadConfigFile("hocon.conf", &props)
```

### 4. Use your properties
Use your struct where you need it
```go
log.Println(props.Greeting)
log.Printf("log level is %s", props.LogLevel)

isSum := props.Add.Add1+props.Add.Add2 == props.Add.Sum
log.Printf("%d + %d = %d is %t", props.Add.Add1, props.Add.Add2, props.Add.Sum, isSum)

isProduct := props.Multi.Multi1+props.Multi.Multi2 == props.Multi.Product
log.Printf("%d * %d = %d is %t", props.Multi.Multi1, props.Multi.Multi2, props.Multi.Product, isProduct)

isDivision := props.Div.Div1/props.Div.Div2 == props.Div.Quot
log.Printf("%d / %d = %d is %t", props.Add.Add1, props.Add.Add2, props.Add.Sum, isDivision)

if props.AdvertEnabled {
    log.Printf("visit %s for more details ...", props.AdvertURL)
}
```
---
You may find full example here: [go-hocon-example](https://github.com/artemkaxboy/go-hocon-example)
