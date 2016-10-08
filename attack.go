package main

import (
	//"crypto/tls"
	//"crypto/x509"
	"errors"
	"flag"
	//"fmt"
	//"io"
	//"io/ioutil"
	//"net/http"
	"os"
	"os/signal"
	//"time"

	vegeta "youbingchen/goClient/lib"
)

func attackCmd() command {
	//fmt.Println("hello attack")
	fs := flag.NewFlagSet("goClient", flag.ExitOnError)
	opts := &attackOpts{

	//laddr: localAddr{&vegeta.DefaultLocalAddr},
	}
	//fs.StringVar(&opts.data, "data", "", "for write")
	fs.StringVar(&opts.targetsf, "targets", "115.28.70.190", "Targets IP")
	fs.StringVar(&opts.port, "port", "9999", "Targets  Port")
	fs.StringVar(&opts.protocol, "protocol", "read", "read or write or open")
	fs.StringVar(&opts.fileName, "file", "a.txt", "the name of file you want to read  or write")
	fs.Uint64Var(&opts.offset, "offset", 0, "the offset you want begin to read or write")
	fs.Uint64Var(&opts.length, "length", 1, "the length you want read or write ")
	fs.Uint64Var(&opts.rate, "rate", 50, "Requests per second")
	fs.Uint64Var(&opts.workers, "workers", vegeta.DefaultWorkers, "Initial number of workers")
	fs.BoolVar(&opts.function, "function", true, "choose the function")

	return command{fs, func(args []string) error {
		fs.Parse(args)
		if opts.function == true {
			return attack(opts)
		} else {
			return attackDiff(opts)
		}
	}}

}

var (
	errZeroRate = errors.New("rate must be bigger than zero")
	//errBadCert  = errors.New("bad certificate")
	errProtocol = errors.New("bad  protocols")
)

// attackOpts aggregates the attack function command options
type attackOpts struct {
	targetsf string //ip地址
	port     string //端口号
	protocol string //协议
	fileName string // 这个是根据自己的需求进行定义的
	offset   uint64 //同上
	length   uint64 //同上

	//outputf     string
	//bodyf       string
	//certf       string
	//keyf        string
	//rootCerts   csl
	//http2       bool
	//insecure    bool
	//lazy        bool
	//duration    time.Duration
	//timeout     time.Duration
	//rate        uint64
	rate    uint64 // 攻击频率
	workers uint64 //开多少协程进行攻击
	//connections int
	//redirects   int
	//headers     headers
	//laddr localAddr
	//	keepalive   bool
	function bool //功能选择
}

func attackDiff(opts *attackOpts) (err error) {
	if opts.protocol != "read" && opts.protocol != "write" && opts.protocol != "open" {
		return errProtocol
	}
	atk := vegeta.NewAttacker(
		vegeta.Addr(opts.targetsf, opts.port),
		vegeta.Workers(opts.workers),
		vegeta.SetFileName(opts.fileName),
		vegeta.SetBlockSize(opts.length),
	)
	if opts.protocol == "read" {
		var fileName string
		if opts.port == "1234" { //我写的cache端口
			fileName = "cacheManager_result.csv"
		} else if opts.port == "12345" { //直接读并且no page cache
			fileName = "nopageCache_result.csv"

		} else {
			fileName = "pageCache_result.csv"
		}
		_, err := os.Stat(fileName)
		if err != nil {
			vegeta.WriteCsv(fileName, []string{"每个文件的协程数", "成功的协程数", "平均协程运行时间(s)"})
		}
		var f = atk.AttackRead()
		f(fileName)
	} else if opts.protocol == "write" {
		atk.AttackWrite()
	} else {
		atk.AttackOpen()
	}
	return nil
}

// attack validates the attack arguments, sets up the
// required resources, launches the attack and writes the results
func attack(opts *attackOpts) (err error) {
	if opts.rate == 0 {
		return errZeroRate
	}
	if opts.protocol != "read" && opts.protocol != "write" && opts.protocol != "open" {
		return errProtocol
	}
	var (
		tr vegeta.Target
	)

	atk := vegeta.NewAttacker(
		vegeta.Addr(opts.targetsf, opts.port),
		vegeta.Workers(opts.workers),
	)
	if opts.protocol == "read" {
		//fmt.Println("H")
		tr, err = vegeta.NewReadTargeter(opts.protocol, opts.length, opts.offset, opts.fileName)

		if err != nil {
			return err
		}
	} else if opts.protocol == "write" {
		tr, err = vegeta.NewWriteTargeterNoData(opts.protocol, opts.length, opts.offset, opts.fileName)

		if err != nil {
			return err
		}
	} else if opts.protocol == "open" {
		tr, err = vegeta.NewOpenTargeter(opts.protocol, opts.fileName, opts.offset)

		if err != nil {
			return err
		}

	}
	atk.Attack(tr)
	//enc := vegeta.NewEncoder(out)
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)

	for {
		select {
		case <-sig:
			//atk.Stop()
			return nil
		}
	}
}
