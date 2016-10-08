package cache

import (
	"bytes"
	"errors"
	//"fmt"
	"log"
	"net"
	"strconv"
	"strings"
)

type cacheHeader struct {
	code   string
	offset uint64
	length uint64
}

//解析读的协议
func ParseReadProtocol(data []byte, offset uint64, plen uint64, conn *net.TCPConn, errorLog *log.Logger) (uint64, error) {
	count := uint64(0)
	end := uint64(0)
	buffer := make([]byte, 128)
	//var err error
	//开始接收数据
	for {
		header, body, err := parseData("read", conn, buffer, &end)
		if err != nil {
			err := errors.New("已经读取字节数为:" + strconv.FormatUint(count, 10) + " " + err.Error())
			return 0, err
		}
		if !strings.EqualFold(header.code, "102") {
			err := errors.New("已经读取字节数为:" + strconv.FormatUint(count, 10) + " 错误代码:" + header.code)
			return 0, err
		}
		if end > 0 {
			var leftbuffer bytes.Buffer
			leftbuffer.Write(buffer[0:end])
			localAddress := conn.LocalAddr()
			errorLog.Printf("IP:%v,多读取的内容是:%v\n", localAddress.String(), leftbuffer.String())
		}
		if body != nil {
			var verifyBuffer bytes.Buffer
			slicelen := len(body)
			if header.length != uint64(slicelen) {
				err := errors.New("header.length=" + strconv.FormatUint(header.length, 10) + "!=body.length=" + strconv.Itoa(slicelen))
				return 0, err
			}
			verifyBody := body[slicelen-2 : slicelen] //  not  include position len
			verifyBuffer.Write(verifyBody)
			if verifyBuffer.String() == "\r\n" {
				copy(data[header.offset-offset:], body[0:slicelen-2])
				count += uint64(slicelen - 2)
			} else {
				err = errors.New("包体校验不成功," + "头部表明包体的长度: " + strconv.FormatUint(header.length, 10) + " 已经接收的包体长度是:" + strconv.FormatUint(count, 10) + " 包体的长度是:" + strconv.Itoa(slicelen) + " 包体校验码是:" + verifyBuffer.String())
				return 0, err
			}
		}
		if count >= plen {
			break
		}
	}
	return count, nil
}

//解析写应答
func ParseWriteProtocol(conn *net.TCPConn) error {
	buffer := make([]byte, 128)
	end := uint64(0)
	header, _, err := parseData("write", conn, buffer, &end)
	if err != nil {
		return err
	}
	if !strings.EqualFold(header.code, "300") {
		err := errors.New("写数据出现错误了，错误代码为:" + header.code)
		return err
	}
	return nil

}

//解析创建文件应答
func ParseOpenProtocol(conn *net.TCPConn) error {
	buffer := make([]byte, 128)
	end := uint64(0)
	header, _, err := parseData("open", conn, buffer, &end)
	if err != nil {
		return err
	}
	if !strings.EqualFold(header.code, "403") {
		err := errors.New("创建文件失败，错误代码是:" + header.code)
		return err
	}
	return nil
}

//将头部和body的数据分离
func parseData(protocol string, conn *net.TCPConn, buffer []byte, end *uint64) (*cacheHeader, []byte, error) {
	var allbuffer bytes.Buffer

	var index int
	if (*end) > 0 {
		allbuffer.Write(buffer[:*end])
		//fileName:="test_log"
	}
	*end = 0
	for {

		index = strings.Index(allbuffer.String(), "\r\n") //找到头部结尾的标记
		if index != -1 {
			break
		}

		newbufLen := uint64(allbuffer.Len())
		if newbufLen >= 128 {
			err := errors.New("头部接收缓冲区都满了,还未找到头部")
			return nil, nil, err
		}
		//fmt.Println("读取数据")
		ncount, err := conn.Read(buffer)
		//fmt.Println("test")
		if err != nil {
			return nil, nil, err
		}
		if ncount != 0 {
			allbuffer.Write(buffer[:ncount])
		} else {
			err = errors.New("接收数据长度为0,服务端的数据未发送过来")
			return nil, nil, err
		}

	}
	hb := allbuffer.Next(index + 2) //这句话已经把buffer里面的数据读取出来了
	header, err := parseHeader(hb[:index])
	if err != nil {
		return nil, nil, err
	}
	buflen := uint64(allbuffer.Len()) //多于的包体内容
	if protocol != "read" {           //不是读的话，接收包只有头部
		//fmt.Println("allbuffer=", allbuffer.String())
		if buflen > 0 && buflen < 128 {
			copy(buffer[0:], allbuffer.Bytes()[0:buflen]) //因为这里面没有包体
			//begin = 0
			*end = buflen
		} else if buflen >= 128 {
			err := errors.New("多读的信息太多了,无法保存")
			return header, nil, err
		}
		return header, nil, nil
	}
	// 开始接收剩余包体

	if header.length > buflen {
		d := make([]byte, header.length-buflen)
		count := uint64(0)
		remain := header.length - buflen
		for {
			n, err := conn.Read(d[count:])
			if err != nil {
				//fmt.Println("err=", err)
				return header, nil, err
			}
			if n != 0 {
				count += uint64(n)
				if count == remain {
					allbuffer.Write(d)
					break
				}
			} else {
				err = errors.New("接收错误,服务器关闭")
				return nil, nil, err
			}
		} //end for
	} else if buflen > header.length {
		if buflen > 0 && buflen < 128 {
			realbody := allbuffer.Bytes()
			copy(buffer[0:], realbody[header.length:buflen])

			*end = buflen - header.length

			return header, realbody[:header.length], nil
		} else if buflen >= 128 {
			err := errors.New("多读的信息太多了,无法保存")
			return header, nil, err
		}
	}
	//begin = 0
	*end = 0
	return header, allbuffer.Bytes(), nil

}

//将头部的信息拿出来
func parseHeader(h []byte) (header *cacheHeader, err error) {
	header = new(cacheHeader)

	//以某种模式分割字符串
	//fmt.Println("头部是:", string(h))
	s := strings.Split(string(h), "/")
	n := len(s)
	//fmt.Println("分割后长度为:", n)
	if n < 1 || (n != 1 && n != 3) {
		err = errors.New("解析头部:" + string(h) + "出现错误")
		return nil, err
	} else if n == 1 {
		header.code = s[0]
	} else {
		header.code = s[0]
		header.offset, err = strconv.ParseUint(s[1], 10, 64)
		if err != nil {
			err = errors.New("头部有问题,当转换offset:" + string(h)) //测试
			return nil, err
		}
		header.length, err = strconv.ParseUint(s[2], 10, 64)
		if err != nil {
			err = errors.New("头部有问题,当转换length:" + string(h)) //测试
			return nil, err
		}
	} //end  else
	return header, nil
}
