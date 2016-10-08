package cache

import (
	//"crypto/tls"
	"fmt"
	//	"io"
	"io/ioutil"
	"net"
	//	"strings"
	"os"
	"sync"
	//"time"
	"bytes"
	"log"
	"math/rand"
	"strconv"
	"time"
	//"golang.org/x/net/http2"
)

// Attacker is an attack executor which wraps an cache.Client
type Attacker struct {
	remote string      // 服务器的ip和端口号  格式 ip:port
	dialer *net.Dialer //本地
	//client  string //客户端命令
	//protocol string //客户端协议
	stopch    chan struct{}
	workers   uint64
	fileName  string //文件名
	blockSize uint64
	//offset int  //偏移量
	//length uint64 //偏移长度
}

//var number int

var mutex chan int = make(chan int, 1) //同步时间

var sumTimes time.Duration = 0 //sumTimes

var successwork uint64 = 0 //成功协程个数

const (
	//DefaultConnections = 10000
	// DefaultWorkers is the default initial number of workers used to carry an attack.
	DefaultWorkers = 10
)

var (
	DefaultLocalAddr = net.IPAddr{IP: net.IPv4zero}
)

func NewAttacker(opts ...func(*Attacker)) *Attacker {
	a := &Attacker{stopch: make(chan struct{}), workers: DefaultWorkers}

	a.dialer = &net.Dialer{
		LocalAddr: &net.TCPAddr{IP: DefaultLocalAddr.IP, Zone: DefaultLocalAddr.Zone},
	}
	for _, opt := range opts {
		opt(a)
	}
	return a
}

//将ip和端口号进行拼接
func Addr(ipCom, port string) func(*Attacker) {
	return func(a *Attacker) {
		var buffer bytes.Buffer
		buffer.WriteString(ipCom)
		buffer.WriteString(":")
		buffer.WriteString(port)
		a.remote = buffer.String()
	}
}

//func

func Workers(n uint64) func(*Attacker) {

	return func(a *Attacker) {
		a.workers = n

	}
}

func SetFileName(file string) func(*Attacker) {
	return func(a *Attacker) {
		a.fileName = file
	}
}

func SetBlockSize(length uint64) func(*Attacker) {
	return func(a *Attacker) {
		a.blockSize = length
	}
}
func (a *Attacker) GetSocket() string {

	return a.remote
}

func (a *Attacker) Attack(tr Target) {
	var workers sync.WaitGroup // 等待所有的goruntine执行完毕
	//results :=make(chan * Result) //处理结果
	//ticks := make(chan time.Time)
	//number = 0
	for i := uint64(0); i < a.workers; i++ {
		fmt.Printf("开始创建第%v个客户端", i)
		go a.attack(tr, &workers)
	}

	defer workers.Wait()

	//return results
}

var files []string = make([]string, 0) //创建容纳用来攻击的文件
func spliceAppend(file string) {
	files = append(files, file)
}
func (a *Attacker) AttackRead() func(string) {
	var workers sync.WaitGroup
	lastSize := int64(a.blockSize)
	errRead := ReadLine(a.fileName, spliceAppend)
	if errRead != nil {

		return func(string) {
			fmt.Println("打印信息:AttackRead发生错误了")
		}
	}
	mutex <- 100
	for _, value := range files {
		fmt.Printf("读取服务器上文件%v,并针对每个文件建立%v个客户端读取\n", value, a.workers)
		fileInfo, err := os.Stat(value)
		if err != nil {
			fmt.Println("ERROR!!!!,本地没有文件:", value)
			continue
		}
		//fmt.Println("上传文件:", value)
		size := fileInfo.Size()
		fileName := value + "_read_log_" + strconv.FormatUint(a.workers, 10)
		logFile, err := os.Create(fileName)

		if err != nil {
			fmt.Println("ERROR!!!日志文件创建失败(并不影响程序的运行)")
		}
		fileError := value + "_head_log_" + strconv.FormatUint(a.workers, 10)
		logError, err := os.Create(fileError)
		errorLog := log.New(logError, "[error log]", log.Llongfile)
		timeLog := log.New(logFile, "[Read time]", log.Llongfile)
		for i := uint64(0); i < a.workers; i++ {
			//srand := rand.Source(i)
			filesize := rand.Int63n(size - lastSize)
			workers.Add(1)
			//fmt.Printf("\n创建第%v个客服端\n", i+1)
			go a.attackRead2(value, filesize, &workers, timeLog, errorLog, i+1)
			//go a.attackRead(value, size, &workers, timeLog)
		}

	}
	//time.Sleep(1)
	defer workers.Wait()
	return func(fileName string) {
		averageTime := float64(sumTimes.Seconds()) / float64(successwork)
		fmt.Printf("统计信息：成功的协程数是:%v,总共花费了时间:%vs,平均每个协程花费的时间是:%vs\n", successwork, sumTimes.Seconds(), averageTime)
		workersString := strconv.FormatUint(a.workers, 10)
		successString := strconv.FormatUint(successwork, 10)
		averageString := strconv.FormatFloat(averageTime, 'f', 50, 64) // 参考http://www.cnblogs.com/golove/p/3262925.html
		err := WriteCsv(fileName, []string{workersString, successString, averageString})
		if err != nil {
			fmt.Printf("[ERROR]协程数:%v,统计信息没写成功,err=%v\n", a.workers, err)
		}
	}
}

