package App

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
)

type Middleware func(Endpoint) Endpoint
//业务最终函数原型
type Endpoint func(ctx context.Context,request interface{}) (response interface{}, err error)

//怎么取参数
type EncodeRequestFunc func(*gin.Context) (interface{}, error)

//怎么处理业务结果
type DecodeResponseFunc func(*gin.Context, interface{}) error

func RegisterHandler(endpoint Endpoint,encodeFunc EncodeRequestFunc, decodeFunc DecodeResponseFunc) func(context *gin.Context){
	return func(context *gin.Context) {

		defer func() {
			if r:=recover();r!=nil{
				fmt.Fprintln(gin.DefaultWriter,fmt.Sprintf("fatal error:%s",r))
				context.JSON(500,gin.H{"error":fmt.Sprintf("fatal error:%s",r)})
				return
			}
		}()
		//参数:=怎么取参数(context)
		//业务结果,error:=业务最终函数(context,参数)
		//
		//
		//怎么处理业务结果(业务结果)
		req,err:=encodeFunc(context) //获取参数
		if err!=nil{
			context.JSON(400,gin.H{"error":"param error:"+err.Error()})
			return
		}
		rsp,err:=endpoint(context,req) //执行业务过程
		if err!=nil{
			fmt.Fprintln(gin.DefaultWriter,"response error:",err)
			context.JSON(400,gin.H{"error":"response error:"+err.Error()})
			return
		}
		err=decodeFunc(context,rsp) //处理 业务执行 结果
		if err!=nil{
			context.JSON(500,gin.H{"error":"server error:"+err.Error()})
			return
		}

	}
}