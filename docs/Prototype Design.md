
# 原型设计

---
+ 创建时间: 2026.02.11
---


## Bot 命令

设置命令

```txt
start - Start your torrent download bot 
magnet - Analyzing magnet links
self - Personal Information
help - Contact for help
```

## 命令消息

### start 命令消息
```txt
Hi,  {username}   
欢迎使用 新币搜 xbso 🔍 中文搜索 

🔍 功能介绍：
- 解析 magnet 链接
- 下载出的解析文件

⌨️ 使用方式：
直接发送 magent 即可开始解析
如：magnet:?xt=urn:btih:E7FC73D9E20697C6C440203F5884EF52F9E4BD28

 免责声明：
- 只提供解析下载功能，下载内容与本Bot无关
- 不存储内容，不提供下载，请自行判断内容真实性与合规性

Bot频道：
下载文件频道：@tgqpXOZ2tzXN
帮助反馈频道：@bt1bot1channel
```

### magnet 命令消息
1. 开始解析消息
```txt
⌛ 正在解析磁力链接，请稍等...
🔗 Magent: {magnet}

[停止解析]
```

2. 解析成功消息
```txt
✅ 磁力链接解析成功
📛 名称: 【高清影视之家发布 www.WHATMV.com】千与千寻[中文字幕].Spirited.Away.2001.Repack.1080p.BluRay.x265.10bit.DTS-QuickIO
🔑 Info Hash: a44d70d77038e7f6bfaa8bd0d0270b246cd2812d
📦 总大小: 6.66 GB
🧩 分片数: 853
📏 分片大小: 8.00 MB
🔗 Magent: {magnet}

📁 文件列表 (4 个文件):
  1.Spirited.Away.2001.Repack.1080p.BluRay.x265.10bit.DTS-QuickIO.mkv (6.66 GB)
  2.【更多无水印蓝光原盘请访问 www.BBQDDQ.com】【更多无水印蓝光原盘请访问 www.BBQDDQ.com】.MP4 (289.14 KB)
  3.【更多无水印蓝光电影请访问 www.BBQDDQ.com】【更多无水印蓝光电影请访问 www.BBQDDQ.com】.DOC (289.14 KB)
  4.【更多无水印高清电影请访问 www.BBQDDQ.com】【更多无水印高清电影请访问 www.BBQDDQ.com】.MKV (622.05 KB)

[全部][1][2][3]
```

3. 解析失败消息

3.1 链接格式错误
```txt
❎ 解析失败
⏱ 等待时长: 
ℹ 错误信息: 链接格式错误
🔗 Magent: {magnet}

正确格式，例如：magnet:?xt=urn:btih:E7FC73D9E20697C6C440203F5884EF52F9E4BD28
```

3.2 解析超时失败
```txt
❎ 解析失败
⏱ 等待时长: 3m0s
ℹ 错误信息: 获取磁力链接元信息超时。
🔗 Magent: {magnet}

可能的原因：
• 网络连接问题
• 磁力链接无效
• 超时（3分钟）
```

3.3 取消解析
```txt
❎ 解析失败
⏱ 等待时长: 3m0s
ℹ 错误信息: 解析链接已经取消。
🔗 Magent: {magnet}
```

4. 下载文件消息

4.1 文件下载中
```txt
⌛ 文件下载中...
🔗 Magent: {magnet}
💾 正在下载文件：
{10%} 文件名

[取消下载]
```

4.2 文件下载成功

正常下载完成
```txt
✅ 文件下载成功
🔑 Info Hash: #
💾 下载文件：


前往消息频道：@tgqpXOZ2tzXN
```

缓存命中
```txt
✅ 文件下载成功
🔑 Info Hash: #
💾 下载文件：

前往消息频道：@tgqpXOZ2tzXN
```


4.3 文件下载失败

并发数限制
```txt
❎ 下载失败
⏱ 等待时长: 3m0s
ℹ 错误信息: 获取磁力链接元信息超时。
🔗 Magent: {magnet}
💾 下载文件：

```

超时错误
```txt
❎ 下载失败
⏱ 等待时长: 3m0s
ℹ 错误信息: 获取磁力链接元信息超时。
🔗 Magent: {magnet}
💾 下载文件：

```

取消下载
```txt
❎ 下载失败
⏱ 等待时长: 3m0s
ℹ 错误信息: 获取磁力链接元信息超时。
🔗 Magent: {magnet}
💾 下载文件：

```

### self 命令消息

```txt
个人消息
```

### help 命令消息

```txt
帮助消息
```


## 命令处理逻辑

### Start 处理逻辑

1. 解析消息出其中的 UserId
2. 用户映射表寻找 UserId 对应的 UUID，不存在顺序执行步骤3，存在则不处理。
3. 创建一个新的 User 和 一个唯一 UUID，创建一个基础的权限赋予该用户，修改用户映射表。
4. 返回开始消息【命令消息 start 命令消息】

### magnet 处理逻辑

magnet 解析逻辑：
1. 解析命令消息，获取 UserId ，和 消息Id 记录在一个全局 Map 中
2. 获取消息中的 magent 链接，校验其格式是否正确。不正常返回格式错误错误信息。
3. 设置解析超时时间：3分钟，通过 torrent 客户端进行解析。
4. 解析失败，返回解析失败消息。
5. 解析成功，记录解析结果到数据库中并返回解析数据。
6. 解析出来的数据需要显示文件列表信息，给出文件列表下载按钮，全部，1～x。
7. 用户点击某个文件进行下载

下载文件逻辑：
1. 解析按钮消息，获取 UserId ，和 消息Id 记录在一个全局 Map 中
2. 解析出需要被下载的文件。判断文件是否已经下载过了，下载过了，在数据库中寻找文件，直接返回下载过的消息
3. 判断当前用户是否正在下载其他文件，校验用户剩余并发下载数。
4. 没有剩余次数返回限制信息。有剩余数量进行下载操作
5. 设置超时时间，通过 torrent 客户端进行下载。实时更新下载消息
6. 超时返回超时消息
7. 下载完成，通过接入的 TG帐号 发送到消息频道。并记录到数据库文件存在和对应的消息
8. 返回下载成功的消息，返回用户

### self 处理逻辑

1. 解析命令消息，获取 UserId ，和 消息Id 记录在一个全局 Map 中
2. 查询用户权限信息并返回用户

### help 命令消息

1. 返回 help 命令消息

## 其他逻辑

### 消息发送帐号

1. 通过一个 tglogin 程序登录帐号，存储一个 session 在本地。
2. BT-BOT 程序，通过 gotd/td 库读取 session 文件，登录该帐号
3. 文件消息通过该帐号发送

