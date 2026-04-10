// Package utils 工具类
package utils

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"math"
	"math/big"
	"mime/multipart"
	"net"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/pkg/log"
	"github.com/sony/sonyflake"
	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

const (
	// MAXNAMELEN 名称最大长度
	MAXNAMELEN = 26

	// GNSPREFIX gns前缀
	GNSPREFIX = "gns://"

	// KeyText aes的加密字符串
	KeyText = "aishu.com12345akljzmknm.ahkjkljl"

	GNSPattern = "^gns:/(/[0-9A-F]{32})+$"
)

// CommonIV 初始化向量
var CommonIV = []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f}

// RemoveRepByMap 通过map主键唯一的特性过滤重复元素
func RemoveRepByMap(slc []string) []string {
	result := []string{}
	tempMap := map[string]byte{} // 存放不重复主键
	for _, e := range slc {
		l := len(tempMap)
		tempMap[e] = 0
		if len(tempMap) != l { // 加入map后，map长度变化，则元素不重复
			result = append(result, e)
		}
	}
	return result
}

// IsContain 字符串是否包含在数组内
func IsContain(target string, arr []string) bool {
	sort.Strings(arr)
	index := sort.SearchStrings(arr, target)
	if index < len(arr) && arr[index] == target {
		return true
	}
	return false
}

// Arrcmp 计算两个数组的差集
func Arrcmp(src, dest []string) ([]string, []string) { //nolint
	msrc := make(map[string]byte) // 按源数组建索引
	mall := make(map[string]byte) // 源+目所有元素建索引
	var set []string              // 交集
	// 1.源数组建立map
	for _, v := range src {
		msrc[v] = 0
		mall[v] = 0
	}
	// 2.目数组中，存不进去，即重复元素，所有存不进去的集合就是并集
	for _, v := range dest {
		l := len(mall)
		mall[v] = 1
		if l != len(mall) { // 长度变化，即可以存
			l = len(mall) //nolint
		} else { // 存不了，进并集
			set = append(set, v)
		}
	}
	// 3.遍历交集，在并集中找，找到就从并集中删，删完后就是补集（即并-交=所有变化的元素）
	for _, v := range set {
		delete(mall, v)
	}
	// 4.此时，mall是补集，所有元素去源中找，找到就是删除的，找不到的必定能在目数组中找到，即新加的
	var added, deleted []string
	for v := range mall {
		_, exist := msrc[v]
		if exist {
			deleted = append(deleted, v)
		} else {
			added = append(added, v)
		}
	}
	return added, deleted
}

// FomatTimeStamp 时间戳转标准格式时间
func FomatTimeStamp(timeStamp int64) string {
	tm := time.Unix(timeStamp, 0)
	return tm.Format("2006-01-02 15:04")
}

// GetLanguage 获取环境变量中的语言资源
func GetLanguage() string {
	language := os.Getenv("SERVICE_LANG")
	pos := strings.Index(language, ".")
	// 不包含.utf-8时直接返回language
	if pos == -1 {
		return language
	}
	// 不然直接删除.后面的内容
	return language[0:pos]
}

// SubString 截取指定长度的字符串，中英文皆适用
func SubString(str string, begin, length int) string {
	rs := []rune(str)
	lth := len(rs)
	if begin < 0 {
		begin = 0
	}
	if begin >= lth {
		begin = lth
	}
	end := begin + length

	if end > lth {
		end = lth
	}
	return string(rs[begin:end])
}

// GetParentDocIDs 根据docID获取所有父级目录的docID
func GetParentDocIDs(docID string) []string {
	var res []string
	id := strings.TrimPrefix(docID, GNSPREFIX)
	ids := strings.Split(id, "/")
	prefix := GNSPREFIX
	for i := 0; i < len(ids)-1; i++ {
		if i == 0 {
			prefix += ids[i]
		} else {
			prefix = prefix + "/" + ids[i]
		}
		res = append(res, prefix)
	}
	return res
}