func (a *Attacker) AttackWrite() {
	var workers sync.WaitGroup
	errRead := ReadLine(a.fileName, spliceAppend)
	if errRead != nil {
		fmt.Println("AttackWrite发生错误了")
		return
	}
	for ix, value := range files {
		fmt.Printf("创建第%v个客服端\n", ix+1)
		fileInfo, err := os.Stat(value)
		if err != nil {
			fmt.Println("ERROR!!!错误,本地没有文件", value)
			continue
		}
		fmt.Println("\n上传文件:", value)
		size := fileInfo.Size()
		fileName := value + "_write_log"

		logFile, err := os.Create(fileName)
		if err != nil {
			fmt.Println("ERROR!!!日志文件创建失败(并不影响程序的运行)")
		}
		timeLog := log.New(logFile, "[Write_time]", log.Llongfile)
		workers.Add(1)
		go a.attackWrite(value, size, &workers, timeLog)
		//time.Sleep(1)
	}
	//time.Sleep(1)
	defer workers.Wait()
}

func (a *Attacker) AttackOpen() {
	var workers sync.WaitGroup
	errRead := ReadLine(a.fileName, spliceAppend)
	if errRead != nil {
		fmt.Println("AttackOpen发生错误了")
		return
	}
	for ix, value := range files {
		//fmt.Println("value=", value)
		fmt.Printf("开始创建第%v个文件,并模拟出%v个客服端攻击", ix+1, a.workers)
		fileInfo, err := os.Stat(value)
		if err != nil {
			fmt.Println("ERROR!!!:错误,本地没有文件", value)
			continue
		}
		//fmt.Println("让服务器创建文件:", value)
		size := fileInfo.Size()
		for i := uint64(0); i < a.workers; i++ {
			workers.Add(1)
			fmt.Printf("\n创建第%v个客服端\n", i+1)
			go a.attackOpen(value, size, &workers)
		}
		//time.Sleep(1)

	}
	defer workers.Wait()

}

//每个go routine只是读取一部分就ok
func (a *Attacker) attackRead2(value string, filesSize int64, workers *sync.WaitGroup, timeLog *log.Logger, errorLog *log.Logger, index uint64) {
	defer workers.Done()
	tcpAddr, err := net.ResolveTCPAddr("tcp4", a.remote)
	if err != nil {
		return
	}
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		timeLog.Printf("[ERROR]第%v个客户端在读取文件%v:偏移量%v,长度为%v,拨号错误:%v\n", index, value, filesSize, a.blockSize, err)
		return
	}
	defer conn.Close()
	localAddress := conn.LocalAddr()
	//fmt.Println("network=", localAddress.Network())
	//fmt.Println("localAddr=", localAddress.String())
	var tr Target
	data := make([]byte, a.blockSize)
	tr, _ = NewReadTargeter("read", a.blockSize, uint64(filesSize), value)
	//localAddress := conn.LocalAddr()
	//fmt.Println("tr.command=", tr.Command)
	_, err = conn.Write([]byte(tr.Command)) //只是写缓冲
	if err != nil {
		timeLog.Printf("[ERROR]第%v个客户端:%v在读取文件%v:偏移量%v,长度为%v,发送命令出错\n", index, localAddress.String(), value, filesSize, a.blockSize)
		return
	}
	start := time.Now()
	count, err := ParseReadProtocol(data, uint64(filesSize), a.blockSize, conn, errorLog)
	end := time.Now()
	if err != nil {
		timeLog.Printf("[ERROR]第%v个客户端:%v在读取文件%v:偏移量%v,长度为%v,接收出错了:%v\n", index, localAddress.String(), value, filesSize, a.blockSize, err)
	} else if count != a.blockSize {
		timeLog.Printf("[ERROR]第%v个客户端:%v在读取文件%v:偏移量%v,长度为%v,接收出错了:接收长度出现不一致:%v\n", index, localAddress.String(), value, filesSize, a.blockSize, count)
	} else {
		delta := end.Sub(start)
		timeLog.Printf("[SUCCESS]第%v个客户端:%v在读取文件%v:偏移量%v,长度为%v,一共花费的时间是:%v\n", index, localAddress.String(), value, filesSize, a.blockSize, delta)
		fmt.Printf("[SUCCESS]第%v个客户端完成读取文件%v,偏移量是%v,长度为%v\n", index, value, filesSize, a.blockSize)
		<-mutex
		sumTimes += delta
		successwork += 1
		mutex <- 100
	}

}

