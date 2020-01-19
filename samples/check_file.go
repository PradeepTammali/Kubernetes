package main

import (
    "fmt"
    "os"
)

func main() {
    if fileExists(os.Getenv("MAPR_TICKETFILE_LOCATION")) {
        fmt.Println("Example file exists")
    } else {
        fmt.Println("Example file does not exist (or is a directory)")
    }
}

func fileExists(filename string) bool {
    info, err := os.Stat(filename)
    if os.IsNotExist(err) {
        return false
    }
    return !info.IsDir()
}
