package yuebao

import (
    "fmt"
    "io/ioutil"
    "net/http"
    "regexp"
    "strconv"
    //"encoding/json"
    //"time"
    //"strings"
    "github.com/jmhodges/levigo"
    //"github.com/bitly/go-simplejson"
    //"errors"
    // HTML of Tianhong Fund Website is encoded by GBK. Use mahonia to decode the html to UTF-8 if needed.
    //"code.google.com/p/mahonia"
)

var DEBUG = false

var db_path = "./my.db"
var db *levigo.DB = nil
var ro *levigo.ReadOptions = nil
var wo *levigo.WriteOptions = nil

// The latest yuebao data is pressed on Tianhong Fund's website.
var url = "http://www.thfund.com.cn/column.dohsmode=searchtopic&pageno=0&channelid=2&categoryid=2435&childcategoryid=2436.htm"
var patterm = `<td>(?P<date>\d{4}-\d{2}-\d{2})</td>\n\s*<td><span>(?P<earn>\d*\.\d{4})</span></td>\n\s*<td><span>(?P<percent>\d*\.\d*)`

// Grab latest yuebao data from tianhong fund website and save into leveldb database.
func GrabLatest() (err error) {
    res, err := http.Get(url)
    if err != nil {
        fmt.Println(err)
        return err
    }
    defer res.Body.Close()

    if DEBUG {
        fmt.Println(res)
    }

    body, err := ioutil.ReadAll(res.Body)
    if err != nil {
        fmt.Println(err)
        return err
    }

    // HTML of Tianhong Fund Website is encoded by GBK. Use mahonia to decode the html to UTF-8 if needed.
    // ---------------------------------------------------------------------------------------------------
    //decoder := mahonia.NewDecoder("gbk")
    //s := decoder.ConvertString(string(body))

    s := string(body)
    if DEBUG {
        fmt.Print(s)
    }

    re := regexp.MustCompile(patterm)
    matches := re.FindStringSubmatch(s)

    if len(matches) != 4 {
        fmt.Println("Data not found.")
        return err
    }

    date := matches[1]
    yield, _ := strconv.ParseFloat(matches[2], 32)
    yieldRate, _ := strconv.ParseFloat(matches[3], 32)

    jsonStr := fmt.Sprintf("\"y\":%.4f,\"r\":%.4f", yield, yieldRate)

    fmt.Printf("key = %s, value = %s\n", date, jsonStr)

    if Get(date) != "" {
        fmt.Printf("date: %s already grabbed.\n", date)
        return nil
    }

    err = db.Put(wo, []byte(date), []byte(jsonStr))
    if err != nil {
        fmt.Println(err)
        return err
    }

    return nil
}

// Get data from day start to day end.
// param: dateBegin, dayEnd in "yyyy-mm-dd" format.
// return: json array. Ex:
// [
//   {"d":"2013-07-22","y":1.1547,"r":4.4470},
//   {"d":"2013-07-21","y":1.1962,"r":4.4710}
// ]
// d -> date, y -> yield(每万份收益), r -> yield rate(7天年化收益率)
func GetRange(dateBegin, dateEnd string) (jsonStr string) {
    it := db.NewIterator(ro)
    defer it.Close()

    it.Seek([]byte(dateBegin))
    i := 0
    jsonStr = "[\n";
    for it = it; it.Valid(); it.Next() {
        //fmt.Printf("k = %s, v = %s\n", string(it.Key()), string(it.Value()))
        s := fmt.Sprintf("  {\"d\":\"%s\",%s}", string(it.Key()), string(it.Value()))
        //fmt.Println(s)
        jsonStr += s
        if string(it.Key()) == dateEnd {
            jsonStr += "\n"
            break
        }else {
            jsonStr += ",\n"
        }
        i++
    }

    jsonStr += "]\n"

    //fmt.Println(jsonStr)

    return jsonStr
}

// Get yuebao data by date.
// param: date in "yyyy-mm-dd" format.
// return: json string. Ex:
// {"d":"2013-07-22","y":1.1547,"r":4.4470}
// d -> date, y -> yield(每万份收益), r -> yield rate(7天年化收益率)
func Get(date string) string {
    v, _ := db.Get(ro, []byte(date))
    s := fmt.Sprintf("{\"d\":\"%s\",%s}", date, string(v))

    return s
}

func OpenDB() (err error){
    opts := levigo.NewOptions()
    opts.SetCache(levigo.NewLRUCache(3<<30))
    opts.SetCreateIfMissing(true)

    db, err = levigo.Open(db_path, opts)
    if err != nil {
        fmt.Println(err)
        return err
    }

    ro = levigo.NewReadOptions()
    wo = levigo.NewWriteOptions()

    return err
}

func CloseDB() {
    if db != nil {
        db.Close()
    }
}

func init() {
    OpenDB()
}
