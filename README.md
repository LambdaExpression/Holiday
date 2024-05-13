# 节假日服务

## 1.介绍

本项目主要为了提供节假日api接口，用于接入 Home Assistant 或 第三方开发使用。

通过请求懒加载的方式从 http://timor.tech/api/holiday/year/${year} 获取当年的节假日信息，存储到项目本地的sqlite数据库文件，已提供api接口查询。

## 2.启动说明

1. 下载应用 `wget https://github.com/LambdaExpression/Holiday/releases/download/v0.1/Holiday_amd64`
2. 赋予执行权限 `chmod +x ./Holiday_amd64`
3. 运行 `./Holiday_amd64 -prot 8080 -path "/data/Holiday/"`

## 3.接口文档


### 3.1 获取今天是否节假日
- **接口说明：** 获取今天是否节假日
- **接口地址：** /holiday/today

#### 3.1.1 请求参数

#### 3.1.2 返回结果

参数名称						|类型		|出现要求	|描述  
:----						|:---		|:------	|:---	
true / false							|string		|R			|直接返回文本 true / false 


示例：

```
true
```


### 3.2 获取昨天是否节假日
- **接口说明：** 获取昨天是否节假日
- **接口地址：** /holiday/yesterday

#### 3.2.1 请求参数

#### 3.2.2 返回结果

参数名称						|类型		|出现要求	|描述  
:----						|:---		|:------	|:---	
true / false							|string		|R			|直接返回文本 true / false 


示例：

```
true
```


### 3.3 获取明天是否节假日
- **接口说明：** 获取明天是否节假日
- **接口地址：** /holiday/tomorrow

#### 3.3.1 请求参数


请求示例：

```
curl http://127.0.0.1:8080/holiday/tomorrow
```

#### 3.3.2 返回结果

参数名称						|类型		|出现要求	|描述  
:----						|:---		|:------	|:---	
true / false							|string		|R			|直接返回文本 true / false 


示例：

```
true
```


### 3.4 获取某天是否节假日
- **接口说明：** 获取某天是否节假日
- **接口地址：** /info/{date}

#### 3.4.1 请求参数

参数名称						|类型		|出现要求	|描述  
:----						|:---		|:------	|:---	
date						|string	|R			|"2024-05-01"格式的日期


请求示例：

```
curl http://127.0.0.1:8080/info/2024-04-28
```


#### 3.4.2 返回结果

参数名称						| 类型		     |出现要求	|描述  
:----						|:---------|:------	|:---	
code						| int		    |R			|响应码 0 成功，-1 异常
msg							| string		 |R			|
data						| object		 |R			|
&emsp;date				    | string		 |R			|时间
&emsp;holiday				| bool		   |R			|是否假期
&emsp;name				    | string		 |R			|名称
&emsp;type					| int		    |R			|类型 0 通过是否周六日判断, 1 通过[接口](http://timor.tech/api/holiday/year/2024)来源判断是否节假日, 3 锁定状态(规划是后期手动修改数据，type变更完为3，数据锁定)


示例：

```
{
	"code": 0,
	"msg": "",
	"data": {
		"date": "2024-04-28",
		"holiday": false,
		"name": "劳动节前补班",
		"type": 1
	}
}
```



### 3.5 刷新某年的节假日数据
- **接口说明：** 刷新某年的节假日数据
- **接口地址：** /update/{year}

#### 3.5.1 请求参数

参数名称						| 类型		 |出现要求	|描述  
:----						|:-----|:------	|:---	
year						| int	 |R			| 具体要刷新的某年份


请求示例：

```
curl http://127.0.0.1:8080/update/2024
```


#### 3.5.2 返回结果

参数名称						| 类型		     |出现要求	|描述  
:----						|:---------|:------	|:---	
code						| int		    |R			|响应码 0 成功，-1 异常
msg							| string		 |R			|
data						| object		 |R			|

示例：

```
{
	"code": 0,
	"msg": "success",
	"data": null
}
```

