package main

import (
	"bufio"
	"container/list"
	"errors"
	"fmt"
	"hash/crc32"
	"io"
	"log"
	"math/rand"
	"net"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

var EOF = errors.New("EOF")
var ErrShortWrite = errors.New("short write")
var proxys = list.New()
var proxyMap = make(map[string]int64)
var index int
var lock sync.Mutex

func debug() {
	for e := proxys.Front(); e != nil; e = e.Next() {
		log.Println(e.Value)
	}
}

func check() {

	for {
		diff := proxys.Len() - index
		log.Println("start check conn , the diff is :", diff)
		endIndex := proxys.Len()
		for i := 0; i < diff; i++ {
			endIndex -= 1
			_err, ip := getAllIPByIndex(endIndex)
			log.Println(_err)
			if _err == nil {
				_conn, err := net.Dial("tcp", ip)
				if err == nil {
					moveEnableServerToFront(endIndex)
				}
				if _conn != nil {
					_conn.Close()
				}
			}

		}
		time.Sleep(30 * time.Second)
	}
}

func getProxyIPByIndex(ii int) (error, string) {
	var ip string
	i := 0
	for e := proxys.Front(); e != nil && i < index; e = e.Next() {
		if i == ii {
			ip = e.Value.(string)
			break
		}
		i++
	}
	if ip == "" {
		errMessage := "index get ip faile ,the index is:" + strconv.Itoa(ii)
		return errors.New(errMessage), ""
	} else {
		return nil, ip
	}
}

func getAllIPByIndex(ii int) (error, string) {
	var ip string
	i := 0
	for e := proxys.Front(); e != nil; e = e.Next() {
		if i == ii {
			ip = e.Value.(string)
			break
		}
		i++
	}
	if ip == "" {
		errMessage := "index get ip faile ,the index is:" + strconv.Itoa(ii)
		return errors.New(errMessage), ""
	} else {
		return nil, ip
	}
}

func moveEnableServerToFront(position int) {
	defer lock.Unlock()
	lock.Lock()
	i := 0
	for e := proxys.Front(); e != nil; e = e.Next() {
		if i == position {
			proxys.MoveToFront(e)
			log.Println("move to front:", e.Value.(string))
			break
		}
		i++
	}

	index++
	if index > proxys.Len() {
		index = proxys.Len()
	}
}

func moveDeadServiceToBack(position int) {
	defer lock.Unlock()
	lock.Lock()
	i := 0
	for e := proxys.Front(); e != nil; e = e.Next() {
		if i == position {
			proxys.MoveToBack(e)
			log.Println("move to back:", e.Value.(string))
			break
		}
		i++
	}

	if index > 0 {
		index--
	}
}

func start(ip string, port string, proxyType string) {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%s", ip, port))
	if err != nil {
		log.Printf("listen tcp :%s", err.Error())
		log.Println("socket Proxy Will Exit ...")
		os.Exit(1)
	}
	log.Println("The service listening on:", ip, ":", port)
	defer listener.Close()

	debug()
	index = proxys.Len()
	go check()
	for {
		tc, err := listener.Accept()
		if err != nil {
			log.Println("accept tcp conn :", err.Error())
			if tc != nil {
				tc.Close()
			}
			continue
		}

		if index == 0 {
			log.Println("No valid server. Please check server and restart the proxy server.")
			if tc != nil {
				tc.Close()
			}
			continue
		}

		accessIP := tc.RemoteAddr()
		log.Println("accept new connecion:", accessIP)
		switch proxyType {
		case "ipHash":
			ipHash := crc32.ChecksumIEEE([]byte(accessIP.String()))
			log.Println("ipHash:", ipHash)
			position := int(ipHash) % index
			go connect(position, &tc)
		case "random":
			position := rand.Intn(index)
			go connect(position, &tc)
		default:
			log.Println("not support proxy type.")
		}
	}
}

func connect(position int, tc *net.Conn) {
	err, remoteAddress := getProxyIPByIndex(position)
	if err == nil {
		log.Println("connection to :", remoteAddress)
		uc, err := net.Dial("tcp", remoteAddress)
		if err != nil {
			log.Println(err.Error())
			if (*tc) != nil {
				(*tc).Close()
			}
			moveDeadServiceToBack(position)
			return
		}
		go send(tc, &uc)
		go revice(tc, &uc)
	}
}

func send(tc *net.Conn, uc *net.Conn) {
	io.Copy(*uc, *tc)
	if (*uc) != nil {
		(*uc).Close()
	}

}

func revice(tc *net.Conn, uc *net.Conn) {
	io.Copy(*tc, *uc)
	if (*tc) != nil {
		(*tc).Close()
	}
}

func loadConfig() (string, string, string) {
	var ip string
	var port string
6	var proxyType string

	f, err := os.Open("./proxy.ini")
	if err != nil {
		panic("读取配置文件失败.")
	}

	defer f.Close()
	br := bufio.NewReader(f)

	for {
		line, err := br.ReadString('\n')
		if err == io.EOF {
			break
		}
		if 0 == len(line) || line == "\n" {
			continue
		}
		if strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "ip") {
			ip = strings.Trim(strings.Split(line, "=")[1], " ")
			ip = strings.Replace(ip, "\n", "", -1)
			continue
		}
		if strings.HasPrefix(line, "port") {
			port = strings.Trim(strings.Split(line, "=")[1], " ")
			port = strings.Replace(port, "\n", "", -1)
			continue
		}
		if strings.HasPrefix(line, "proxyType") {
			proxyType = strings.Trim(strings.Split(line, "=")[1], " ")
			proxyType = strings.Replace(proxyType, "\n", "", -1)
			continue
		}
		line = strings.Replace(line, "\n", "", -1)
		proxys.PushBack(line)

	}
	return ip, port, proxyType
}

func main() {
	ip, port, proxyType := loadConfig()
	coreCPU := runtime.NumCPU()/2 + 1
	log.Println("system cpu core is:", runtime.NumCPU())
	log.Println("set cpu num is :", coreCPU)
	runtime.GOMAXPROCS(coreCPU)
	start(ip, port, proxyType)

}

/*
#proxy.ini
#load balancing server ip and port config
ip=0.0.0.0
port=9527
#ipHash,random
proxyType=random
#proxy server config
127.0.0.1:7070
127.0.0.1:9090
127.0.0.1:8080
#proxy config end
*/
