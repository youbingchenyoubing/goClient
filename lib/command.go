package cache

import (
	//"fmt"
	"bytes"
	"errors"
	"strconv"
)

type Target struct {
	Protocol string
	Command  string
	Length   uint64
	Offset   uint64
	Socket   string
	FileName string
}

//type Targeter func(*Target) error
var (
	ErrInvalidPara = errors.New("Invaid parameters when generate command")
	ErrInvalidData = errors.New("Invalid body data when generate command")
)

func NewReadTargeter(Pro string, Len uint64, Off uint64, Fil string) (Target, error) {

	var tar Target
	//err  error
	if Pro == "" || Fil == "" || Len <= 0 {

		return tar, ErrInvalidPara
	}
	tar.Protocol = Pro

	tar.Length = Len

	tar.Offset = Off

	tar.FileName = Fil

	//tar.Socket = Soc

	var buffer bytes.Buffer

	buffer.WriteString(tar.Protocol)

	buffer.WriteString("/")

	buffer.WriteString(tar.FileName)

	buffer.WriteString("/")

	buffer.WriteString(strconv.FormatUint(tar.Offset, 10))

	buffer.WriteString("/")

	buffer.WriteString(strconv.FormatUint(tar.Length, 10))

	buffer.WriteString("\r\n")

	tar.Command = buffer.String()
	return tar, nil

}

//写的协议
func NewWriteTargeter(Pro string, Len uint64, Off uint64, Fil string, Data string) (Target, error) {

	var tar Target
	//err  error
	if Pro == "" || Fil == "" || Len <= 0 {

		return tar, ErrInvalidPara
	}
	realLength := len(Data)
	if uint64(realLength) != Len {
		return tar, ErrInvalidData
	}
	tar.Protocol = Pro

	tar.Length = Len

	tar.Offset = Off

	tar.FileName = Fil

	//tar.Socket = Soc

	var buffer bytes.Buffer

	buffer.WriteString(tar.Protocol)

	buffer.WriteString("/")

	buffer.WriteString(tar.FileName)

	buffer.WriteString("/")

	buffer.WriteString(strconv.FormatUint(tar.Offset, 10))

	buffer.WriteString("/")

	buffer.WriteString(strconv.FormatUint(tar.Length, 10))

	buffer.WriteString("\r\n")

	buffer.WriteString(Data)

	tar.Command = buffer.String()
	return tar, nil
}

//写的协议自己读取本地的文件（测试时使用）
func NewWriteTargeterNoData(Pro string, Len uint64, Off uint64, Fil string) (Target, error) {

	var tar Target
	//err  error
	if Pro == "" || Fil == "" || Len <= 0 {

		return tar, ErrInvalidPara
	}
	tar.Protocol = Pro

	//tar.Length = Len

	tar.Offset = Off

	tar.FileName = Fil

	//tar.Socket = Soc
	bodyData, newlen := ReadFilePart(Fil, Off, Len) // 读取文件的包体
	tar.Length = newlen

	var buffer bytes.Buffer

	buffer.WriteString(tar.Protocol)

	buffer.WriteString("/")

	buffer.WriteString(tar.FileName)

	buffer.WriteString("/")

	buffer.WriteString(strconv.FormatUint(tar.Offset, 10))

	buffer.WriteString("/")

	buffer.WriteString(strconv.FormatUint(newlen, 10))

	buffer.WriteString("\r\n")

	buffer.WriteString(bodyData) //拼接包体

	tar.Command = buffer.String()
	return tar, nil
}

// 建立文件的协议 Pro = open Fil 文件名 Len(文件大小)
func NewOpenTargeter(Pro string, Fil string, Len uint64) (Target, error) {

	var tar Target
	//err  error
	if Pro == "" || Fil == "" || Len <= 0 {

		return tar, ErrInvalidPara
	}
	tar.Protocol = Pro

	tar.Length = Len

	//tar.Offset = Off

	tar.FileName = Fil

	//tar.Socket = Soc

	var buffer bytes.Buffer

	buffer.WriteString(tar.Protocol)

	buffer.WriteString("/")

	buffer.WriteString(tar.FileName)

	buffer.WriteString("/")

	//buffer.WriteString(strconv.FormatUint(tar.Offset, 10))

	//buffer.WriteString("/")

	buffer.WriteString(strconv.FormatUint(tar.Length, 10))

	buffer.WriteString("\r\n")

	//bodyData := ReadFilePart(Fil, Off, Len) // 读取文件的包体

	//buffer.WriteString(bodyData) //拼接包体

	tar.Command = buffer.String()
	return tar, nil
}

//两种协议的结合
func NewCommand(Pro string, Len uint64, Off uint64, Fil string) (Target, error) {
	var tar Target
	var err error
	if Pro == "read" {

		return NewReadTargeter(Pro, Len, Off, Fil)

	} else if Pro == "write" {
		return NewWriteTargeterNoData(Pro, Len, Off, Fil)
	} else {
		return tar, err
	}

}