// CreateAddress 合并成ip:port
func CreateAddress(ip string, port int) (address string) {
	address = ip
	if port != 0 {
		address = fmt.Sprintf("%s:%d", ip, port)
	}
	return
}

// GenDocDestPath 生成文档流转目标路径
func GenDocDestPath(destDir, subDirRule, extPath, userName string) (destPath string) {
	switch subDirRule {
	case "name":
		destPath = fmt.Sprintf("%v/%v", destDir, userName)
	case "date":
		destPath = fmt.Sprintf("%v/%v", destDir, time.Now().Format("2006-01-02"))
	case "name_date":
		destPath = fmt.Sprintf("%v/%v/%v", destDir, userName, time.Now().Format("2006-01-02"))
	case "date_name":
		destPath = fmt.Sprintf("%v/%v/%v", destDir, time.Now().Format("2006-01-02"), userName)
	default:
		destPath = destDir
	}
	if extPath != "" {
		destPath = fmt.Sprintf("%v/%v", destPath, extPath)
	}
	return
}

// ReflectCmp 按传入fieldName 排序
func ReflectCmp(i, j interface{}, order string, params ...string) bool { //nolint
	// 排序非指定顺序，默认指定为升序排列
	if order != "asc" && order != "desc" {
		order = "asc"
	}
	NotDesc := order != "desc"
	// 列表为空时，异常处理,asc 默认升序，desc 默认降序
	if len(params) == 0 {
		return true && NotDesc
	}
	fildName := params[0]
	if len(params) != 0 {
		params = params[1:]
	}
	valI := reflect.ValueOf(i).FieldByName(fildName).Interface()
	valJ := reflect.ValueOf(j).FieldByName(fildName).Interface()
	var result bool
	switch s := valI.(type) {
	case string:
		lager := Compare(s, valJ.(string))
		if lager {
			result = XNOR(true, NotDesc)
		} else if !lager {
			result = XNOR(false, NotDesc)
		} else if lager && len(params) == 0 {
			// 所有字段比较完成 此时所有字段值都相等 返回true
			result = XNOR(true, NotDesc)
		} else {
			result = ReflectCmp(i, j, order, params...)
		}
	case int64:
		t := valJ.(int64)
		if s == -1 {
			s = int64(math.MaxInt64)
		}
		if t == -1 {
			t = int64(math.MaxInt64)
		}
		if s < t {
			result = XNOR(true, NotDesc)
		} else if s > t {
			result = XNOR(false, NotDesc)
		} else if s == t && len(params) == 0 {
			// 所有字段比较完成 此时所有字段值都相等 返回true
			result = XNOR(true, NotDesc)
		} else {
			result = ReflectCmp(i, j, order, params...)
		}
	case int:
		if s < valJ.(int) {
			result = XNOR(true, NotDesc)
		} else if s > valJ.(int) {
			result = XNOR(false, NotDesc)
		} else if s == valJ.(int) && len(params) == 0 {
			// 所有字段比较完成 此时所有字段值都相等 返回true
			result = XNOR(true, NotDesc)
		} else {
			result = ReflectCmp(i, j, order, params...)
		}
	default:
		fmt.Println("type is unknown")
		result = XNOR(true, NotDesc)
	}
	return result
}

// XNOR 同或运算
func XNOR(x, y bool) bool {
	return x && y || !x && !y
}

// Compare 字符串按GBK编码比较
func Compare(a, b string) bool {
	astr, _ := UTF82GBK(a)
	bstr, _ := UTF82GBK(b)
	bLen := len(bstr)
	for idx, chr := range astr {
		if idx > bLen-1 {
			return false
		}
		if chr != bstr[idx] {
			return chr < bstr[idx]
		}
	}
	return true
}

// UTF82GBK utf82转gbk
func UTF82GBK(src string) ([]byte, error) {
	GB18030 := simplifiedchinese.All[0]
	return io.ReadAll(transform.NewReader(bytes.NewReader([]byte(src)), GB18030.NewEncoder()))
}

// GetMD5String 使用MD5对数据进行哈希运算
func GetMD5String(str string) (s string) {
	b := []byte(str)
	sArr := md5.Sum(b)
	s = hex.EncodeToString(sArr[:])
	return
}

