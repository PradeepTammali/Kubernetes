package main

import (
        "bytes"
        "fmt"
        "log"
        "os"
        "os/exec"
)

func main() {
    // var password string = "password"
    //cmd := exec.Command("maprlogin password -cluster ${MAPR_CLUSTER} -user ${MAPR_CONTAINER_USER}")
    // cmd := exec.Command("maprlogin", "password","-user", os.Getenv("MAPR_CONTAINER_USER"))
    cmd := exec.Command("cat", os.Getenv("MAPR_TICKETFILE_LOCATION"))
    // cmd.Stdin = strings.NewReader(password)
    var out bytes.Buffer
    cmd.Stdout = &out
    err := cmd.Run()
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf(out.String())
}
