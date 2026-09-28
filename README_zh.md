<div align="center">

<img src="gui/public/static/v2raya-icon.svg" width="96" alt="x-Ray VPN">

# x-Ray VPN

基于 v2rayA-resilient 的实验版本。

包含探测与 DNS 取消、重复探测、进程清理、资源下载、XHTTP 上传和 REALITY 启动验证的修复。

TLS 会话恢复和可变 POST 大小默认关闭。仅部分匹配 Chrome 133；不宣称与当前浏览器完全一致或普遍降低 CPU/内存使用。

[Experimental release notes](EXPERIMENTAL.md)

**基于自带 Xray 内核的 Web 客户端，在 Linux、Windows 与 macOS 上提供透明代理。**

[English](README.md) · 简体中文 · [Русский](README_ru.md)

[运行环境](#运行环境) • [安装](#安装) • [首次运行](#首次运行) • [透明代理](#透明代理) • [数据与升级](#数据与升级) • [支持](#支持)

</div>

x-Ray VPN 以服务形式运行，通过浏览器操作，可部署在本机、路由器或 NAS 上。它导入 VMess、VLESS、Shadowsocks、Trojan、Hysteria2、TUIC、[Juicity](https://github.com/juicity)、AnyTLS、WireGuard、SOCKS5 与 HTTP(S) 的订阅和分享链接。节点编成组后，由实测延迟最低的成员或固定的成员出站，分流由 RoutingA 规则决定。不支持 ShadowsocksR。

## 运行环境

| 组件 | 要求 |
| --- | --- |
| 内核 | 与 `v2raya` **版本相同**的 `v2raya_core`，放在同一目录或 `PATH` 中。下列软件包和安装程序都同时安装两者；手动安装时两者版本必须一致 |
| 规则数据 | `/usr/share/v2raya` 或 `/usr/local/share/v2raya` 中的 `geoip.dat` 与 `geosite.dat`。软件包自带；手动安装时首次启动从 GitHub 下载，下载失败则退出 |
| Linux 透明代理 | root 权限；`redirect` 与 `tproxy` 需要 `iptables` 或 `nftables`；`tun` 需要 `/dev/net/tun` 与 iproute2 的 `ip` 命令 |
| Windows | Windows 10 build 14393 及以上；`tun` 需要管理员权限；未提权启动时运行在 lite 模式，仍可设置当前用户的系统代理 |
| macOS | `tun` 需要 root 权限；以普通用户启动时运行在 lite 模式，改为提供系统代理 |
| 浏览器 | 当前版本的 Chrome、Edge、Firefox 或 Safari |

## 安装

实验版目标平台为 OpenWrt **24.10.4 ARM64**（`aarch64_cortex-a53`）。请使用[实验版本](https://github.com/wywywywycloud/v2rayA-resilient/releases/tag/experimental-2026.09.28.1)中配套的三个 IPK。参阅[可复现构建与打包说明](install/openwrt-experimental/README.md)及该版本的验证记录。服务与内核必须一起安装或回退。

上游发行包和之前签名的 Resilient 软件源属于其他构建；其安装测试矩阵不能证明本实验版本已经通过相同测试。

## 首次运行

服务监听 `0.0.0.0:2017`，部署在路由器或 NAS 上时，界面地址为 `http://<设备地址>:2017`。第一个注册的账户就是管理员，注册无需登录，因此在共享网络中先限制访问再注册，或者只在本机使用时以 `--address 127.0.0.1:2017` 启动。

打开界面并创建管理员。随后的引导流程依次导入订阅或分享链接、把节点加入组、选择路由规则；内核由仪表板的启动按钮或页面顶部的状态按钮启动。

内核运行后，把应用指向 SOCKS5 `127.0.0.1:20170` 或 HTTP `127.0.0.1:20171`（`20172` 对 HTTP 应用路由规则），或者在仪表板开启透明代理并从下表选择模式。组内有多个已连接成员时，使用实测延迟最低的成员；固定某个成员后，该组只经它出站。订阅可在订阅页手动更新，也可按设置中的间隔自动更新。

<img src="docs/images/screenshot.png" alt="仪表板，明暗两种主题" width="100%">

## 透明代理

| 平台 | 模式 |
| --- | --- |
| Linux | `redirect`、`tproxy`、`tun`；`--lite` 下为 GNOME 与 KDE 的系统代理 |
| macOS | `tun`；`--lite` 下为系统代理 |
| Windows | `tun`、系统代理 |

系统代理模式修改桌面的代理设置，只覆盖遵守该设置的应用；其他模式拦截流量。

`tun` 模式下，内核打开 TUN 设备并配置地址；自动路由（默认开启）时由 x-Ray VPN 安装路由和 DNS 设置，否则由用户自己的脚本完成。内核自己的连接、direct 出站与 DNS 模块的上游查询不会进入 TUN（Linux 通过套接字标记，Windows 与 macOS 通过绑定物理网卡）。到达 TUN 的 53 端口明文 DNS 由内核的 DNS 模块回答，加密 DNS 不被拦截。x-Ray VPN 与内核始终被排除，其他进程可在设置里按可执行文件名排除。更精确的直连路由与静态路由不经过 TUN。

已知限制：Windows 与 macOS 上直接向局域网 DNS 查询的应用仍会绕过 TUN。系统解析器不受影响：Windows 指向 TUN 网关，macOS 指向内核在 `127.0.0.1` 上的监听，后者要求 53 端口空闲。

## 路由规则

<img src="docs/images/routinga.png" alt="RoutingA 编辑器，明暗两种主题" width="100%">

RoutingA 支持列表和文本两种编辑模式。列表模式为每条规则提供表单；文本模式提供行号、着色和逐行检查。编辑器旁有语法参考，规则可导入导出为文本文件。

## 数据与升级

SQLite 数据库 `v2raya.db` 与生成的内核配置位于配置目录：Linux 与 macOS 为 `/etc/v2raya`，Windows 服务为 `%ProgramData%\SYSTEM\v2rayA`，`--lite` 为用户配置目录，也可用 `--config` 指定。systemd 与 OpenRC 单元把日志写到 `/var/log/v2raya/v2raya.log`，也可用 `--log-file` 或 `V2RAYA_LOG_FILE` 指定。程序在更新订阅、下载规则数据、测量延迟、解析 DNS、启动时和每周一次向 GitHub 检查新版本，以及 VMess 节点失败时向 `ntp.aliyun.com` 查询时间（用于区分时钟错误与节点故障）时发起网络请求。

升级前先停止服务并备份配置目录。从 2.4 以下版本升级时，首次启动会迁移 BoltDB 数据库，旧文件保留为 `bolt.db.bak`（该名称已存在时为 `bolt.db.bak.1` 及后续编号），账户需要重新注册。`v2raya` 与 `v2raya_core` 必须一起升级：仪表板显示内核版本，版本不一致时显示横幅提示。

## 支持

问题请到 [Discussions](https://github.com/wywywywycloud/v2rayA-resilient/discussions) 提问，缺陷请到 [Issues](https://github.com/wywywywycloud/v2rayA-resilient/issues) 报告。

不要将本项目用于不合法用途。

## 致谢

由 [@mzz2017](https://github.com/mzz2017) 创立。Material Design 3 界面、内核内建的 TUN 与 2.5 的服务端重构由 [@Zakkaus](https://github.com/Zakkaus) 完成。OpenRC 文件来自 [Gentoo 中文社区](https://gentoozh.org)。路由数据来自 [v2fly/domain-list-community](https://github.com/v2fly/domain-list-community) 与 [v2fly/geoip](https://github.com/v2fly/geoip)，GFWList 模式使用 [Loyalsoldier/v2ray-rules-dat](https://github.com/Loyalsoldier/v2ray-rules-dat)。透明代理规则参考了 [zfl9/ss-tproxy](https://github.com/zfl9/ss-tproxy) 与 [hq450/fancyss](https://github.com/hq450/fancyss)。

## 许可证

[AGPL-3.0-only](LICENSE)。内核是 [Xray-core](https://github.com/XTLS/Xray-core) 的分支，采用 MPL-2.0。
