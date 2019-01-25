package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"
)

type Player struct {
	uuid string
	rtsp string
	rtmp string
	pid  int
}

var pidMap sync.Map
var heartMap sync.Map
var shellPath string
var rtmpIp string
var rtmpPort string

/*
×  实时查看监控
*  input:  camera uuid
*  output: 0 正常 1 视频播放异常 2参数错误
*/
func playVideo(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	_uuid := r.Form.Get("uuid")
	_rtsp := r.Form.Get("rtsp")

	if _uuid == "" || _rtsp == "" {
		io.WriteString(w, "{\"status\":\"2\"}")
		return
	}
	_rtmp := "rtmp://" + rtmpIp + "/" + _uuid
	_playURL := "http://" + rtmpIp + ":" + rtmpPort + "/" + _uuid

	log.Println("uuid:", _uuid, " rtsp:", _rtsp, " rtmp:", _rtmp, " playURL:", _playURL)
	_, ok := heartMap.Load(_uuid)
	if ok {
		result := "{\"status\":\"1\",\"playURL\":\"" + _playURL + "\"}"
		io.WriteString(w, result)
		return
	}

	_player := Player{
		uuid: _uuid,
		rtsp: _rtsp,
		rtmp: _rtmp,
		pid:  0}
	pidMap.Store(_uuid, _player)
	heartMap.Store(_uuid, time.Now().Unix())
	go exeShell(_uuid, _rtsp, _rtmp)
	result := "{\"status\":\"0\",\"playURL\":\"" + _playURL + "\"}"
	io.WriteString(w, result)
}

/*
×  查看拉流状态+心跳
*  input:  camera uuid
*  output: 0 正常 1 视频播放异常 2参数错误
*/
func live(w http.ResponseWriter, r *http.Request) {
	uuid := r.Form.Get("uuid")
	if uuid == "" {
		io.WriteString(w, "{\"status\":\"2\"}")
	} else {
		_, ok := heartMap.Load(uuid)
		if ok {
			heartMap.Store(uuid, time.Now().Unix())
			io.WriteString(w, "{\"status\":\"0\"}")
		} else {
			io.WriteString(w, "{\"status\":\"1\"}")
		}
	}
}

func killPid(uuid string) {
	player, ok := pidMap.Load(uuid)
	_player := player.(Player)
	if ok {
		err := syscall.Kill(-_player.pid, syscall.SIGKILL)
		if err != nil {
			log.Println("kill subprocess fail", err)
		} else {
			log.Println("kill subprocess success, pid:", _player.pid)
		}
		syscall.Kill(_player.pid+1, syscall.SIGKILL)
		syscall.Kill(_player.pid+2, syscall.SIGKILL)
	} else {
		log.Println("kill subprocess fail, not found pid:", _player.pid)
	}
}

func exeShell(uuid string, rtsp string, rtmp string) {
	script := shellPath + " '" + rtsp + "' '" + rtmp + "'"
	cmd := exec.Command("/bin/sh", "-c", script)
        cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	
	if err := cmd.Start(); err != nil {
		log.Println(err.Error())
	}

	pid := cmd.Process.Pid
	log.Println("cmd.Process.Pid:", pid)

	player, ok := pidMap.Load(uuid)
	_player := player.(Player)
	if ok {
		_player.pid = pid
		pidMap.Store(uuid, _player)
	}
	if err := cmd.Wait(); err != nil {
		log.Println(err.Error())
	}
	clean(uuid)

}

func clean(uuid string) {
	log.Println("clean map,the uuid is:", uuid)
	pidMap.Delete(uuid)
	heartMap.Delete(uuid)
}

func checkTimeOut() {

	f := func(k, v interface{}) bool {
		nowtime := time.Now().Unix()
		lasttime := v.(int64)
		uuid := k.(string)
		diff := nowtime - lasttime
		fmt.Println(uuid, "the time diff:", diff, "second.")
		if diff > 60 {
			killPid(uuid)
			clean(uuid)
		}
		return true
	}

	for {
		heartMap.Range(f)
		time.Sleep(15 * time.Second)
	}
}

func main() {
	arg_num := len(os.Args)
	if arg_num != 5 {
		log.Println("usage: nohup ./you_server_name port shell_path rtmp_ip rtmp_port &")
		log.Println("port:", "listen on port")
		log.Println("shell_path:", "shell path like: /x/x/ffmpeg_run.sh")
		log.Println("rtmp_ip:", "rtmp server ip")
		log.Println("rtmp_port:", "rtmp server port")
		os.Exit(0)
	}
	shellPath = os.Args[2]
	rtmpIp = os.Args[3]
	rtmpPort = os.Args[4]

	go checkTimeOut()

	log.Println("Rtsp pull and rtmp push server is run , listen on:" + os.Args[1])
	log.Println("the shell path is: ", shellPath)
	log.Println("the rtmp ip and port is: ", rtmpIp, rtmpPort)

	http.HandleFunc("/video", playVideo)
	http.HandleFunc("/live", live)
	err := http.ListenAndServe(":"+os.Args[1], nil)
	if err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