// GetParentFilePath 获取父目录路径
func GetParentFilePath(filePath string) (p string) {
	fs := strings.Split(filePath, "/")
	if len(fs) > 1 {
		p = strings.Join(fs[:len(fs)-1], "/")
		return
	}
	return
}

// String2Uint64 string转int64
func String2Uint64(s string) uint64 {
	intNum, _ := strconv.Atoi(s)
	return uint64(intNum)
}

// IsDoclib 根据文档id判断是否为文档库
func IsDoclib(docID string) bool {
	prefix := GNSPREFIX
	if !strings.HasPrefix(docID, prefix) {
		return false
	}
	s := docID[6:]
	splits := strings.Split(s, "/")
	return len(splits) == 1
}

// GetDoclibID 根据文档id截取为文档库ID
func GetDoclibID(docID string) (doclibID string) {
	prefix := GNSPREFIX
	if !strings.HasPrefix(docID, prefix) {
		return
	}
	s := docID[6:]
	splits := strings.Split(s, "/")
	doclibID = prefix + splits[0]
	return
}

// GetDocCurID 根据文档id截取为当前层级ID
func GetDocCurID(docID string) (curID string) {
	prefix := GNSPREFIX
	if !strings.HasPrefix(docID, prefix) {
		return
	}
	s := docID[6:]
	splits := strings.Split(s, "/")
	curID = splits[len(splits)-1]
	return
}

// GetCurDeptID 根据部门path截取为当前层级ID
func GetCurDeptID(path string) (curID string) {
	if path == "" {
		return
	}
	splits := strings.Split(path, "/")
	curID = splits[len(splits)-1]
	return
}

var (
	sf *sonyflake.Sonyflake
)

// NewMachineID 根据ip获取唯一id
// https://github.com/tinrab/makaroni/tree/master/utilities/unique-id
func NewMachineID() func() (uint16, error) {
	return func() (uint16, error) {
		ipStr := os.Getenv("POD_IP")
		if ipStr == "" {
			ipStr = "127.0.0.1"
		}
		ip := net.ParseIP(ipStr)
		ip = ip.To16()
		if ip == nil || len(ip) < 4 {
			return 0, errors.New("invalid IP")
		}
		return uint16(ip[14])<<8 + uint16(ip[15]), nil
	}
}

// GetUniqueID 生成唯一id
// 使用sonyflake获取唯一、自增id
// 传入ip，使用传入的ip作为机器码
// 不传入ip，使用ipv4作为机器码
func GetUniqueID() (uint64, error) {
	return sf.NextID()
}

func GetUniqueIDStr() (string, error) {
	id, err := GetUniqueID()
	if err != nil {
		return "", err
	}
	return strconv.FormatUint(id, 10), nil
}

// 初始化sonyflake
func init() {
	var st sonyflake.Settings
	// st.StartTime = time.Now()  nolint
	st.MachineID = NewMachineID()
	sf = sonyflake.NewSonyflake(st)
	if sf == nil {
		panic("sonyflake not created")
	}
}

// GetEmailTemplate 获取邮件模板
func GetEmailTemplate(path string, data interface{}) (string, error) {
	var buff bytes.Buffer
	apath, _ := os.Getwd()
	language := strings.ToLower(GetLanguage())
	switch language {
	case "zh_cn":
		path = fmt.Sprintf("%s/resource/zh_cn/%s", apath, path)
	case "zh_tw":
		path = fmt.Sprintf("%s/resource/zh_tw/%s", apath, path)
	case "en_us":
		path = fmt.Sprintf("%s/resource/en_us/%s", apath, path)
	}

	htmlTemplate, err := template.ParseFiles(path)
	if err != nil {
		return "", err
	}

	err = htmlTemplate.Execute(&buff, data)
	if err != nil {
		return "", err
	}

	content := buff.String()
	return content, nil
}

