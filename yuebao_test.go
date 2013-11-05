package yuebao

import (
    "fmt"
    //"strconv"
    "testing"
    "time"
)

var err error

func Test_GrabLatestData(t *testing.T) {
    fmt.Println("\nTesting GrabLatestData()...")
    fmt.Printf("Try to grab latest yuebao data.\n")
    err := GrabLatestData()
    if err != nil {
        fmt.Println(err)
        t.Error(err)
    }
}

func Test_GrabHistoryData(t *testing.T) {
    fmt.Println("\nTesting GrabHistoryData()...")
    fmt.Printf("Try to grab history yuebao data.\n")
    err := GrabHistoryData()
    if err != nil {
        fmt.Println(err)
        t.Error(err)
    }
}

func Test_GetData(t *testing.T) {
    fmt.Println("\nTesting GetData()...")
    str := GetData("2013-10-30")
    fmt.Println(str)
}

func Test_GetDataByRange(t *testing.T) {
    fmt.Println("\nTesting GetDataByRange()...")
    tm := time.Now()
    today := fmt.Sprintf("%04d-%02d-%02d", tm.Year(), tm.Month(), tm.Day())
    str := GetDataByRange("2013-05-30", today)
    fmt.Println(str)
}
