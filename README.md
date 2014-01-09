go-yuebao
=========

Get the latest yuebao(Tianhong Fund's "Zenglibao") data from Tianhong Fund's website.

# Install

* Requirements

    go-yuebao requires [levigo](https://github.com/jmhodges/levigo).

    To install levigo, run:

        CGO_CFLAGS="-I/path/to/leveldb/include" CGO_LDFLAGS="-L/path/to/leveldb/lib" go get github.com/jmhodges/levigo

* Install go-yuebao

    The main method of installation is through "go get" (provided in $GOROOT/bin)

        go get github.com/northbright/go-yuebao

# Usage

To use the 'yuebao' package, you'll need the appropriate import statement:

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
    // return: json array if data exist or "" if no data found. Ex:
    // [
    //   {"d":"2013-07-22","y":1.1547,"r":4.447},
    //   {"d":"2013-07-21","y":1.1962,"r":4.471}
    // ]
    // d -> date, y -> yield(每万份收益), r -> yield rate(7天年化收益率)
    func GetDataByRange(dateBegin, dateEnd string) (jsonStr string)

#### GetData()

    // Get yuebao data by date.
    // param: date in "yyyy-mm-dd" format.
    // return: json string if data exist or "" if no data found. Ex:
    // {"d":"2013-07-22","y":1.1547,"r":4.447}
    // d -> date, y -> yield(每万份收益), r -> yield rate(7天年化收益率)
    func GetData(date string) string

# Config File

Make sure the config file "config.json" exists under "./":

    db_path: leveldb database path. Default is "./my.db".
    latest_url: url to grab latest yuebao data.
    latest_pattern: regexp pattern string to grab latest yuebao data.
    history_url: url to grab all history yuebao data.
    history_pattern: regexp pattern string to grab all history yuebao data.

# Test
    run "go test".
