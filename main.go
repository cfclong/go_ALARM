package main

import (
	"awesomeProject4/tools"
	"fmt"
	"github.com/bitly/go-simplejson"
	"github.com/garyburd/redigo/redis"
	"log"
	"os"
	"strconv"
	"time"
)

const N = 1 //权重
//定义的mess结构体用于保存各项目数据
type mess struct {
	ips           []string
	wellocean     map[string][]string
	cpu           map[string]map[string]float64
	disk          map[string]map[string]float64
	wellInterface map[string][]string
	mem           map[string]map[string]float64
	gpu           map[string]map[string]float64
	sysproc       map[string][]string
	times         map[string]string
	nowTimestamp  int64
	camera        []string
	logMes        map[string][]string
}

func (dj *delJsons) messTomess(c redis.Conn) mess {
	m := mess{
		ips:           dj.Ips, //ip列表
		wellocean:     dj.delWellocean(),
		cpu:           dj.delCpu(),
		disk:          dj.delDiskUsage(),
		wellInterface: dj.delInterface(),
		mem:           dj.delMem(),
		gpu:           dj.delGpu(),
		sysproc:       dj.delSysproc(),
		times:         dj.get_time(),     //获取服务告警信息生成的时间
		nowTimestamp:  time.Now().Unix(), //当前时间戳
		camera:        dj.Camers(),
		logMes:        dj.delLog(),
	}
	//fmt.Println(m)
	messTmp := mess{
		ips:           m.ips,
		wellocean:     make(map[string][]string),
		cpu:           make(map[string]map[string]float64),
		disk:          make(map[string]map[string]float64),
		wellInterface: make(map[string][]string),
		mem:           make(map[string]map[string]float64),
		gpu:           make(map[string]map[string]float64),
		sysproc:       make(map[string][]string),
		times:         m.times,
		nowTimestamp:  m.nowTimestamp,
		camera:        nil,
		logMes:        make(map[string][]string),
	}
	//fmt.Println(messTmp)
	if len(m.wellocean) != 0 {
		for ip, serviceNameSlice := range m.wellocean {
			if b, err := tools.ExistKey(c, ip+"welloceanServiceErr"); !b {
				if err != nil {
					panic(err)
				} else {
					messTmp.wellocean[ip] = serviceNameSlice
				}
			}
		}
	}
	if len(m.wellInterface) != 0 {
		for ip, serviceNameSlice := range m.wellInterface {
			if b, err := tools.ExistKey(c, ip+"wellInterfaceServiceErr"); !b {
				if err != nil {
					panic(err)
				} else {
					messTmp.wellInterface[ip] = serviceNameSlice
				}
			}
		}
	}
	if len(m.sysproc) != 0 {
		for ip, serviceNameSlice := range m.sysproc {
			if b, err := tools.ExistKey(c, ip+"sysprocServiceErr"); !b {
				if err != nil {
					panic(err)
				} else {
					messTmp.sysproc[ip] = serviceNameSlice
				}
			}
		}
	}
	if len(m.cpu) != 0 {
		for ip, cpuMes := range m.cpu {
			if b, _ := tools.ExistN(c, ip+"cpuMesErr"); b {
				if b1, _ := tools.ExistKey(c, ip+"cpuMesErrKeyTime"); !b1 {
					messTmp.cpu[ip] = cpuMes
				}
			}
		}
	}
	if len(m.disk) != 0 {
		for ip, diskMes := range m.disk {
			if b, _ := tools.ExistN(c, ip+"diskMesErr"); b {
				if b1, _ := tools.ExistKey(c, ip+"diskMesErrKeyTime"); !b1 {
					messTmp.disk[ip] = diskMes
				}
			}
		}
	}
	if len(m.mem) != 0 {
		for ip, memMes := range m.mem {
			if b, _ := tools.ExistN(c, ip+"memMesErr"); b {
				if b1, _ := tools.ExistKey(c, ip+"memMesErrKeyTime"); !b1 {
					messTmp.mem[ip] = memMes
				}
			}
		}
	}
	if len(m.gpu) != 0 {
		for ip, gpuMes := range m.gpu {
			if b, _ := tools.ExistN(c, ip+"gpuMesErr"); b {
				if b1, _ := tools.ExistKey(c, ip+"gpuMesErrKeyTime"); !b1 {
					messTmp.gpu[ip] = gpuMes
				}
			}
		}
	}
	if len(m.camera) != 0 {
		if b, err := tools.ExistKey(c, "camersErr"); !b {
			if err != nil {
				panic(err)
			} else {
				messTmp.camera = m.camera
			}
		}
	}
	if len(m.logMes) != 0 {
		for ip, LogErrMes := range m.logMes {
			if b, err := tools.ExistKey(c, ip+"LogErr"); !b {
				if err != nil {
					panic(err)
				} else {
					messTmp.logMes[ip] = LogErrMes
				}
			}
		}
	}
	return messTmp
}

