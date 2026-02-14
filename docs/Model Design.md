
# 模型设计

---
+ 创建时间: 2026.02.11
---

## 用户 User

用户结构 {
    UUID        string      用户唯一标识
    UserIds     []string    用户多个TG号
    Premium     string      高级用户标识
    Language    string      用户消息语言（zh，en）
}

用户映射结构 {
    UserId      string      用户TG号userId
    UUID        string      用户唯一标识
}

## 权限 Permissions

权限结构 {
    UUID                    string      权限唯一标识
    Type                    string      权限类型
    AsyncDownloadQuantity   number      并发下载数量（-1:无限制）
    AsyncDownloadRemain     number      并发下载剩余数量
    DailyDownloadQuantity   number      每日下载数量（-1:无限制）
    DaliyDownloadRemain     number      每日下载剩余数量
    DaliyDownloadDate       timestamp   每日下载时间戳（记录哪天剩余数量，单位：秒10位）
    FileDownloadSize        number      下载文件大小限制（单位：byte）
}

基础权限：
    UUID                    "xxxxxxxxx"
    AsyncDownloadQuantity   1
    AsyncDownloadRemain     1
    DailyDownloadQuantity   10
    DaliyDownloadRemain     10
    DaliyDownloadDate       1770881636
    FileDownloadSize        1024 * 1024 * 1024 * 1.5

高级权限：
    UUID                    ""
    AsyncDownloadQuantity   3
    AsyncDownloadRemain     1
    DailyDownloadQuantity   100
    DaliyDownloadRemain     10
    DaliyDownloadDate       1770881636
    FileDownloadSize        1024 * 1024 * 1024 * 10

// 等待确定更高权限

## torrent 

torrent {
    infoHash    string      
    
}