// RetryTimes 重试,限制次数
func RetryTimes(name string, tryTimes int, sleep time.Duration, callback func(retryCount int) error) (err error) {
	for i := 1; i <= tryTimes; i++ {
		err = callback(i)
		if err == nil {
			return nil
		}
		time.Sleep(sleep)
	}
	return err
}

// CutName 截取任务名称超过26个字符后的内容
func CutName(name string) string {
	tmpSubject := name
	if len(tmpSubject) > MAXNAMELEN {
		tmpSubject = fmt.Sprintf("%s...", SubString(tmpSubject, 0, MAXNAMELEN))
	}
	return tmpSubject
}

// GetIntersection get two string arr Intersection
func GetIntersection(strs1, strs2 []string) []string {
	var intersection = make([]string, 0)
	set1 := map[string]struct{}{}
	for _, v := range strs1 {
		set1[v] = struct{}{}
	}
	set2 := map[string]struct{}{}
	for _, v := range strs2 {
		set2[v] = struct{}{}
	}
	if len(set1) > len(set2) {
		set1, set2 = set2, set1
	}
	for v := range set1 {
		if _, has := set2[v]; has {
			intersection = append(intersection, v)
		}
	}
	return intersection
}

// Encrypt 加密
func Encrypt(str string) string {
	// 创建加密算法 aes
	c, err := aes.NewCipher([]byte(KeyText))
	if err != nil {
		fmt.Printf("Error: NewCipher(%d bytes) = %s", len(KeyText), err)
		return str
	}

	s := []byte(str)
	// 加密字符串
	cfb := cipher.NewCFBEncrypter(c, CommonIV)
	ciphertext := make([]byte, len(s))
	cfb.XORKeyStream(ciphertext, s)
	return string(ciphertext)
}

// Decrypt 解密
func Decrypt(str string) string {
	// 创建加密算法 aes
	c, err := aes.NewCipher([]byte(KeyText))
	if err != nil {
		fmt.Printf("Error: NewCipher(%d bytes) = %s", len(KeyText), err)
		return str
	}

	s := []byte(str)
	// 解密字符串
	cfbdec := cipher.NewCFBDecrypter(c, CommonIV)
	plaintextCopy := make([]byte, len(s))
	cfbdec.XORKeyStream(plaintextCopy, s)
	return string(plaintextCopy)
}

func IsGNS(str string) bool {
	match, err := regexp.MatchString(GNSPattern, str)
	if err != nil {
		return false
	}
	return match
}

func IsAdminRole(curRole []string) bool {
	var roles = map[string]string{"super_admin": "", "org_manager": "", "org_audit": "", "sec_admin": ""}
	var isAdmin = false
	for _, v := range curRole {
		if _, ok := roles[v]; ok {
			isAdmin = true
			break
		}
	}

	return isAdmin
}

func TimeCost() func() {
	startTime := time.Now()
	funcName := getFunctionName(1)

	return func() {
		elapsedTime := time.Since(startTime)
		log.Infof("Function %s took %s\n", &funcName, elapsedTime)
	}
}

func getFunctionName(skip int) string {
	pc, _, _, _ := runtime.Caller(skip)
	fullName := runtime.FuncForPC(pc).Name()
	lastSlash := 0

	for i := 0; i < len(fullName); i++ {
		if fullName[i] == '/' {
			lastSlash = i
		}
	}

	return fullName[lastSlash+1:]
}

// ParseInterface 解析data
func ParseInterface(in, out interface{}) {
	inByte, err := json.Marshal(in)
	if err != nil {
		return
	}
	_ = json.Unmarshal(inByte, out)
}

func GetFileExtension(filename string) string {
	extension := filepath.Ext(filename)
	return strings.ToLower(extension)
}

func ConvertTimeStringToMsTimestamp(timeStr string) (int64, error) {
	// 解析时间字符串
	parsedTime, err := time.Parse(time.RFC3339, timeStr)
	if err != nil {
		floatVal, err := strconv.ParseFloat(timeStr, 64)
		if err != nil {
			return 0, err
		}

		intVal := new(big.Float)
		intVal.SetFloat64(floatVal)
		int64Val, _ := intVal.Int64()
		return int64Val, err
	}

	// 转换为微秒级时间戳
	microsecondTimestamp := parsedTime.UnixNano() / 1000
	return microsecondTimestamp, nil
}