func (m *mess) getMessage() map[string]string {
	strMap := make(map[string]string)
	for _, ip := range m.ips {
		if len(m.wellocean[ip]) != 0 {
			strMap[ip] = " The Server " + ip + " at " + m.times[ip] + " detected:"
			strMap[ip] = strMap[ip] + " [wellocean program not running] "
		}
		if len(m.cpu[ip]) != 0 {
			if len(strMap[ip]) == 0 {
				strMap[ip] = "The Server " + ip + " at " + m.times[ip] + " detected:"
			}
			strMap[ip] = strMap[ip] + " [High CPU load] "
		}
		if len(m.disk[ip]) != 0 {
			if len(strMap[ip]) == 0 {
				strMap[ip] = "The Server " + ip + " at " + m.times[ip] + " detected:"
			}
			var s string
			for k, _ := range m.disk[ip] {
				s = s + k +" "
			}
			strMap[ip] = strMap[ip] + " [high disk lode: " + s + "]"
		}
		if len(m.wellInterface[ip]) != 0 {
			if len(strMap[ip]) == 0 {
				strMap[ip] = "The Server " + ip + " at " + m.times[ip] + " detected:"
			}
			s := tools.StrSliceToStr(m.wellInterface[ip])
			strMap[ip] = strMap[ip] + " [interface program not running:" + s + "] "
		}
		if len(m.mem[ip]) != 0 {
			if len(strMap[ip]) == 0 {
				strMap[ip] = "The Server " + ip + " at " + m.times[ip] + " detected:"
			}
			strMap[ip] = strMap[ip] + " [High memory load] "
		}
		if len(m.gpu[ip]) != 0 {
			if len(strMap[ip]) == 0 {
				strMap[ip] = " The Server " + ip + " at " + m.times[ip] + " detected:"
			}
			strMap[ip] = strMap[ip] + " [High GPU load] "
		}
		if len(m.sysproc[ip]) != 0 {
			if len(strMap[ip]) == 0 {
				strMap[ip] = " The Server " + ip + " at " + m.times[ip] + " detected:"
			}
			s := tools.StrSliceToStr(m.sysproc[ip])
			strMap[ip] = strMap[ip] + " [sysproc program not running:" + s + "] "
		}
		if len(m.logMes[ip]) != 0 {
			if len(strMap[ip]) == 0 {
				strMap[ip] = "The Server " + ip + " at " + m.times[ip] + " detected:"
			}
			s := tools.StrSliceToStr(m.logMes[ip])
			strMap[ip] = strMap[ip] + " [LogErr:" + s + "] "
		}
	}
	if len(m.camera) != 0 {
		strMap["camers_info"] = "Camera lost connection: " + tools.StrSliceToStr(m.camera)
	}
	return strMap
}

type delJsons struct {
	Ips      []string
	j        []byte //json data
	AlarmStd map[string]float64
}

//wellocean
func (dj *delJsons) delWellocean() map[string][]string {
	strMap := make(map[string][]string)
	res, _ := simplejson.NewJson(dj.j)
	for _, ip_v := range dj.Ips {
		welloceans, _ := res.Get(ip_v).Get("monitor").Get("wellocean").Map()
		for k, v := range welloceans {
			if v == false {
				strMap[ip_v] = append(strMap[ip_v], k)
			}
		}
	}
	return strMap
}

//DiskUsage
func (dj *delJsons) delDiskUsage() map[string]map[string]float64 {
	rootUsageStd, _ := dj.AlarmStd["disk_root_usage"] //从文件中获取cpu使用率标准
	cvUsageStd, _ := dj.AlarmStd["disk_cv_usage"]
	strMap := make(map[string]map[string]float64)
	res, _ := simplejson.NewJson(dj.j)
	for _, ip_v := range dj.Ips {
		result01, _ := res.Get(ip_v).Get("monitor").Get("disk").Get("disk_cv_usage").String()
		data01 := tools.StrToNum(result01) / N
		if data01 >= rootUsageStd {
			strMap[ip_v] = map[string]float64{"cvUsage": data01}
		}
		result02, _ := res.Get(ip_v).Get("monitor").Get("disk").Get("disk_root_usage").String()
		data02 := tools.StrToNum(result02) / N
		if data02 >= cvUsageStd {
			strMap[ip_v] = map[string]float64{"rootUsage": data02}
		}
	}
	return strMap
}

//CPU

