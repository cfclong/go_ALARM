package tools

import (
"fmt"
"gopkg.in/gomail.v2"
"log"
"strconv"
)

func SendMail(mailTo []string, subject string, body string) error {
	//定义邮箱服务器连接信息，如果是网易邮箱 pass填密码，qq邮箱填授权码
	mailConn := map[string]string{
		"user": "fuchang.chen@westwell-lab.com",
		"pass": "PmQWzThPKzjbgnQe",
		"host": "smtp.exmail.qq.com",
		"port": "465",
		//"user": "donotreply@cspadt.ae",
		//"pass": "cosco%456",
		//"host": "172.50.3.4",
		//"port": "25",
	}

	port, _ := strconv.Atoi(mailConn["port"]) //转换端口类型为int

	m := gomail.NewMessage()

	m.SetHeader("From",  m.FormatAddress(mailConn["user"], "Abu Dhabi")) //这种方式可以添加别名，即“XX官方”
	//m.SetHeader("From", "FB Sample"+"<"+mailConn["user"]+">") //这种方式可以添加别名，即“FB Sample”， 也可以直接用<code>m.SetHeader("From",mailConn["user"])</code>
	//m.SetHeader("From", mailConn["user"])
	m.SetHeader("To", mailTo...)    //发送给多个用户
	m.SetHeader("Subject", subject) //设置邮件主题
	m.SetBody("text/html", body)    //设置邮件正文
	//m.Attach("C:/Users/15737/go/src/awesomeProject4/conf/ips.conf")
	//m.Attach("C:/Users/15737/go/src/awesomeProject4/conf/wellocean.conf")
	d := gomail.NewDialer(mailConn["host"], port, mailConn["user"], mailConn["pass"])
	err := d.DialAndSend(m)
	return err

}
func main() {
	//定义收件人
	mailTo := []string{
		"donotreply@cspadt.ae",
		//"fuchang.chen@westwell-lab.com",
		//"yinghao.li@westwell-lab.com",
		//"yumeng.hu@westwell-lab.com",
		//"geiliang.yang@westwell-lab.com",
		//"wenfu.yu@westwell-lab.com",
		//"mingyu.li@westwell-lab.com",
	}
	//邮件主题为"Hello"
	subject := "Abu Dhabi project status alert"
	// 邮件正文
	body := "Hello,by gomail sent"

	err := SendMail(mailTo, subject, body)
	if err != nil {
		log.Println(err)
		fmt.Println("send fail")
		return
	}
	fmt.Println("send successfully")
}