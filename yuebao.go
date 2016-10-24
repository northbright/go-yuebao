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

// Flage to output debug messages.
var DEBUG = false

var db *levigo.DB = nil
var ro *levigo.ReadOptions = nil
var wo *levigo.WriteOptions = nil
var cache *levigo.Cache = nil // leveldb cache

var def_cache_size int = 2 * 1024 * 1024 // default leveldb cache size

// Global channel to make leveldb thread safe when GrabXX functions are called in goroutines.
var ch_writer = make(chan int, 1)

// Config File
var config_file = "./config.json"

var db_path = ""
var latest_url = ""
var latest_pattern = ""
var history_url = ""
var history_pattern = ""

// Default Settings
var def_db_path = "./my.db"
var def_latest_url = "http://www.thfund.com.cn/column.dohsmode=searchtopic&pageno=0&channelid=2&categoryid=2435&childcategoryid=2436.htm"
var def_latest_pattern = "<td>(?P<date>\\d{4}-\\d{2}-\\d{2})</td>\\n\\s*<td><span>(?P<earn>\\d*\\.\\d{4})</span></td>\\n\\s*<td><span>(?P<percent>\\d*\\.\\d*)"
var def_history_url = "http://www.thfund.com.cn/website/hd/zlb/newzlbrev2.jsp"
var def_history_pattern = "<td>(?P<date>\\d{4}-\\d{2}-\\d{2})</td>\\r\\n\\s*<td>(?P<earn>\\d*\\.\\d{4})</td>\\r\\n\\s*<td>(?P<percent>\\d*\\.\\d*)"

var def_min_date = "2013-05-30" // yuebao(zenglibao) started from 2013-05-30

// Used to lock goroutine to write into leveldb to make thread safe.
func Lock(ch chan int) {
	ch <- 1
}

// Used to unlock goroutine to write into leveldb to make thread safe.
func UnLock(ch chan int) {
	<-ch
}

// Validate input date string.
// Date string must:
// 1. in yyyy-mm-dd format
// 2. > def_min_date(2013-05-30)
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

	if date > today || date < def_min_date {
		return false
	}
	return true
}

// Save into leveldb database from matched string slice by grabbing data from website.
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

	Lock(ch_writer) // write lock for leveldb if function is called in different goroutines.
	err = db.Put(wo, []byte(date), []byte(jsonStr))
	UnLock(ch_writer) // unlock
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

// Grab latest yuebao data from tianhong fund website and save into leveldb database.
// It reads the "latest_url" and "latest_pattern" settings from config file(./config.json).
func GrabLatestData() (err error) {
	res, err := http.Get(latest_url)
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

	re := regexp.MustCompile(latest_pattern)
	matches := re.FindStringSubmatch(s)

	return SaveFromRegexpMatches(matches)
}

// Grab all history yuebao data from tianhong fund website and save into leveldb database.
// It reads the "history_url" and "history_pattern" settings from config file(./config.json).
func GrabHistoryData() (err error) {
	res, err := http.Get(history_url)
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
	re := regexp.MustCompile(history_pattern)
	matches := re.FindAllStringSubmatch(s, -1)

	for i := 0; i < len(matches); i++ {
		if err = SaveFromRegexpMatches(matches[i]); err != nil {
			return err
		}
	}

	return nil
}

// Get data from day start to day end.
// param: dateBegin, dayEnd in "yyyy-mm-dd" format.
// return: json array if data exist or "" if no data found. Ex:
// [
//   {"d":"2013-07-22","y":1.1547,"r":4.447},
//   {"d":"2013-07-21","y":1.1962,"r":4.471}
// ]
// d: -> date, y -> yield(每万份收益), r -> yield rate(7天年化收益率)
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

// Get yuebao data by date.
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

// Open leveldb database.
// It reads "db_path" in config file(./config.json). The default value is "./my.db".
func OpenDB() (err error) {
	cache = levigo.NewLRUCache(def_cache_size)
	if cache == nil {
		return errors.New("levigo.NewLRUCache() == nil")
	}
	opts := levigo.NewOptions()
	opts.SetCache(cache)
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

// Close leveldb instance.
func CloseDB() {
	if cache != nil {
		cache.Close()
	}

	if db != nil {
		db.Close()
	}
}

// Load default settings
func LoadDefConfig() {
	db_path = def_db_path
	latest_url = def_latest_url
	latest_pattern = def_latest_pattern
	history_url = def_history_url
	history_pattern = def_history_pattern
}

// Load settings from config file.

func LoadConfig() {
	buffer, err := ioutil.ReadFile(config_file)
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

	db_path = obj.Get("db_path").MustString(def_db_path)
	latest_url = obj.Get("latest_url").MustString(def_latest_url)
	latest_pattern = obj.Get("latest_pattern").MustString(def_latest_pattern)
	history_url = obj.Get("history_url").MustString(def_history_url)
	history_pattern = obj.Get("history_pattern").MustString(def_history_pattern)

	fmt.Println("Settings: \n================================")
	fmt.Println("db_path: " + db_path)
	fmt.Println("latest_url: " + latest_url)
	fmt.Println("latest_pattern: " + latest_pattern)
	fmt.Println("history_url: " + history_url)
	fmt.Println("history_pattern: " + history_pattern)
}

func init() {
	LoadConfig()
	OpenDB()
}