func (dj *delJsons) delCpu() map[string]map[string]float64 {
	cpuUsageStd, _ := dj.AlarmStd["cpu_usage"] //从文件中获取cpu使用率标准
	cpuLoadStd, _ := dj.AlarmStd["cpu_load"]
	strMap := make(map[string]map[string]float64)
	res, _ := simplejson.NewJson(dj.j)
	for _, ip_v := range dj.Ips {
		result01, _ := res.Get(ip_v).Get("monitor").Get("cpu").Get("cpu_usage").String()
		data01 := tools.StrToNum(result01) / N
		if data01 >= cpuUsageStd {
			strMap[ip_v] = map[string]float64{"cpu_usage": data01}
		}
		result02, _ := res.Get(ip_v).Get("monitor").Get("cpu").Get("cpu_load").String()
		data02 := tools.StrToNum(result02) / N
		if data02 >= cpuLoadStd {
			strMap[ip_v] = map[string]float64{"cpu_load": data02}
		}
	}
	return strMap
}

//Interface
func (dj *delJsons) delInterface() map[string][]string {
	strMap := make(map[string][]string)
	res, _ := simplejson.NewJson(dj.j)
	for _, ip_v := range dj.Ips {
		interfaces, _ := res.Get(ip_v).Get("monitor").Get("interface").Map()
		for k, v := range interfaces {
			if v == false {
				strMap[ip_v] = append(strMap[ip_v], k)
			}
		}
	}
	return strMap
}

//Memory
func (dj *delJsons) delMem() map[string]map[string]float64 {
	memUsageStd := dj.AlarmStd["mem_usage"]
	strMap := make(map[string]map[string]float64)
	res, _ := simplejson.NewJson(dj.j)
	for _, ipV := range dj.Ips {
		result01, _ := res.Get(ipV).Get("monitor").Get("mem").Get("mem_usage").String()
		data01 := tools.StrToNum(result01) / N
		if data01 >= memUsageStd {
			strMap[ipV] = map[string]float64{"mem_usage": data01}
		}
	}
	return strMap
}

//GPU
func (dj *delJsons) delGpu() map[string]map[string]float64 {
	gpu_usage_1_Std := dj.AlarmStd["gpu_usage_1"]
	gpu_usage_2_Std := dj.AlarmStd["gpu_usage_2"]
	gpu_tempStd := dj.AlarmStd["gpu_temp"]
	strMap := make(map[string]map[string]float64)
	res, _ := simplejson.NewJson(dj.j)
	for _, ip_v := range dj.Ips {
		result01, _ := res.Get(ip_v).Get("monitor").Get("gpu").Get("gpu_usage_1").String()
		data01 := tools.StrToNum(result01) / N
		result02, _ := res.Get(ip_v).Get("monitor").Get("gpu").Get("gpu_usage_2").String()
		data02 := tools.StrToNum(result02) / N
		result03, _ := res.Get(ip_v).Get("monitor").Get("gpu").Get("gpu_temp").String()
		data03 := tools.StrToNum(result03) / N
		if data01 >= float64(gpu_usage_1_Std) {
			strMap[ip_v] = map[string]float64{"gpu_usage_1": data01}
		}
		if data02 >= float64(gpu_usage_2_Std) {
			strMap[ip_v] = map[string]float64{"gpu_usage_2": data02}
		}
		if data03 >= float64(gpu_tempStd) {
			strMap[ip_v] = map[string]float64{"gpu_temp": data03}
		}
	}
	return strMap
}

//Sysproc
func (dj *delJsons) delSysproc() map[string][]string {
	strMap := make(map[string][]string)
	res, _ := simplejson.NewJson(dj.j)
	for _, ip_v := range dj.Ips {
		welloceans, _ := res.Get(ip_v).Get("monitor").Get("sysproc").Map()
		for k, v := range welloceans {
			if v == false {
				strMap[ip_v] = append(strMap[ip_v], k)
			}
		}
	}
	return strMap
}

//log
func (dj *delJsons) delLog() map[string][]string {
	strMap := make(map[string][]string)
	logType := []string{"CraneDataSender", "GetPLC_siemens", "OSK"}
	res, _ := simplejson.NewJson(dj.j)
	for _, ip_v := range dj.Ips {
		for _, v := range logType {
			logErrMes, err := res.Get(ip_v).Get("monitor").Get("log").Get(v).Get("log_status").Int()
			if err != nil {
				fmt.Println("LOG DEBUG")
				//continue
			} else {
				if logErrMes != 0 {
					strMap[ip_v] = append(strMap[ip_v], v+" LogErr "+strconv.Itoa(logErrMes))
				}
			}
		}
	}
	return strMap
}

