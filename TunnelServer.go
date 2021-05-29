package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
)

var (
	ssPort int
	ccPort int
	ddPort int
)

func init() {
	flag.IntVar(&ssPort, "s", 3330, "the user listen port")
	flag.IntVar(&ccPort, "c", 3331, "client control listen port")
	flag.IntVar(&ddPort, "d", 3332, "client data listen port")
}

func cc(notifyChan chan bool) {
	defer func() {
		err := recover()
		if err != nil {
			log.Println(err)
		}
	}()

	//监听客户端控制
	ccListener, err := net.Listen("tcp", fmt.Sprintf(":%d", ccPort))
	if err != nil {
		panic(err)
	}
	log.Printf("监听:%d端口, 等待client控制连接... \n", ccPort)
	exitChan := make(chan bool)

	for {
		ccConn, err := ccListener.Accept()
		if err != nil {
			log.Println(err)
			if ccConn != nil {
				ccConn.Close()
			}
			continue
		}
		log.Println("control connection is come in.")

		go func(exitChan chan bool, conn *net.Conn) {
			data := make([]byte, 16)
			_, err := (*conn).Read(data)
			if err != nil && err == io.EOF {
				exitChan <- true
			}
		}(exitChan, &ccConn)

		isExit := false
		for {
			select {
			case <-exitChan:
				isExit = true
			case <-notifyChan:
				n, err := ccConn.Write([]byte("OK"))
				log.Println("send OK to tunnel client. n is:", n)
				if err != nil {
					log.Println(err)
					if err == io.EOF {
						log.Println("exit notify.")
						isExit = true
					}
				}
			}
			if isExit {
				log.Println("exit loop.")
				break
			}
		}
	}
}

func dd(connChan chan *net.Conn) {
	//监听客户端数据传输
	ddListener, err := net.Listen("tcp", fmt.Sprintf(":%d", ddPort))
	if err != nil {
		panic(err)
	}
	log.Printf("监听:%d端口, 等待client 数据传输连接... \n", ddPort)
	for {
		// 有Client来连接了
		ddConn, err := ddListener.Accept()
		if err != nil {
			panic(err)
		}
		log.Println(ddConn.RemoteAddr(), " 新连接到来.")
		select {
		case ssConn := <-connChan:
			log.Println(ddConn.RemoteAddr(), " 开始交换数据.")
			go send(&ddConn, ssConn)
			go revice(&ddConn, ssConn)
		}
	}
}

func send(tc *net.Conn, uc *net.Conn) {
	n, err := io.Copy(*tc, *uc)
	log.Println("send:", n, err)
	if (*tc) != nil {
		(*tc).Close()
		log.Println("send tc close.")
	}
}

func revice(tc *net.Conn, uc *net.Conn) {
	n, err := io.Copy(*uc, *tc)
	log.Println("revive:", n, err)
	if (*uc) != nil {
		(*uc).Close()
		log.Println("revice uc close.")
	}
}

func start() {
	notifyChan := make(chan bool, 1024)
	go cc(notifyChan)
	ssConnChan := make(chan *net.Conn, 1024)
	go dd(ssConnChan)

	// 监听User连接
	ssListener, err := net.Listen("tcp", fmt.Sprintf(":%d", ssPort))
	if err != nil {
		panic(err)
	}
	log.Printf("监听:%d端口, 等待用户连接.... \n", ssPort)

	for {
		ssConn, err := ssListener.Accept()
		if err != nil {
			log.Println(err)
			if ssConn != nil {
				ssConn.Close()
			}
			continue
		}
		log.Printf("新用户连接: %s \n", ssConn.RemoteAddr())
		notifyChan <- true
		ssConnChan <- &ssConn
	}
}

func main() {
	flag.Parse()
	flag.Usage()
	go start()
	interrupt := make(chan os.Signal)
	signal.Notify(interrupt, os.Interrupt)
	signal.Notify(interrupt, os.Kill)
	<-interrupt
	fmt.Println("'Ctrl + C' received, exiting...")
}
