package oss

import "github.com/zjutjh/WeJH-SDK/cube"

// Client 对象存储客户端
var Client *cube.Client

func init() {
	Client = cube.New(getConfig())
}