func (dj *delJsons) get_time() map[string]string {
	strMap := make(map[string]string)
	res, _ := simplejson.NewJson(dj.j)
	for _, ip_v := range dj.Ips {
		timestamp, _ := res.Get(ip_v).Get("timestamp").Int64()
		//str := time.Unix(timestamp, 0).Format("2006-01-02 15:04:05")
		strMap[ip_v] = time.Unix(timestamp, 0).Format("2006-01-02 15:04:05")
	}
	return strMap
}

func (dj *delJsons) Camers() []string {
	var cam []string
	res, _ := simplejson.NewJson(dj.j)
	c, _ := res.Get("camers_info").Map()
	for k, v := range c {
		if v == false {
			cam = append(cam, k)
		}
	}
	return cam
}

func (dj *delJsons) GetMes(c redis.Conn) map[string]string {
	Messages := dj.messTomess(c)
	tmp := Messages.getMessage()
	return tmp
}


func main() {
	confPath, _ := os.Getwd() //获取当前路径
	jsonStr := tools.Get("http://172.90.109.119:22122/wellx/monitor-api")
	wellConf := tools.NewJsonConf(confPath + "/AlarmStandard.json")
	mailToConf := wellConf.MailTo
	std := wellConf.AlarmStd
	IPs := tools.ReadLineFile(confPath + "/ips.conf")
	c, err := redis.Dial("tcp", "172.90.109.119:6379")
	if err != nil {
		panic(err)
	}
	defer c.Close()
	_, err = c.Do("select", "11")
	if err != nil {
		panic(err)
	}
	bytes := []byte(jsonStr)
	var dJson = delJsons{
		Ips:      IPs,
		j:        bytes,
		AlarmStd: std,
	}
	Messages := dJson.GetMes(c)
	//fmt.Println(Messages)
	mailTo := mailToConf
	body := tools.StoS(Messages)
	if body != "emptyData" {
		err := tools.SendToMail(body, mailTo)
		if err != nil {
			log.Println(err)
			fmt.Println("send fail")
			return
		}
	}
}


/*
func main() {
	confPath, _ := os.Getwd() //获取当前路径
	jsonStr := tools.Get("http://172.90.109.119:22122/wellx/monitor-api")
	//wellConf := tools.NewJsonConf("C:/Users/15737/go/src/awesomeProject4/conf/AlarmStandard.json")
	wellConf := tools.NewJsonConf(confPath + "/AlarmStandard.json")
	mailToConf := wellConf.MailTo
	std := wellConf.AlarmStd
	//IPs := tools.ReadLineFile("C:/Users/15737/go/src/awesomeProject4/conf/ips.conf") //读取文件为[]string
	IPs := tools.ReadLineFile(confPath + "/ips.conf")
	//c, err := redis.Dial("tcp", "101.132.137.161:6379")
	//c, err := redis.Dial("tcp", "172.90.109.119:6379")
	//if err != nil {
	//	panic(err)
	//}
	//defer c.Close()
	//_, err = c.Do("select", "11")
	//if err != nil {
	//	panic(err)
	//}
	bytes := []byte(jsonStr)
	//bytes, _ := ioutil.ReadFile("C:/Users/15737/go/src/awesomeProject4/conf/monitor_cache.json")
	//bytes, _ := ioutil.ReadFile(confPath+"/monitor_cache.json")
	var dJson = delJsons{
		Ips:      IPs,
		j:        bytes,
		AlarmStd: std,
	}
	m := mess{
		ips:           dJson.Ips, //ip列表
		wellocean:     dJson.delWellocean(),
		cpu:           dJson.delCpu(),
		disk:           dJson.delDiskUsage(),
		wellInterface: dJson.delInterface(),
		mem:           dJson.delMem(),
		gpu:           dJson.delGpu(),
		sysproc:       dJson.delSysproc(),
		times:         dJson.get_time(),  //获取服务告警信息生成的时间
		nowTimestamp:  time.Now().Unix(), //当前时间戳
		camera:        dJson.Camers(),
		logMes:        dJson.delLog(),
	}
	fmt.Println(m)
	Messages := m.getMessage()
	fmt.Println(Messages)
	//for _,v :=range m.getMessage() {
	//	fmt.Println(v)
	//}
	//fmt.Println(m.getMessage())
	//Messages := dJson.GetMes(c)
	//for _, v := range Messages {
	//	fmt.Println(v)
	//}
	//定义收件人
	mailTo := mailToConf
	//mailTo = []string {
	//	"fuchang.chen@westwell-lab.com",
	//}
	//邮件主题
	//subject := "Abu Dhabi project status alert"
	// 邮件正文
	body := tools.StoS(Messages)
	if body != "emptyData" {
		err := tools.SendToMail(body, mailTo)
		if err != nil {
			log.Println(err)
			fmt.Println("send fail")
			return
		} else {
			fmt.Println("send success")
		}
	} else {
		fmt.Println("emptyData")
	}

}
*/