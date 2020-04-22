package tools

import (
	"awesomeProject4/my_Smtp/smtp" //修改标准库smtp包
	"bufio"
	"bytes"
	"encoding/json"
	"github.com/garyburd/redigo/redis"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

//将配置文件json内容转化为JsonConf结构体C:\Users\15737\go\pkg\mod\github.com\bitly\go-simplejson@v0.5.0C:\Users\15737\go\pkg\mod\github.com\garyburd\redigo@v1.6.0
type JsonConf struct {
	AlarmStd map[string]float64 `json:"high"`
	MailTo   []string           `json:"mailTo"`
}

//工厂函数，传入json配置文件路径，生成上面的jsonConf实例
func NewJsonConf(fileName string) JsonConf {
	tmpJsonConf := JsonConf{}
	bytes2, _ := ioutil.ReadFile(fileName)
	err := json.Unmarshal(bytes2, &tmpJsonConf)
	if err != nil {
		log.Fatalln(err)
	}
	return tmpJsonConf
}

//读取文件,把每行内容生成为string切片
func ReadLineFile(fileName string) []string {
	var s []string
	if file, err := os.Open(fileName); err != nil {
		panic(err)
	} else {
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			s = append(s, scanner.Text())
		}
	}
	return s
}

//带单位的字符串转数字用于判断
func StrToNum(s string) float64 {
	var tmp string
	for k, v := range s {
		if v == 37 || v == 75 || v == 71 || v == 77 || v == 111 {
			tmp = s[0:k]
		} else {
			tmp = s
		}
	}
	f, _ := strconv.ParseFloat(tmp, 64)
	return f
}

//字符串切片转为字符串
func StrSliceToStr(s []string) string {
	var str string
	for _, v := range s {
		str = str + "[" + v + "]"
	}
	return str
}

func StoS(s map[string]string) string {
	if len(s) == 0 {
		return "emptyData"
	}
	tplStr := `
		<html>
		<body>
		<h3>
		{{range $k, $v := .}}
		<h3>{{$v}}</h3>
  		{{end}}
		</h3>
		</body>
		</html>
		`
	outBuf := &bytes.Buffer{}
	tpl := template.New("email notify template")
	tpl, _ = tpl.Parse(tplStr)
	_ = tpl.Execute(outBuf, s)
	//fmt.Println("outBuf is: ",outBuf)
	return outBuf.String()
}

//传入redis键值，判断是否存在，不存则在生成该键值（存活时间60分钟）
func ExistKey(c redis.Conn, s string) (bool, error) {
	b, err := redis.Int(c.Do("exists", s))
	if err != nil {
		return false, err
	} else if b == 1 {
		return true, nil
	} else {
		_, _ = c.Do("set", s, 1)
		_, _ = c.Do("EXPIRE", s, "3600")
		return false, nil
	}
}

//传入字符串，判断十分钟内是否连续n次出现
func ExistN(c redis.Conn, s string) (bool, error) {
	//b, err := redis.Int(c.Do("exists", s))
	b, err := ExistKey(c,s)
	if err != nil {
		return false, err
	}
	if b  {
		n, _ := redis.Int(c.Do("get", s))
		if n < 5 {
			n++
			_, _ = c.Do("set", s, n)
			return false, nil
		}
		return true, nil
	} else { //如果不存在就创建，并设置存活时间十分钟
		_, _ = c.Do("set", s, 1)
		_, _ = c.Do("EXPIRE", s, "600")
		return false, nil
	}
}

func Get(url string) string {
	// 超时时间：5秒
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	var buffer [512]byte
	result := bytes.NewBuffer(nil)
	for {
		n, err := resp.Body.Read(buffer[0:])
		result.Write(buffer[0:n])
		if err != nil && err == io.EOF {
			break
		} else if err != nil {
			panic(err)
		}
	}

	return result.String()
}

//邮件发送函数
func SendToMail(body string, sendTo []string) error {
	user := "donotreply@cspadt.ae"
	password := "cosco%456"
	host := "172.50.3.4:25"
	//user := "fuchang.chen@westwell-lab.com"
	//password := "PmQWzThPKzjbgnQe"
	//host := "smtp.exmail.qq.com:465"
	subject := "Abu_Dhabi_OCR_Alarm"
	mailType := "html"
	hp := strings.Split(host, ":")
	auth := smtp.PlainAuth("", user, password, hp[0])
	var contentType string
	contentType = "Content-Type: text/" + mailType + "; charset=UTF-8"
	//if mailType == "html" {
	//	contentType = "Content-Type: text/" + mailType + "; charset=UTF-8"
	//} else {
	//	contentType = "Content-Type: text/plain" + "; charset=UTF-8"
	//}
	var to string
	for _, v := range sendTo {
		to = to + v + ";"
	}
	msg := []byte("To: " + to + "\r\nFrom: " + user + ">\r\nSubject: " + subject + "\r\n" + contentType + "\r\n\r\n" + body)
	//sendTo := strings.Split(to, ";")
	err := smtp.SendMail(host, auth, user, sendTo, msg)
	return err
}
//C:\Users\15737\go\pkg\mod\github.com\bitly\go-simplejson@v0.5.0
//C:\Users\15737\go\pkg\mod\gopkg.in\gomail.v2@v2.0.0-20160411212932-81ebce5c23df
//C:\Users\15737\go\pkg\mod\gopkg.in\alexcesaro\quotedprintable.v3@v3.0.0-20150716171945-2caba252f4dc
//C:\Users\15737\go\pkg\mod\github.com\garyburd\redigo@v1.6.0