//读取整个文件，无法达到测试的效果
func (a *Attacker) attackRead(value string, fileSize int64, workers *sync.WaitGroup, timeLog *log.Logger) {
	//workers.Add(1)
	defer workers.Done()
	var offset uint64 = 0

	tcpAddr, err1 := net.ResolveTCPAddr("tcp4", a.remote)
	if err1 != nil {
		return
	}
	conn, err2 := net.DialTCP("tcp", nil, tcpAddr)
	if err2 != nil {
		return
	}
	var tr Target
	defer conn.Close()
	var requestCount uint64
	start := time.Now() //计算时间
	for {

		if uint64(fileSize)-offset > a.blockSize {
			requestCount = a.blockSize
		} else {
			requestCount = uint64(fileSize) - offset
		}
		data := make([]byte, requestCount)
		tr, _ = NewReadTargeter("read", requestCount, offset, value)
		_, err3 := conn.Write([]byte(tr.Command))
		if err3 != nil {
			fmt.Println("发送读取命令出错")
			return
		}
		//_, err4 := conn.Read(data)
		count, err := ParseReadProtocol(data, offset, requestCount, conn, nil)
		if err != nil {
			fmt.Printf("ERROR!!!在读取文件%v,%v\n", value, err)
			return
		}
		if count != requestCount {
			fmt.Printf("在读取文件%v,接收的数据长度%v和请求的长度不一致%v", value, count, requestCount)
		}
		//fmt.Println("接收客服端的数据是:", data)

		offset = offset + requestCount
		if offset >= uint64(fileSize) {
			end := time.Now()
			delta := end.Sub(start)
			//log.
			timeLog.Printf("完整读取文件:%v花费了Time:%v\n", value, delta)
			fmt.Println("\nsuccess:成功请求文件:", value)
			break
		}
		rate := float64(offset) / float64(fileSize) * 100
		fmt.Printf("文件%v,读取了:%6.4f%\n", value, rate)
	}

}
func (a *Attacker) attackWrite(value string, fileSize int64, workers *sync.WaitGroup, timeLog *log.Logger) {
	//workers.Add(1)
	defer workers.Done()
	var offset uint64 = 0
	/*
		var conn net.TCPConn
		err := a.buildTCP(&conn)
		if err != nil {
			fmt.Println("创建TCP连接发生错误")
			return
		}*/

	tcpAddr, err1 := net.ResolveTCPAddr("tcp4", a.remote)
	if err1 != nil {
		return
	}
	conn, err2 := net.DialTCP("tcp", nil, tcpAddr)
	if err2 != nil {
		return
	}

	var tr Target
	//var errtr error
	defer conn.Close()
	start := time.Now()
	//data := make([]byte, 1024)
	for {

		tr, _ = NewWriteTargeterNoData("write", a.blockSize, offset, value)
		_, err3 := conn.Write([]byte(tr.Command))
		//fmt.Println("客服端发送的命令是:", tr.Command)
		if err3 != nil {
			fmt.Println("发送上传命令失败")
			return
		}
		//fmt.Println("等待解析")
		err := ParseWriteProtocol(conn)
		//fmt.Println("解析完毕")
		if err != nil {
			fmt.Printf("ERROR!!!,在上传文件%v,%v\n", value, err)
			return
		}
		//fmt.Println("收到服务器的数据:", string(data))

		offset = offset + tr.Length
		fmt.Printf("offset=%v,fileSize=%v\n", offset, fileSize)
		if offset >= uint64(fileSize) {
			end := time.Now()
			delta := end.Sub(start)
			timeLog.Printf("完整上传文件:%v花费了时间:%v", value, delta)
			fmt.Println("\nsuccess:成功上传文件:", value)
			break
		}
		rate := float64(offset) / float64(fileSize) * 100
		fmt.Printf("文件%v,上传了:%6.4f%\n", value, rate)
	}

}
func (a *Attacker) attackOpen(value string, fileSize int64, workers *sync.WaitGroup) {
	//tcpAddr,err1:=net.
	//fmt.Println("test problem1")
	//workers.Add(1)
	defer workers.Done()
	/*
		var conn net.TCPConn
		err := a.buildTCP(&conn)
		if err != nil {
			fmt.Println("创建TCP连接发生错误")
			return
		}*/

	//fmt.Println("test problem")
	tcpAddr, err1 := net.ResolveTCPAddr("tcp4", a.remote)
	if err1 != nil {
		return
	}
	conn, err2 := net.DialTCP("tcp", nil, tcpAddr)
	if err2 != nil {
		return
	}

	var tr Target
	//var errtr error
	tr, _ = NewOpenTargeter("open", value, uint64(fileSize)) //拼凑成协议命令

	_, err3 := conn.Write([]byte(tr.Command))
	if err3 != nil {

		fmt.Println("发送创建命令失败,错误是:", err3)
		return

	}
	//data := make([]byte, 1024)
	for {

		err := ParseOpenProtocol(conn)
		if err != nil {
			fmt.Printf("ERROR!!!在创建文件%v,%v\n", value, err)
			//conn.Close()
			return
		} else {
			fmt.Println("\nsuccess:成功创建文件:", value)
			break
			//fmt.Printf("收到的数据是:")
			//fmt.Println(string(data))
		}

	}
	defer conn.Close()

}

