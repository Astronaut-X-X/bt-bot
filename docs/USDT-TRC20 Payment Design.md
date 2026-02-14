# USDT-TRC20 收款并检测收款成功的原理与实现

---

+ 创建时间: 2026.02.11
---

## 一、基本原理

### 1.1 USDT-TRC20 概述

USDT-TRC20 是基于波场（TRON）区块链网络的 USDT 代币标准。与以太坊的 ERC-20 类似，TRC-20 是波场网络上的代币标准。

**关键特点：**
- 交易速度快（3秒出块）
- 手续费低（通常只需 TRX 作为燃料费）
- 使用 TRON 地址（以 T 开头）

### 1.2 收款地址生成

每个用户订单需要生成一个唯一的收款地址。有两种方案：

**方案一：使用主钱包地址 + 订单号标记**
- 所有用户向同一个主钱包地址转账
- 通过转账备注（Memo）字段区分不同订单
- 优点：简单，无需管理多个地址
- 缺点：需要解析 Memo 字段

**方案二：为每个订单生成唯一地址**
- 使用 HD 钱包（分层确定性钱包）生成子地址
- 每个订单对应一个唯一地址
- 优点：无需解析 Memo，地址即订单标识
- 缺点：需要管理更多地址

**推荐方案一**（更简单实用）

## 二、技术实现流程

### 2.1 整体流程图

```
用户发起支付请求
    ↓
系统生成订单（订单号 + 金额 + 收款地址 + Memo）
    ↓
返回支付信息给用户（地址、金额、Memo）
    ↓
用户使用钱包转账
    ↓
系统轮询/监听区块链交易
    ↓
检测到符合条件的交易
    ↓
验证交易（金额、Memo、确认数）
    ↓
更新订单状态为已支付
    ↓
通知用户支付成功
```

### 2.2 详细步骤说明

#### 步骤 1: 创建订单

当用户需要支付时：
1. 生成唯一订单号（UUID 或时间戳+随机数）
2. 记录订单信息：
   - 订单号
   - 用户 UUID
   - 支付金额（USDT）
   - 收款地址（主钱包地址）
   - Memo（订单号或订单号+用户ID）
   - 订单状态（待支付）
   - 创建时间
   - 过期时间（如 30 分钟）

#### 步骤 2: 返回支付信息

向用户返回：
```
收款地址: Txxxxxxxxxxxxxxxxxxxxxxxxxxxxx
支付金额: 10 USDT
备注信息: ORDER-20260211-123456
```

#### 步骤 3: 监听区块链交易

**方法一：轮询方式（推荐）**

定期查询主钱包地址的交易记录：

```go
// 伪代码示例
func CheckPayment(orderID string) {
    // 1. 通过 TRON API 查询主钱包地址的交易
    // 2. 筛选符合条件的交易：
    //    - 转入交易（toAddress = 主钱包地址）
    //    - 代币类型 = USDT-TRC20
    //    - Memo 字段匹配订单号
    //    - 金额匹配
    // 3. 检查交易确认数（建议至少 19 个确认）
    // 4. 更新订单状态
}
```

**方法二：WebSocket 实时监听**

使用 TRON 节点的 WebSocket 接口实时监听：
- 优点：实时性高
- 缺点：需要维护 WebSocket 连接，实现复杂

**推荐方法一**（轮询方式，每 10-30 秒检查一次）

#### 步骤 4: 交易验证

检测到交易后需要验证：

1. **金额验证**
   - 交易金额 >= 订单金额（允许小额误差）
   - 代币类型为 USDT-TRC20（合约地址：TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t）

2. **Memo 验证**
   - 交易的 Memo 字段必须匹配订单号
   - 防止误判其他用户的转账

3. **确认数验证**
   - TRON 网络建议至少 19 个确认（约 1 分钟）
   - 确保交易不会被回滚

4. **时间验证**
   - 交易时间应在订单创建时间之后
   - 交易时间应在订单过期时间之前

5. **重复验证**
   - 检查该交易哈希是否已处理过
   - 防止重复确认

#### 步骤 5: 更新订单状态

验证通过后：
1. 更新订单状态为"已支付"
2. 记录交易哈希（TxID）
3. 记录支付时间
4. 更新用户账户余额或订阅状态
5. 通知用户支付成功

## 三、技术实现要点

### 3.1 TRON API 选择

**方案一：使用 TronGrid API（推荐）**
- 官方提供的公共 API
- 免费使用，有速率限制
- 文档：https://www.trongrid.io/

**方案二：自建 TRON 节点**
- 需要同步完整区块链数据
- 资源消耗大
- 适合高频交易场景

**方案三：使用第三方服务**
- TronScan API
- 其他区块链浏览器 API

**推荐方案一**（TronGrid API）

### 3.2 关键 API 接口

#### 查询账户交易记录

```
GET https://api.trongrid.io/v1/accounts/{address}/transactions/trc20
```

参数：
- `address`: 钱包地址
- `limit`: 返回数量限制
- `only_confirmed`: 是否只返回已确认交易
- `only_to`: 是否只返回转入交易

#### 查询交易详情

```
GET https://api.trongrid.io/v1/transactions/{txid}
```

#### 查询账户 TRC-20 代币余额

