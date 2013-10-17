go-yuebao
=========

Get the latest yuebao(Tianhong Fund's "Zenglibao") data from Tianhong Fund's website.

# Install

The main method of installation is through "go get" (provided in $GOROOT/bin)

    go get github.com/northbright/go-yuebao

# Usage

To use the package, you'll need the appropriate import statement:

    import (
        "github.com/northbright/go-yuebao"
    )

## yuebao data type 

    type YuebaoData struct {
        Date string  // date in string format(yyyy-mm-dd, ex: 2013-10-16)
        Yield float32 // yield per 10000 units / day(每万份收益/天)
        YieldRate float32 // 7-day annual yield rate(七天年化收益率)
    }

## yuebao.Get()

    if data, err := yuebao.Get(); err != nil {
        fmt.Println(err)
    } else {
        fmt.Println("Date: " + data.Date)
        fmt.Println("Yield: " + strconv.FormatFloat(float64(data.Yield), 'f', -1, 32))
        fmt.Println("YieldRate: " + strconv.FormatFloat(float64(data.YieldRate), 'f', -1, 32) + "%")
    }

You may run "go test" and check yuebao_test.go for more information.