func (a *Attacker) buildTCP(conn *net.TCPConn) error {
	tcpAddr, err1 := net.ResolveTCPAddr("tcp4", a.remote)
	if err1 != nil {
		return err1
	}
	conn, err2 := net.DialTCP("tcp", nil, tcpAddr)
	if err2 != nil {
		return err2
	}
	return nil
}
func (a *Attacker) attack(tr Target, workers *sync.WaitGroup) {
	workers.Add(1)
	defer workers.Done()
	//创建socket
	tcpAddr, err1 := net.ResolveTCPAddr("tcp4", a.remote)
	if err1 != nil {
		fmt.Println(err1)
		return
	}
	conn, err2 := net.DialTCP("tcp", nil, tcpAddr)

	//conn, err := a.BulidTcpV4()
	if err2 != nil {
		fmt.Println("拨号错误")
		fmt.Println(err2)
		return
	}
	_, err3 := conn.Write([]byte(tr.Command))
	if err3 != nil {
		fmt.Println("发送错误")
		fmt.Println(err3)
		return
	}
	//fmt.Printf("正在创建第%v个客户端\n", number)
	fmt.Print("发送的客户端数据是:", tr.Command)
	for {
		fmt.Print("等待接收服务端的数据")
		result, err4 := ioutil.ReadAll(conn)
		if err4 != nil {
			fmt.Println(err4)
			return
		}
		fmt.Println(result)
	}

}

/*
func (a *Attacker) BulidTcpV4() (*TCPConn, error) {
	tcpAddr, err := net.ResolveTCPAddr("tcp4", a.remote)
	if err != nil {
		return nil, err
	}
	conn, err := net.DialTCP("tcp", nil, tcpAddr)

	return conn, err

}
*/
/*
func Length(n uint64) func(*Attacker) {
	return func (a *Attacker) {
		a.length = n
	}

}

func Offset(n uint64) func(*Attacker){
	return func (a *Attacker) {
		a.offset = n

	}
}


func FileName(file string) func(*Attacker){
     return func(a *Attacker){
     	a.fileName = file
     }
}

func GetProtocol(pro string) func (*Attacker) {
	return func(a *Attacker){
		a.protocol = pro
	}

}*/
