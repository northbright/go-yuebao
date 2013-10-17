package yuebao

import (
    "fmt"
    "strconv"
    "testing"
)

func Test_Get(t *testing.T) {
    if data, err := Get(); err != nil {
        t.Error(err)
    } else {
        fmt.Println("Date: " + data.Date)
        fmt.Println("Yield: " + strconv.FormatFloat(float64(data.Yield), 'f', -1, 32))
        fmt.Println("YieldRate: " + strconv.FormatFloat(float64(data.YieldRate), 'f', -1, 32) + "%")
    }
}
