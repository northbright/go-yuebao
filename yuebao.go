// Package yuebao grabs the latest or history yuebao data from tianhong fund's web site and save them into a leveldb database.
// It also provides query methods to get yuebao data by date or date range.
package yuebao

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/bitly/go-simplejson"
	"github.com/jmhodges/levigo"
	// HTML of Tianhong Fund Website is encoded by GBK. Use mahonia to decode the html to UTF-8 if needed.
	//"code.google.com/p/mahonia"
)

// DEBUG is debug mode to output debug messages.
var DEBUG = false

var db *levigo.DB
var ro *levigo.ReadOptions
var wo *levigo.WriteOptions
var cache *levigo.Cache // leveldb cache

var defCacheSize int = 2 * 1024 * 1024 // default leveldb cache size

// Global channel to make leveldb thread safe when GrabXX functions are called in goroutines.
var chWriter = make(chan int, 1)

// Config File
var configFile = "./config.json"

var dbPath = ""
var latestURL = ""
var latestPattern = ""
var historyURL = ""
var historyPattern = ""

// Default Settings
var defDBPath = "./my.db"
var defLatestURL = "http://www.thfund.com.cn/column.dohsmode=searchtopic&pageno=0&channelid=2&categoryid=2435&childcategoryid=2436.htm"
var defLatestPattern = "<td>(?P<date>\\d{4}-\\d{2}-\\d{2})</td>\\n\\s*<td><span>(?P<earn>\\d*\\.\\d{4})</span></td>\\n\\s*<td><span>(?P<percent>\\d*\\.\\d*)"
var defHistoryURL = "http://www.thfund.com.cn/website/hd/zlb/newzlbrev2.jsp"
var defHistoryPattern = "<td>(?P<date>\\d{4}-\\d{2}-\\d{2})</td>\\r\\n\\s*<td>(?P<earn>\\d*\\.\\d{4})</td>\\r\\n\\s*<td>(?P<percent>\\d*\\.\\d*)"

var defMinDate = "2013-05-30" // yuebao(zenglibao) started from 2013-05-30

// Lock locks goroutine to write into leveldb to make thread safe.
func Lock(ch chan int) {
	ch <- 1
}

// UnLock unlocks goroutine to write into leveldb to make thread safe.
func UnLock(ch chan int) {
	<-ch
}

// IsDateValid validates input date string.
// Date string must:
// 1. in yyyy-mm-dd format
// 2. > defMinDate(2013-05-30)
// 3. <= today
func IsDateValid(date string) bool {
	if len(date) == 0 {
		return false
	}

	p := `^\d{4}-\d{2}-\d{2}$`
	re := regexp.MustCompile(p)
	matches := re.FindStringSubmatch(date)
	if len(matches) != 1 {
		return false
	}

	t := time.Now()
	today := fmt.Sprintf("%04d-%02d-%02d", t.Year(), t.Month(), t.Day())

	if date > today || date < defMinDate {
		return false
	}
	return true
}

