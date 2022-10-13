package main

import (
	"fmt"
	"os"
	"net"
	"time"
	"net/http"
	netUrl "net/url"
	"encoding/base64"
	"sync"
	"io/ioutil"
	"io"
	"errors"
	"strings"
	"path/filepath"

	"github.com/farsightsec/golang-framestream"
	"github.com/pelletier/go-toml/v2"
)

var wg sync.WaitGroup
var wgHTTP sync.WaitGroup

type URL struct {
	Client *http.Client
	Path string
}

var url URL

type ConfigFile struct {
	User string `json:"user"`
	Password string `json:"password"`
	Socket string `json:"socket"`
	Destination string `json:"destination"`
	SrcIP string `json:"srcip"`
	Retry_Delay time.Duration `json:"retry_delay"`
	Flush_MS time.Duration `json:"flush_ms"`
}

func createENV() {
	data, err := ioutil.ReadFile(ConfigName)
	if err != nil {
		fmt.Println("Config file error:", err)
		fmt.Println(VERSION)
		os.Exit(1)
	}
	Config = ConfigFile{}
	toml.Unmarshal(data, &Config)

	if Config.User == "" {
		fmt.Println("No user set!")
		fmt.Println(VERSION)
		os.Exit(1)
	}

	if Config.Password == "" {
		fmt.Println("No password set!")
		fmt.Println(VERSION)
		os.Exit(1)
	}

	if Config.Destination == "" {
		fmt.Println("No destination set!")
		fmt.Println(VERSION)
		os.Exit(1)
	}

	if Config.Socket == "" {
		err = os.MkdirAll("/etc/dnstap", 0755)
		if err != nil {
			fmt.Println("Could not create socket directory:", err)
			os.Exit(1)
		}
		Config.Socket = "/etc/dnstap/dnstap.sock"
	} else {
		dir := filepath.Dir(Config.Socket)
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			fmt.Println("Could not create socket directory:", err)
			os.Exit(1)
		}
	}
	if Config.SrcIP != "" {
                ip := net.ParseIP(Config.SrcIP)
                if ip == nil {
                        fmt.Println("Invalid srcip set!")
                        fmt.Println(VERSION)
                        os.Exit(1)
                }
	}

	Config.Retry_Delay = time.Duration(5) * time.Second

	Config.Flush_MS = time.Duration(500) * time.Millisecond
}

func handlepanic() {
	if a := recover(); a != nil {
		go checkSocket()
		run()
	}
}

func checkSocket() {
	defer handlepanic()
	for {
		time.Sleep(5 * time.Second)
		socket := Config.Socket
		_, err := os.Stat(socket)
		if err != nil {
			panic("Socket Missing!")
		}
	}
}

func run() {
	fname := Config.Socket
	os.Remove(fname)

	socket, err := net.Listen("unix", fname)
	if err != nil {
		fmt.Println(err, "Could not create socket!")
		fmt.Println(VERSION)
		os.Exit(1)
	}
	defer socket.Close()
	_ = os.Chmod(fname, 0777)

	for {
		conn, err := socket.Accept()
		if err != nil {
			time.Sleep(Config.Retry_Delay)
			continue
		}
		wg.Add(1)
		go handleConn(conn)
	}
	wg.Wait()
}

func handleConn(conn net.Conn) {
	defer wg.Done()
	defer conn.Close()

	packets := make(chan []string, 50)
	wgHTTP.Add(1)
	go httpSender(packets)

	var FSContentType = []byte("protobuf:dnstap.Dnstap")
	bi := true
	timeout := time.Second

	readerOptions := framestream.ReaderOptions{
		ContentTypes:	[][]byte{FSContentType},
		Bidirectional:	bi,
		Timeout:	timeout,
	}
	fs, err := framestream.NewReader(conn, &readerOptions)
	if err != nil {
		return
	}

	buf := make([]byte, BUFFER_SIZE * KILOBYTE)

	var count int
	var buffer []string
	ticker := time.NewTicker(Config.Flush_MS)
	defer ticker.Stop()
mainLoop:
	for {
		select {
		case <- ticker.C:
			if len(buffer) != 0 {
				packets <- buffer
				count = 0
				buffer = []string{}
			}
		default:
			n, err := fs.ReadFrame(buf)
			if err == framestream.EOF {
				break mainLoop
			}
			if err != nil {
				continue
			}
			frame := base64.StdEncoding.EncodeToString(buf[:n])
			count = count + len(frame)
			buffer = append(buffer, frame)
			if count >= BUFFER_SIZE * KILOBYTE * 10 {
				packets <- buffer
				count = 0
				buffer = []string{}
				ticker.Reset(Config.Flush_MS)
			}
		}
	}
	close(packets)
	wgHTTP.Wait()
}

func httpSender(packets <-chan []string) {
	defer wgHTTP.Done()
	dial()
	for packet := range packets {
		err := call(packet)
		if err != nil {
			time.Sleep(Config.Retry_Delay)
			dial()
		}
	}
}

func dial() {
	for {
		hosts, err := getHosts()
		if err != nil {
			time.Sleep(Config.Retry_Delay)
			continue
		}
		uri, err := tryConnect(hosts)
		if err != nil {
			time.Sleep(Config.Retry_Delay)
			continue
		}
		tr := getTransport()
		url.Client = &http.Client{Transport: &tr}
		url.Path = uri
		return
	}
}

func tryConnect(hosts []string) (string, error) {
	for _, collector := range hosts {
		uri := fmt.Sprintf("https://%s:%s@%s/dnstap/receiver",
			Config.User,
			Config.Password,
			collector,
		)
		tr := getTransport()
		client := &http.Client{Transport: &tr}
		resp, err := client.Get(uri)
		if err != nil {
			continue
		}
		io.Copy(ioutil.Discard, resp.Body)
		resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			return uri, nil
		}
	}
	return "", errors.New("Could not connect to any hosts")
}

func getTransport() (http.Transport) {
	if Config.SrcIP == "" {
		tr := http.Transport{
			MaxIdleConnsPerHost: 1024,
			TLSHandshakeTimeout: 0 * time.Second,
		}
		return tr
	} else {
		ip := net.ParseIP(Config.SrcIP)
		if ip == nil {
			fmt.Println("Invalid srcip set!")
			fmt.Println(VERSION)
			os.Exit(1)
		}
		addr := &net.TCPAddr{ip,0,""}
		tr := http.Transport{
			MaxIdleConnsPerHost: 1024,
			TLSHandshakeTimeout: 0 * time.Second,
			DialContext: (&net.Dialer{
				LocalAddr: addr,
			}).DialContext,
		}
		return tr

	}

}

func getHosts() ([]string, error) {
	var hosts []string
	service := "https"
	destination := Config.Destination

	_, srv, err := net.LookupSRV(service, "tcp", destination)
	if err != nil {
		return hosts, err
	}
	for _, s := range srv {
		hosts = append(hosts, fmt.Sprintf("%s:%d", strings.TrimSuffix(s.Target, "."), s.Port))
	}
	return hosts, nil

}

func call(packet []string) error {
	method := "POST"

	form := netUrl.Values{}
	for _, frame := range packet {
		form.Add("data", frame)
	}

	req, err := http.NewRequest(method, url.Path, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rsp, err := url.Client.Do(req)
	if err != nil {
		return err
	}
	io.Copy(ioutil.Discard, rsp.Body)
	defer rsp.Body.Close()
	if rsp.StatusCode != http.StatusOK {
		status := fmt.Sprintf("Request failed with response code: %d", rsp.StatusCode)
		return errors.New(status)
	}
	return nil
}
