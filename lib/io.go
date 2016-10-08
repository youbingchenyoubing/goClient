package cache

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

// 按照行读取文件
func ReadLine(fileName string, handler func(string)) error {
	f, err := os.Open(fileName)
	if err != nil {
		return err
	}
	defer f.Close()
	inputReader := bufio.NewReader(f)
	for {
		inputString, readError := inputReader.ReadString('\n')
		if readError != nil {
			if readError == io.EOF {
				return nil
			}
			return readError
		}
		inputString = strings.Replace(inputString, "\n", "", -1)
		fmt.Println("The input was", inputString)
		handler(inputString)

	}
	return nil
}

/*
func GetFileSize(File string) (uint64, error) {
	fileInfo, err := os.Stat(File)
	if err != nil {
		return 0, err
	}
	return fileInfo, nil
}
*/

// 读取文件部分内容(Seek 函数和Read函数组合)
func ReadFilePart(File string, offset uint64, length uint64) (string, uint64) {

	buff := make([]byte, length) // 创建切片引用

	fi, err := os.Open(File)

	if err != nil {
		fmt.Println("打开文件错误,请确认文件是否存在")
		panic(err)

	}
	defer fi.Close()
	fi.Seek(int64(offset), 0) //相当于开始位置偏移
	//realRead, err1 := fi.ReadAt(buff, int64(length)) //读取文件的固定长度
	realRead, err1 := fi.Read(buff) //读取文件
	//fmt.Println("offset=%v,length=%v,realRead=%v", offset, length, realRead)
	if err1 != nil {
		fmt.Println("ERROR!!!读取发生错误")
		panic(err1)
	}

	if int64(realRead) != int64(length) {
		fmt.Printf("WARNING!!!,实际长度和要读的长度不一致,本来要读的是:%v,结果只读了:%v,很可能是文件末尾", length, realRead)
		buff1 := buff[:realRead]
		return string(buff1), uint64(realRead)

	}
	//fmt.Printf("realRead=%v", len(string(buff)))
	return string(buff), uint64(realRead)

}

// 跟上面实现的是一样的功能，不过只是用ReadAt函数实现
func ReadFilePart2(File string, offset uint64, length uint64) (string, uint64) {

	buff := make([]byte, length) // 创建切片引用

	fi, err := os.Open(File)

	if err != nil {
		fmt.Println("打开文件错误,请确认文件是否存在")
		panic(err)

	}
	defer fi.Close()
	//fi.Seek(int64(offset), 0) //相当于开始位置偏移
	//realRead, err1 := fi.ReadAt(buff, int64(length)) //读取文件的固定长度
	realRead, err1 := fi.ReadAt(buff, int64(offset)) //读取文件
	//fmt.Println("offset=%v,length=%v,realRead=%v", offset, length, realRead)
	if err1 != nil {
		fmt.Println("ERROR!!!读取发生错误")
		panic(err1)
	}

	if int64(realRead) != int64(length) {
		fmt.Println("WARNING!!!,实际长度和要读的长度不一致,本来要读的是:%v,结果只读了:%v", length, realRead)

	}
	return string(buff), uint64(realRead)
}
