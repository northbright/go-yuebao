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
    s := []string{"2013-04-01", "2013-10-30", "2014-10-02"}
    for _, v := range s {
        fmt.Printf("%s, data = %s\n", v, GetData(v))
    }
}

func Test_GetDataByRange(t *testing.T) {
    fmt.Println("\nTesting GetDataByRange()...")
    tm := time.Now()
    today := fmt.Sprintf("%04d-%02d-%02d", tm.Year(), tm.Month(), tm.Day())
    fmt.Printf("Range: 2013-05-30 - %s\n", today)
    str := GetDataByRange("2013-05-30", today)
    fmt.Println(str)
    fmt.Printf("Range: 2012-01-01 - 2013-04-30\n")
    str =  GetDataByRange("2012-01-01", "2013-04-30")
    fmt.Println(str)
}

func Test_IsDateValid(t *testing.T) {
    fmt.Println("\nTesting IsDateValid()...")
    s := []string{"2013-4-20", "July, 4, 1999", "2013-05-29", "2013-05-30", "2015-01-01"}
    tm := time.Now()
    s = append(s, fmt.Sprintf("%04d-%02d-%02d", tm.Year(), tm.Month(), tm.Day()))  // today
    s = append(s, fmt.Sprintf("%04d-%02d-%02d", tm.Year(), tm.Month(), tm.Day() + 1))  // tomorrow
    for _, v := range s {
        fmt.Printf("date = %s, valid = %v\n", v, IsDateValid(v))
    }
}