```
GET https://api.trongrid.io/v1/accounts/{address}/tokens
```

### 3.3 USDT-TRC20 合约地址

```
USDT-TRC20 合约地址: TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t
```

### 3.4 数据结构设计

```go
// 订单结构
type PaymentOrder struct {
    OrderID      string    // 订单号
    UserUUID     string    // 用户 UUID
    Amount       float64   // 支付金额（USDT）
    ReceiveAddr  string    // 收款地址
    Memo         string    // 备注（订单号）
    Status       string    // 状态：pending, paid, expired, failed
    TxHash       string    // 交易哈希
    CreatedAt    time.Time // 创建时间
    ExpiredAt    time.Time // 过期时间
    PaidAt       time.Time // 支付时间
}

// 交易信息结构
type TRC20Transaction struct {
    TxID         string    // 交易哈希
    From         string    // 发送地址
    To           string    // 接收地址
    Value        string    // 金额（字符串，需要转换）
    TokenAddress string    // 代币合约地址
    Memo         string    // 备注
    BlockTime    int64     // 区块时间戳
    Confirmed    bool      // 是否已确认
    Confirmations int      // 确认数
}
```

## 四、安全注意事项

### 4.1 防止重复支付

- 使用交易哈希（TxID）作为唯一标识
- 在数据库中记录已处理的交易哈希
- 处理前检查是否已存在

### 4.2 防止金额篡改

- 服务端验证金额，不信任客户端
- 使用精确的金额比较（考虑浮点数精度问题）

### 4.3 防止时间攻击

- 设置订单过期时间
- 验证交易时间在有效范围内

### 4.4 防止 Memo 伪造

- 使用强随机性生成订单号
- Memo 验证必须严格匹配

### 4.5 API 限流

- TronGrid API 有速率限制
- 实现请求重试和退避策略
- 考虑使用多个 API Key 轮询

## 五、实现建议

### 5.1 轮询策略

```go
// 伪代码
func StartPaymentMonitor() {
    ticker := time.NewTicker(30 * time.Second) // 每 30 秒检查一次
    
    for range ticker.C {
        // 获取所有待支付订单
        pendingOrders := GetPendingOrders()
        
        for _, order := range pendingOrders {
            // 检查订单是否过期
            if time.Now().After(order.ExpiredAt) {
                UpdateOrderStatus(order.OrderID, "expired")
                continue
            }
            
            // 检查支付
            CheckOrderPayment(order)
        }
    }
}
```

### 5.2 错误处理

- API 请求失败：记录日志，下次继续检查
- 网络超时：实现重试机制
- 数据解析错误：记录详细错误信息

### 5.3 日志记录

记录关键操作：
- 订单创建
- 支付检测
- 支付确认
- 异常情况

## 六、测试建议

### 6.1 测试网测试

使用 TRON 测试网（Shasta）进行测试：
- 测试网 API: https://api.shasta.trongrid.io
- 获取测试网 TRX 和 USDT

### 6.2 测试场景

1. 正常支付流程
2. 金额不足
3. Memo 错误
4. 过期订单
5. 重复支付检测
6. 网络异常处理

## 七、参考资料

- TRON 官方文档: https://developers.tron.network/
- TronGrid API 文档: https://www.trongrid.io/
- USDT-TRC20 合约: TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t
- TRON 区块浏览器: https://tronscan.org/

---

## 附录：Go 语言实现示例（伪代码）

```go
package payment

import (
    "encoding/json"
    "fmt"
    "net/http"
    "time"
)

// 查询账户 TRC-20 交易
func GetTRC20Transactions(address string, limit int) ([]TRC20Transaction, error) {
    url := fmt.Sprintf(
        "https://api.trongrid.io/v1/accounts/%s/transactions/trc20?limit=%d&only_confirmed=true&only_to=true",
        address, limit,
    )
    
    resp, err := http.Get(url)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    var result struct {
        Data []TRC20Transaction `json:"data"`
    }
    
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, err
    }
    
    return result.Data, nil
}

// 检查订单支付
func CheckOrderPayment(order *PaymentOrder) error {
    // 1. 查询收款地址的交易
    transactions, err := GetTRC20Transactions(order.ReceiveAddr, 50)
    if err != nil {
        return err
    }
    
    // 2. 筛选符合条件的交易
    for _, tx := range transactions {
        // 检查代币类型（USDT-TRC20）
        if tx.TokenAddress != "TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t" {
            continue
        }
        
        // 检查 Memo
        if tx.Memo != order.Memo {
            continue
        }
        
        // 检查金额（需要将字符串转换为数值）
        amount := parseUSDTAmount(tx.Value)
        if amount < order.Amount {
            continue
        }
        
        // 检查确认数
        if tx.Confirmations < 19 {
            continue
        }
        
        // 检查时间范围
        txTime := time.Unix(tx.BlockTime/1000, 0)
        if txTime.Before(order.CreatedAt) || txTime.After(order.ExpiredAt) {
            continue
        }
        
        // 检查是否已处理
        if IsTxProcessed(tx.TxID) {
            continue
        }
        
        // 验证通过，更新订单
        return ConfirmPayment(order.OrderID, tx.TxID)
    }
    
    return nil
}
```

