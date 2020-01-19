package main

import (
        "strings"
        "fmt"
)

func main() {
var data string = `Opening a file 
file opened at a location /home/pradeep/file`
s := strings.Split(data, "\n")
fmt.Println(s[1])
}
