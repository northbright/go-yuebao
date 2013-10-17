package yuebao

import (
    "fmt"
    "io/ioutil"
    "net/http"
    "regexp"
    "strconv"
    // HTML of Tianhong Fund Website is encoded by GBK. Use mahonia to decode the html to UTF-8 if needed.
    //"code.google.com/p/mahonia"
)

type YuebaoData struct {
    Date string  // date in string format(yyyy-mm-dd, ex: 2013-10-16)
    Yield float32 // yield per 10000 units / day(每万份收益/天)
    YieldRate float32 // 7-day annual yield rate(七天年化收益率)
}

var DEBUG = false

// The latest yuebao data is pressed on Tianhong Fund's website.
var url = "http://www.thfund.com.cn/column.dohsmode=searchtopic&pageno=0&channelid=2&categoryid=2435&childcategoryid=2436.htm"

func Get() (data *YuebaoData, err error) {
    res, err := http.Get(url)
    if err != nil {
        fmt.Println(err)
        return nil, err
    }
    defer res.Body.Close()

    if DEBUG {
        fmt.Println(res)
    }

    body, err := ioutil.ReadAll(res.Body)
    if err != nil {
        fmt.Println(err)
        return nil, err
    }

    // HTML of Tianhong Fund Website is encoded by GBK. Use mahonia to decode the html to UTF-8 if needed.
    // ---------------------------------------------------------------------------------------------------
    //decoder := mahonia.NewDecoder("gbk")
    //s := decoder.ConvertString(string(body))

    s := string(body)
    if DEBUG {
        fmt.Print(s)
    }

    p := `<td>(?P<date>\d{4}-\d{2}-\d{2})</td>\n\s*<td><span>(?P<earn>\d*\.\d{4})</span></td>\n\s*<td><span>(?P<percent>\d*\.\d*)`
    re := regexp.MustCompile(p)
    matches := re.FindStringSubmatch(s)

    if len(matches) != 4 {
        fmt.Println("Data not found.")
        return nil, err
    }

    data = new(YuebaoData)
    data.Date = matches[1]
    yield, _ := strconv.ParseFloat(matches[2], 32)
    yieldRate, _ := strconv.ParseFloat(matches[3], 32)
    data.Yield = float32(yield)
    data.YieldRate = float32(yieldRate)

    return data, nil
}
