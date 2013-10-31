package yuebao

import (
    "fmt"
    //"strconv"
    "testing"
)

var err error

func Test_GrabLatest(t *testing.T) {
    fmt.Println("\nTesting GrabLatest()...")
    fmt.Printf("Try to grab latest yuebao data.\n")
    err := GrabLatest()
    if err != nil {
        fmt.Println(err)
        t.Error(err)
    }
}

func Test_Get(t *testing.T) {
    fmt.Println("\nTesting Get()...")
    str := Get("2013-10-30")
    fmt.Println(str)
}

func Test_GetRange(t *testing.T) {
    fmt.Println("\nTesting GetRange()...")
    str := GetRange("2013-05-30", "2013-10-30")
    fmt.Println(str)
}
