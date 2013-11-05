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

#### GrabLatestData()

    // Grab latest yuebao data from tianhong fund website and save into leveldb database.
    // It reads the "latest_url" and "latest_pattern" settings from config file(./config.json).
    func GrabLatestData() (err error)

#### GrabHistoryData()

    // Grab all history yuebao data from tianhong fund website and save into leveldb database.
    // It reads the "history_url" and "history_pattern" settings from config file(./config.json).
    func GrabHistoryData() (err error)

#### GetDataByRange()

    // Get data from day start to day end.
    // param: dateBegin, dayEnd in "yyyy-mm-dd" format.
    // return: json array. Ex:
    // [
    //   {"d":"2013-07-22","y":1.1547,"r":4.4470},
    //   {"d":"2013-07-21","y":1.1962,"r":4.4710}
    // ]
    // d -> date, y -> yield(每万份收益), r -> yield rate(7天年化收益率)
    func GetDataByRange(dateBegin, dateEnd string) (jsonStr string)

#### GetData()

    // Get yuebao data by date.
    // param: date in "yyyy-mm-dd" format.
    // return: json string. Ex:
    // {"d":"2013-07-22","y":1.1547,"r":4.4470}
    // d -> date, y -> yield(每万份收益), r -> yield rate(7天年化收益率)
    func GetData(date string) string

# Test
    run "go test".