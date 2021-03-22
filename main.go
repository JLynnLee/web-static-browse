package main

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
)

var (
	gzipWriter                  *gzip.Writer
	file                        *os.File
	realPath, addressIp, domain string
	denyList                    map[string]int
)

var change = 1
var console = ""
var gzipMap = map[string]int{}

func init() {
	denyList = map[string]int{".idea": 1, "gzip": 1}
}

func main() {
	getIpAddress()
	fmt.Println(addressIp)
	domain = "<a href='http://" + addressIp + ":8080"
	realPath, _ = os.Getwd()

	http.HandleFunc("/", staticService)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}

func staticService(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	isDir := strings.Index(path, ".")
	params := r.URL.Query()
	_, isDownload := params["download"]
	//文件
	if isDir >= 0 {
		requestType := path[strings.LastIndex(path, "."):]
		switch requestType {
		case ".css":
			w.Header().Set("content-type", "text/css")
		case ".js":
			w.Header().Set("content-type", "text/javascript")
		default:
		}

		gzipPath := realPath + path
		if isDownload {
			w.Header().Set("Content-Type", "application/octet-stream")
		}

		if !isDownload {
			if change == 1 {
				w.Header().Set("Content-Encoding", "gzip")
				w.Header().Set("Accept-Encoding", "gzip")
				//gzip操作
				gzipPath = create(gzipPath)
				compress()
				gzipWriter.Close()
				change = 0
			} else {
				change = 1
			}
		}

		fin, err := os.Open(gzipPath)
		if err != nil {
			w.Write([]byte(err.Error()))
		} else {
			fd, _ := ioutil.ReadAll(fin)
			key := ""

			if change == 0 {
				key = path + "默认压缩等级—— "
			} else {
				key = path + "无压缩—— "
			}
			if _, ok := gzipMap[key]; !ok {
				temp := key + strconv.Itoa(len(fd))
				console += "<h5>" + temp + "</h5>"
				gzipMap[key] = 1
			}
			w.Write(fd)
			fin.Close()
		}
	} else { //文件夹
		fileInfoList, err := ioutil.ReadDir(realPath + path)
		if err != nil {
			w.Write([]byte(err.Error()))
		} else {
			//根目录及上级目录
			html := "<html><body>"
			if path != "/" {
				paths := strings.Split(path, "/")
				lens := len(paths)
				if lens > 1 {
					html += domain + "'>/</a></br>"
					lePath := ""
					if lens > 2 {
						for i := 0; i < lens-1; i++ {
							lePath += "/" + paths[i]
						}
						html += domain + "/" + lePath + "'>..</a></br>"
					}
				}
			}
			endA := "<a/>"
			//<a href='http://" + addressIp + ":8080/XXXX/"
			public := domain + path + "/"
			for i := range fileInfoList {
				//名单内的文件夹或文件不可访问
				if _, ok := denyList[fileInfoList[i].Name()]; !ok {
					tmp := public + fileInfoList[i].Name()
					//浏览
					html += tmp + "'>" + fileInfoList[i].Name() + endA
					//下载
					isFile := strings.Index(fileInfoList[i].Name(), ".")
					if isFile != -1 {
						html += tmp + "?download' style='position: absolute;left: 300px;'>" + "download" + endA
					}
					html += "<br/>"
				}
			}
			fd := []byte(html)
			fd = BytesCombine(fd, []byte(console))
			w.Write(fd)
		}

	}
}

func BytesCombine(pBytes ...[]byte) []byte {
	return bytes.Join(pBytes, []byte(""))
}

func getIpAddress() {
	netInterfaces, errs := net.Interfaces()
	if errs != nil {
		fmt.Println("net.Interfaces failed, err:", errs.Error())
	}

	for i := 0; i < len(netInterfaces); i++ {
		if (netInterfaces[i].Flags & net.FlagUp) != 0 {
			addrs, _ := netInterfaces[i].Addrs()

			for _, address := range addrs {
				if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
					if ipnet.IP.To4() != nil {
						addressIp = ipnet.IP.String()
					}
				}
			}
		}
	}
}

func create(filePath string) string {
	//初始化创建一个压缩文件
	h := md5.New()
	h.Write([]byte(filePath))
	cipherStr := h.Sum(nil)
	tmpPath := realPath + "/gzip/" + hex.EncodeToString(cipherStr) + ".gz"
	outputFile, err := os.Create(tmpPath)
	if err != nil {
		log.Fatal(err)
	}
	gzipWriter = gzip.NewWriter(outputFile)
	//打开普通文件
	file, err = os.Open(filePath)
	if err != nil {
		panic(err)
	}

	return tmpPath
}

func compress() {
	reader := bufio.NewReader(file)
	for {
		s, e := reader.ReadString('\n')
		if e == io.EOF {
			break
		}
		// 写入gzip writer数据时，它会依次压缩数据并写入到底层的文件中。
		_, err := gzipWriter.Write([]byte(s))
		if err != nil {
			log.Fatal(err)
		}
	}
}