func BsonToInterface(b interface{}) interface{} {
	switch val := b.(type) {
	case map[string]interface{}:
		// 如果是map类型，递归调用
		for k, v := range val {
			val[k] = BsonToInterface(v)
		}
	case []interface{}:
		// 如果是切片类型，递归调用
		for i, item := range val {
			val[i] = BsonToInterface(item)
		}
	case bson.D:
		// 如果是bson.D，转换为map[string]interface{}
		result := make(map[string]interface{})
		for _, elem := range val {
			result[elem.Key] = BsonToInterface(elem.Value)
		}
		return result
	case bson.M:
		// 如果是bson.M，递归调用
		for k, v := range val {
			val[k] = BsonToInterface(v)
		}

	case bson.A:
		// 如果是bson.A，将其转换为[]interface{}
		result := make([]interface{}, len(val))
		for i, item := range val {
			result[i] = BsonToInterface(item)
		}
		return result
	}

	return b
}

// StructFormData 二进制流文件转换为form-data
func StructFormData(fieldname, filename string, con *[]byte) (*bytes.Buffer, *multipart.Writer, error) {
	var body = &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// 添加二进制数据字段
	part, err := writer.CreateFormFile(fieldname, filename)
	if err != nil {
		return body, nil, err
	}
	_, err = part.Write(*con)
	if err != nil {
		return body, nil, err
	}

	// 必须在写入完毕后关闭 writer
	writer.Close()
	return body, writer, nil
}

// TimeToTimestamp 时间字符串转13位时间戳
func TimeToTimestamp(timeStr string) (int64, error) {
	layout := "2006-01-02 15:04:05"
	t, err := time.Parse(layout, timeStr)
	if err != nil {
		return 0, err
	}

	timestamp := t.UnixNano() / 10e5
	return timestamp, nil
}

// IsByteLoss 判断字节是否丢失
func IsByteLoss(lenBuff int, startByte, endByte, spacingByte, fileSize int64) bool {
	if endByte == startByte {
		if lenBuff != 1 {
			return true
		}
	} else {
		if fileSize <= spacingByte {
			if lenBuff != int(fileSize) {
				return true
			}
		} else {
			if startByte == 0 {
				if lenBuff != int(endByte-startByte)+1 {
					return true
				}
			} else if endByte != fileSize {
				if lenBuff != int(endByte-startByte)+1 {
					return true
				}
			} else {
				//末字节为空
				if lenBuff != int(endByte-startByte) {
					return true
				}
			}
		}
	}
	return false
}

// ReadFile 读取文件到内存
func ReadFile(path string) ([]byte, error) {
	var buf []byte
	f, err := os.Open(path)
	if err != nil {
		return buf, err
	}
	defer f.Close()

	buf, err = io.ReadAll(f)
	if err != nil {
		return buf, err
	}

	return buf, nil
}

// CreateFolder 在当前工作目录创建文件夹
func CreateFolder(folder string) (string, error) {
	currentDir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	if !strings.HasPrefix(folder, "/") {
		folder = "/" + folder
	}
	filePath := filepath.Join(currentDir, folder)
	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		return filePath, nil
	}
	// 文件夹不存在，创建文件夹
	return filePath, os.MkdirAll(filePath, 0755)
}

// ComputeLevelDifference 获取两个路径的层级差 -1 表示两个路径非父子关系
func ComputeLevelDifference(parentPath, childPath string) int {
	parentPath = strings.Trim(parentPath, "/")
	childPath = strings.Trim(childPath, "/")

	if !strings.HasPrefix(childPath, parentPath) {
		return -1
	}

	relativePath := strings.TrimPrefix(childPath, parentPath)

	return len(strings.Split(strings.Trim(relativePath, "/"), "/"))
}

