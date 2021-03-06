package bing

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/lxn/walk"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

const base_bing_url = "https://cn.bing.com"
const DEFAULT_PATH = "./bing壁纸"

type ItemImage struct {
	Url       string `json:"url"`
	Name      string `json:"copyright"`
	StartDate string `json:"startdate"`
}
type MetaData struct {
	Images []ItemImage `json:"images"`
}

func IntToBytes(n int) []byte {
	data := int64(n)
	bytebuf := bytes.NewBuffer([]byte{})
	binary.Write(bytebuf, binary.BigEndian, data)
	return bytebuf.Bytes()
}
func DownloadAllData(dirPath string, data *MetaData, infoBytes *walk.TextEdit) *sync.WaitGroup {
	wg := new(sync.WaitGroup)
	i := 1
	isHasDownload := false
	for _, item := range data.Images {
		targetNameBuild := new(strings.Builder)
		targetNameBuild.WriteString(item.StartDate[:4])
		targetNameBuild.WriteString("-")
		targetNameBuild.WriteString(item.StartDate[4:6])
		targetNameBuild.WriteString("-")
		targetNameBuild.WriteString(item.StartDate[6:8])
		targetNameBuild.WriteString(item.Name)
		targetNameBuild.WriteString(".jpg")
		targeetName := targetNameBuild.String()
		targetPath := filepath.Join(dirPath, targeetName)
		if _, err := os.Stat(targetPath); os.IsNotExist(err) {
			isHasDownload = true
			wg.Add(1)
			infoBytes.AppendText(strconv.Itoa(i))
			i++
			infoBytes.AppendText(". ")
			infoBytes.AppendText(targeetName)
			infoBytes.AppendText("\r\n")
			go DownloadUrlFile(targetPath, item.Url, wg)
		}
	}
	if !isHasDownload {
		infoBytes.AppendText("没有新的壁纸了\r\n")
	}
	return wg
}

func DownloadUrlFile(saveFileName string, url string, wg *sync.WaitGroup) {
	defer wg.Done()

	if req, err := http.Get(url); err == nil {
		body := req.Body
		file, err := os.Create(saveFileName)
		defer file.Close()
		if err == nil {
			defer body.Close()
			io.CopyBuffer(file, body, nil)
		} else {
			fmt.Println("存储文件错误", err)
		}
	} else {
		fmt.Println("下载错误", url)
	}
}

func FixData(data *MetaData) {
	var buffer bytes.Buffer
	images := data.Images
	for i := range images {
		buffer.Reset()
		buffer.WriteString(base_bing_url)
		buffer.WriteString(images[i].Url)
		images[i].Name = regexp.MustCompile("([^-\u4e00-\u9fa5，！＠＆＃、“”（）【】《》〖〗『』。])+").ReplaceAllString(images[i].Name, "")
		images[i].Url = buffer.String()
	}
}

func EnsureDir(path string) {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		os.MkdirAll(path, os.ModeDir)
	}
}
func ClickDownload(infoLabel *walk.TextEdit) {
	if get, err := http.Get("https://cn.bing.com/HPImageArchive.aspx?format=js&idx=1&n=10"); err == nil {
		body := get.Body
		defer body.Close()
		data := new(MetaData)
		if err := json.NewDecoder(body).Decode(data); err != nil {
			infoLabel.AppendText("error: ")
			infoLabel.AppendText(err.Error())
		} else {
			FixData(data)
			EnsureDir(DEFAULT_PATH)
			infoLabel.AppendText("--------start-------------\r\n")
			waitGroup := DownloadAllData(DEFAULT_PATH, data, infoLabel)
			infoLabel.AppendText("--------end-------------\r\n")
			waitGroup.Wait()
		}
	}
}
