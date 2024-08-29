# pairec
A Go web framework for quickly building recommendation online services based on JSON configuration.

[中文技术文档](https://help.aliyun.com/zh/airec/basic-introduction-1?spm=a2c4g.11186623.0.0.3a8c3672NtpB9B)

## Install
```bash
go get github.com/alibaba/pairec/v2
```

## Quick Start

You can use [pairecmd](https://help.aliyun.com/zh/airec/quickly-create-projects?spm=a2c4g.11186623.0.0.4aca340cgsOHu4) to quickly create project and start service .

From [here](https://help.aliyun.com/zh/airec/engine-configuration-doc/?spm=a2c4g.11186623.0.0.5d353672En0nlQ), you can find a lot of useful configuration information.

PAIREC  comes with a variety of built-in model functionalities, making it easy and fast to build recommendation services.

![yuque_diagram](http://pai-vision-data-hz.oss-cn-zhangjiakou.aliyuncs.com/pairec/docs/pairec/html/_images/yuque_diagram.png)

## Introduction

### Overall architecture

![framework](http://pai-vision-data-hz.oss-cn-zhangjiakou.aliyuncs.com/pairec/docs/pairec/html/_images/framework.jpg)

When you use aliyun to deploy recommend service . The following diagram illustrates the overall deployment architecture.

![image-20230727192436463](http://pai-vision-data-hz.oss-cn-zhangjiakou.aliyuncs.com/pairec/docs/pairec/html/_images/image-20230727192436463.png)
