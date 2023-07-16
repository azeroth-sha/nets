<p align="center">
<b>Go 语言的轻量网络库</b>
<br/>
<a title="Go Report Card" target="_blank" href="https://goreportcard.com/report/github.com/azeroth-sha/nets"><img src="https://goreportcard.com/badge/github.com/azeroth-sha/nets?style=flat-square" /></a>
<a title="Release" target="_blank" href="https://github.com/azeroth-sha/nets/releases"><img src="https://img.shields.io/github/v/release/azeroth-sha/nets.svg?color=161823&style=flat-square&logo=smartthings" /></a>
<a title="Tag" target="_blank" href="https://github.com/azeroth-sha/nets/tags"><img src="https://img.shields.io/github/v/tag/azeroth-sha/nets?color=%23ff8936&logo=fitbit&style=flat-square" /></a>
<a title="Doc for nets" target="_blank" href="https://pkg.go.dev/github.com/azeroth-sha/nets?tab=doc"><img src="https://img.shields.io/badge/go.dev-doc-007d9c?style=flat-square&logo=read-the-docs" /></a>
</p>

## 📖 简介

`nets`是一个高性能、轻量级的Go标准库[net.Conn](https://pkg.go.dev/net#Conn)封装，每个conn仅启用一个goroutine，资源复用，达到更优的任务效果。

特别说明: 设计灵感来自[Gnet](https://github.com/panjf2000/gnet)（包括本文档 ^_^），不依赖第三方库。

## 🚀 功能：

- 非阻塞的异步网络工具库
- 使用sync.Pool管理buff资源，达到复用目的并合理自旋。
- 提供简约而不简单的连接管理
- 优雅处理连接panic，防止程序崩溃（未处理OnBoot/OnShutdown/OnTick, 个人认为在服务启动、停止时进行回收错误没有任何意义并可能照成无法预估的后果。 欢迎讨论）
- 连接事务非阻塞机制，优雅的处理多重关闭事件，杜绝重复关闭。

## 🛠 使用

- [客户端](examples/client/main.go)
- [服务端](examples/server/main.go)

## 📄 证书

`nets` 的源码允许用户在遵循 [MIT 开源证书](./LICENSE) 规则的前提下使用。
