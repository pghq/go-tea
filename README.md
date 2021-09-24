# go-museum

SDK for building go apps within the organization.

## Installation

go-museum may be installed using the go get command:

```
go get github.com/pghq/go-museum
```
## Usage

```
import "github.com/pghq/go-museum/museum"
```

To create a new app:

```
app, err := museum.New()
if err != nil{
    panic(err)
}
```