// SaveFromRegexpMatches saves into leveldb database from matched string slice by grabbing data from website.
func SaveFromRegexpMatches(matches []string) (err error) {
	// len(matches) = 4: entile string, date, yield, yield rate
	if len(matches) != 4 {
		fmt.Println("Data not found.")
		return err
	}

	date := matches[1]
	yield, _ := strconv.ParseFloat(matches[2], 32)
	yieldRate, _ := strconv.ParseFloat(matches[3], 32)

	jsonStr := fmt.Sprintf("\"y\":%.4f,\"r\":%.3f", yield, yieldRate)

	if DEBUG {
		fmt.Printf("key = %s, value = %s\n", date, jsonStr)
	}

	s := GetData(date)
	if s != "" {
		fmt.Printf("date: %s already grabbed. data = %s\n", date, s)
		return nil
	}

	Lock(chWriter) // write lock for leveldb if function is called in different goroutines.
	err = db.Put(wo, []byte(date), []byte(jsonStr))
	UnLock(chWriter) // unlock
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

// GrabLatestData grabs latest yuebao data from tianhong fund website and save into leveldb database.
// It reads the "latestURL" and "latestPattern" settings from config file(./config.json).
func GrabLatestData() (err error) {
	res, err := http.Get(latestURL)
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

	re := regexp.MustCompile(latestPattern)
	matches := re.FindStringSubmatch(s)

	return SaveFromRegexpMatches(matches)
}

// GrabHistoryData grabs all history yuebao data from tianhong fund website and save into leveldb database.
// It reads the "historyURL" and "historyPattern" settings from config file(./config.json).
func GrabHistoryData() (err error) {
	res, err := http.Get(historyURL)
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

	s := string(body)
	if DEBUG {
		fmt.Print(s)
	}
	re := regexp.MustCompile(historyPattern)
	matches := re.FindAllStringSubmatch(s, -1)

	for i := 0; i < len(matches); i++ {
		if err = SaveFromRegexpMatches(matches[i]); err != nil {
			return err
		}
	}

	return nil
}

// GetDataByRange gets data from day start to day end.
//
//     param: dateBegin, dayEnd in "yyyy-mm-dd" format.
//     return: json array if data exist or "" if no data found. Ex:
//     [
//       {"d":"2013-07-22","y":1.1547,"r":4.447},
//       {"d":"2013-07-21","y":1.1962,"r":4.471}
//     ]
//     d: -> date, y -> yield(每万份收益), r -> yield rate(7天年化收益率)
func GetDataByRange(dateBegin, dateEnd string) (jsonStr string) {
	it := db.NewIterator(ro)
	defer it.Close()

	if (!IsDateValid(dateBegin)) || (!IsDateValid(dateEnd)) {
		return ""
	}

	it.Seek([]byte(dateBegin))
	i := 0
	jsonStr = "[\n"
	for ; it.Valid(); it.Next() {
		s := fmt.Sprintf("  {\"d\":\"%s\",%s}", string(it.Key()), string(it.Value()))
		jsonStr += s
		if string(it.Key()) == dateEnd {
			jsonStr += "\n"
			break
		} else {
			jsonStr += ",\n"
		}
		i++
	}

	jsonStr += "]\n"

	return jsonStr
}

// GetData gets yuebao data by date.
// param: date in "yyyy-mm-dd" format.
// return: json string if data exist or "" if no data found.  Ex:
// {"d":"2013-07-22","y":1.1547,"r":4.447}
// d -> date, y -> yield(每万份收益), r -> yield rate(7天年化收益率)
func GetData(date string) string {
	if !IsDateValid(date) {
		return ""
	}

	v, _ := db.Get(ro, []byte(date))
	if len(v) == 0 {
		return ""
	}

	s := fmt.Sprintf("{\"d\":\"%s\",%s}", date, string(v))
	return s
}

// OpenDB opens leveldb database.
// It reads "dbPath" in config file(./config.json). The default value is "./my.db".
func OpenDB() (err error) {
	cache = levigo.NewLRUCache(defCacheSize)
	if cache == nil {
		return errors.New("levigo.NewLRUCache() == nil")
	}
	opts := levigo.NewOptions()
	opts.SetCache(cache)
	opts.SetCreateIfMissing(true)

	db, err = levigo.Open(dbPath, opts)
	if err != nil {
		fmt.Println(err)
		return err
	}

	ro = levigo.NewReadOptions()
	wo = levigo.NewWriteOptions()

	return err
}

// CloseDB closes leveldb instance.
func CloseDB() {
	if cache != nil {
		cache.Close()
	}

	if db != nil {
		db.Close()
	}
}

// LoadDefConfig loads default settings
func LoadDefConfig() {
	dbPath = defDBPath
	latestURL = defLatestURL
	latestPattern = defLatestPattern
	historyURL = defHistoryURL
	historyPattern = defHistoryPattern
}

// LoadConfig loads settings from config file.
func LoadConfig() {
	buffer, err := ioutil.ReadFile(configFile)
	if err != nil {
		fmt.Println(err)
		fmt.Println("Load default settings.")
		LoadDefConfig()
		return
	}

	if DEBUG {
		fmt.Println(len(buffer))
		fmt.Println(string(buffer))
	}

	obj, err := simplejson.NewJson([]byte(buffer))
	if err != nil {
		fmt.Println(err)
		fmt.Println("Load default settings.")
		LoadDefConfig()
		return
	}

	dbPath = obj.Get("dbPath").MustString(defDBPath)
	latestURL = obj.Get("latestURL").MustString(defLatestURL)
	latestPattern = obj.Get("latestPattern").MustString(defLatestPattern)
	historyURL = obj.Get("historyURL").MustString(defHistoryURL)
	historyPattern = obj.Get("historyPattern").MustString(defHistoryPattern)

	fmt.Println("Settings: \n================================")
	fmt.Println("dbPath: " + dbPath)
	fmt.Println("latestURL: " + latestURL)
	fmt.Println("latestPattern: " + latestPattern)
	fmt.Println("historyURL: " + historyURL)
	fmt.Println("historyPattern: " + historyPattern)
}

func init() {
	LoadConfig()
	OpenDB()
}