// SpliceDepPath 拼接部门全路径
func SpliceDepPath(parentDeps []interface{}, spliceType string) string {
	var depts []string
	for _, parentDep := range parentDeps {
		_parentDeps, ok := parentDep.([]interface{})
		if !ok {
			continue
		}
		for _, parentDep := range _parentDeps {
			dept_map, ok := parentDep.(map[string]interface{})
			if !ok {
				continue
			}
			if spliceType == "name" {
				depts = append(depts, dept_map["name"].(string))
			} else {
				depts = append(depts, dept_map["id"].(string))
			}
		}
	}
	return strings.Join(depts, ",")
}

func Contains[T comparable](slice []T, item T) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}

func ContainsInterface(slice []interface{}, item interface{}) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}

func StructToMap(s interface{}) (map[string]interface{}, error) {
	if v, ok := s.(map[string]interface{}); ok {
		return v, nil
	}
	data, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	err = json.Unmarshal(data, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// ConvertToRFC3339 将 2025-03-27T02:46:36.208Z 此种时间戳转换为RFC3339 格式时间戳秒级
func ConvertToRFC3339(timeStr string) (string, error) {
	// 尝试用 RFC3339Nano 解析（兼容带纳秒的时间）
	t, err := time.Parse(time.RFC3339Nano, timeStr)
	if err != nil {
		// 如果失败，尝试用 RFC3339 解析（不带纳秒）
		t, err = time.Parse(time.RFC3339, timeStr)
		if err != nil {
			return "", fmt.Errorf("invalid time format, expected RFC3339 or RFC3339Nano: %v", err)
		}
	}

	return t.Format(time.RFC3339), nil
}

// TimestampToRFC3339 数字类型时间戳转 RFC3339
// 参数：
// timestamp: 时间戳(支持秒/毫秒/微秒/纳秒)
// precision: 精度("s","ms","us","ns")，默认秒级
// withNano: 是否包含纳秒部分
// timezone: 时区(如"UTC","Local","Asia/Shanghai")，默认UTC
func TimestampToRFC3339(timestamp int64, precision string, withNano bool, timezone string) (string, error) {
	// 确定时间戳的纳秒数
	var nanos int64
	switch precision {
	case "s": // 秒级
		nanos = timestamp * 1e9
	case "ms": // 毫秒级
		nanos = timestamp * 1e6
	case "us": // 微秒级
		nanos = timestamp * 1e3
	case "ns": // 纳秒级
		nanos = timestamp
	default: // 默认秒级
		nanos = timestamp * 1e9
	}

	// 加载时区
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return "", fmt.Errorf("invalid timezone: %v", err)
	}

	// 创建时间对象
	t := time.Unix(0, nanos).In(loc)

	// 格式化为RFC3339
	if withNano {
		return t.Format(time.RFC3339Nano), nil
	}
	return t.Format(time.RFC3339), nil
}

// ComputeHash 计算文本块的 SHA-256 哈希值
func ComputeHash(content string) string {
	h := md5.New()
	h.Write([]byte(content))
	return hex.EncodeToString(h.Sum(nil))
}

// DeepCopy 深拷贝对象
func DeepCopy[T any](src T) T {
	b, _ := json.Marshal(src)
	var dst T
	err := json.Unmarshal(b, &dst)
	if err != nil {
		return src
	}
	return dst
}

// 微秒級
func TimeParse(timeStr string) int64 {
	var targetTime int64

	if timeStr == "1970-01-01T08:00:00+08:00" {
		return -1 // -1 表示无限期
	}
	timeParse, err := time.Parse(time.RFC3339, timeStr)
	if err != nil {
		return 0
	}

	targetTime = timeParse.UnixMicro()
	return targetTime
}

// tokenParser方法用于处理token字符串，如果tokenStr以Bearer开头，则返回tokenStr，否则返回Bearer + tokenStr
func TokenParser(tokenStr string) string {
	if strings.HasPrefix(tokenStr, "Bearer ") {
		return tokenStr
	}
	return "Bearer " + tokenStr
}

// IfNot 类似三元表达式 cond ? a : b
func IfNot[T any](cond bool, a, b T) T {
	if cond {
		return a
	}
	return b
}
