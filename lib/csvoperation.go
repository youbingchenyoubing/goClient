package cache

import (
	"encoding/csv"
	"fmt"
	"os"
)

func WriteCsv(fileName string, info []string) error {

	f, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		return err
	}
	defer f.Close()
	//n, err := f.Seek(0, os.SEEK_END)

	r := csv.NewReader(f)
	files, err := r.ReadAll()
	if err != nil {
		return err
	}
	//files = append(files, []string{"4", "赵六", "26"})
	f.WriteString("\xEF\xBB\xBF") // 写入UTF-8 BOM
	fmt.Println("files=", files)
	w := csv.NewWriter(f)

	w.Write(info)
	w.Flush()
	return nil
}
