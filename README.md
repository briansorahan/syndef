# syndef
simple command line program to read supercollider synthdef files and output json, xml, or dot format

# usage

First install the latest version of [go](https://golang.org/dl/).

```shell
go get github.com/briansorahan/syndef
syndef -format=dot MySynthDef.scsyndef >MySynthDef.dot
dot -Tsvg MySynthDef.dot >MySynthDef.svg
```
