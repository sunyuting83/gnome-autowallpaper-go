package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"
	"unsafe"
)

func GetAllFile(pathname string) []string {
	// 获取文件目录下所有文件
	a := []string{}
	rd, err := ioutil.ReadDir(pathname)
	if err != nil {
		fmt.Println("read dir fail:", err)
		return []string{}
	}
	for _, fi := range rd {
		if fi.IsDir() {
			// fmt.Printf("[%s]\n", pathname+"/"+fi.Name())
			GetAllFile(strings.Join([]string{pathname, fi.Name(), "/"}, ""))
		} else {
			var fileSuffix string
			fileSuffix = path.Ext(fi.Name())
			if fileSuffix == ".jpg" || fileSuffix == ".jpeg" {
				a = append(a, strings.Join([]string{pathname, fi.Name()}, "/"))
			}
		}
	}
	// fmt.Println(a)
	// return a
	return a
}

func getLast() (int, string) {
	// 获取最后一次设置背景图的数值 文件不存在创建0
	home, err := homeUnix()
	if err != nil {
		fmt.Println(err)
	}
	lastpath := strings.Join([]string{home, ".config/wallpaperlast"}, "/")
	// fmt.Println(home, last)
	check := IsExist(lastpath)
	if check {
		number, err := ioutil.ReadFile(lastpath)
		if err != nil {
			return 0, lastpath
		}
		return Byte2Int(number), lastpath
	} else {
		os.Create(lastpath)
		number := Int2Byte(0)
		ioutil.WriteFile(lastpath, number, 0644)
		return 0, lastpath
	}
}

func Int2Byte(data int) (ret []byte) {
	// 数字转Byte
	var len uintptr = unsafe.Sizeof(data)
	ret = make([]byte, len)
	var tmp int = 0xff
	var index uint = 0
	for index = 0; index < uint(len); index++ {
		ret[index] = byte((tmp << (index * 8) & data) >> (index * 8))
	}
	return ret
}

func Byte2Int(data []byte) int {
	// Byte转数字
	var ret int = 0
	var len int = len(data)
	var i uint = 0
	for i = 0; i < uint(len); i++ {
		ret = ret | (int(data[i]) << (i * 8))
	}
	return ret
}

func IsExist(path string) bool {
	// 判断文件是否存在
	_, err := os.Stat(path)
	return err == nil || os.IsExist(err)
}

func homeUnix() (string, error) {
	// 获取用户文件目录Unix
	// First prefer the HOME environmental variable
	if home := os.Getenv("HOME"); home != "" {
		return home, nil
	}

	// If that fails, try the shell
	var stdout bytes.Buffer
	cmd := exec.Command("sh", "-c", "eval echo ~$USER")
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		return "", err
	}

	result := strings.TrimSpace(stdout.String())
	if result == "" {
		return "", errors.New("blank output when reading home directory")
	}

	return result, nil
}

func saveLast(number int, lastpath string) {
	// 保存最后一次设置数字
	i := Int2Byte(number)
	ioutil.WriteFile(lastpath, i, 0644)
	return
}

func getNow(hours int) bool {
	if hours >= 6 && hours <= 18 {
		return true
	}
	return false
}

func setWallpaper(path string) {
	// 获取当前时间
	Hours := time.Now().Hour()
	now := getNow(Hours)
	// 设置壁纸
	dorn := "picture-uri-dark"
	if now {
		dorn = "picture-dark"
	}
	cmd := exec.Command("gsettings", "set", "org.gnome.desktop.background", dorn, path)
	if err := cmd.Run(); err != nil {
		fmt.Println(err)
		return
	}
	return
}

func start(path string, outtime int) {
	// 开始
	list := GetAllFile(path)    // 目录下所有图片
	length := len(list)         // 所有图片数量
	last, lastpath := getLast() // 获取最后设置数值
	for i := last; i <= length; i++ {
		if i == length {
			saveLast(0, lastpath) // 保存最后设置数值
			start(path, outtime)  //重新开始
		}
		setWallpaper(list[i])                            //设置壁纸
		saveLast(i, lastpath)                            //保存最后设置数值
		time.Sleep(time.Duration(outtime) * time.Second) //延迟换壁纸
	}
	return
}

func main() {
	home, err := homeUnix() //获取用户目录
	if err != nil {
		fmt.Println(err)
	}
	imgpath := strings.Join([]string{home, "图片"}, "/")
	message := strings.Join([]string{"图片目录默认是", imgpath}, ":")
	var path string
	var outtime int
	flag.StringVar(&path, "p", imgpath, message)
	flag.IntVar(&outtime, "t", 300, "默认更换时间是300秒")
	flag.Parse()
	// fmt.Println(path)
	start(path, outtime)
}
