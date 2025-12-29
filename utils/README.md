# 工具函数（utils/）

提供丰富的常用工具函数，涵盖加密解密、IP处理、时间操作、随机数生成、数据脱敏、类型判断等功能：

### 加密解密（encryption.go）
- **AES-CBC 加密解密**：支持 AES-CBC 模式的加密解密，包含安全包装函数防止 panic
- **MD5 哈希**：`Md5X()` 计算字符串的 MD5 值
- **HMAC-SHA1**：`HMACSHA1()` 生成 HMAC-SHA1 签名

### IP 地址处理（ip.go）
- **IP 白名单检查**：`IsBan()` 支持具体 IP 和 CIDR 格式的白名单验证
- **本地 IP 获取**：`GetLocalIP()` 获取本地优先级最高的非回环 IP 地址，支持 IPv4/IPv6
- **私有 IP 判断**：`IsPrivateIP()` 判断是否为私有 IP 地址

### 时间处理（time.go）
- **日期格式验证**：`IsDate()` 检查日期格式是否正确
- **时间获取**：`Today()` 获取当前日期，`Now()` 获取当前日期时间，`NowTimeStamp()` 获取当前时间戳
- **时间转换**：`TimestampToDate()` 时间戳转日期，`DatetimeToTime()` 字符串转时间对象
- **时间计算**：`TimeDifference()` 计算两个时间的时间差，`GetDayTimeRange()` 获取一天的开始和结束时间

### 随机数生成（rand.go）
- **随机数生成**：`RandomNum()` 生成指定范围的随机整数
- **随机元素**：`RandomElement()` 从 map 中随机选择元素
- **随机字符串**：`RandomString()` 生成指定长度的随机字符串（字母数字组合）
- **随机验证码**：`RandomCode()` 生成指定长度的随机数字验证码

### 类型判断与验证（type.go）
- **字符串验证**：`IsOnlyChinese()` 判断是否只包含中文，`IsOnlyNumber()` 判断是否只包含数字
- **零值判断**：`IsZeroValue()` 判断变量是否为零值
- **格式验证**：`IsChinesePhoneNumber()` 中国手机号验证，`IsEmail()` 邮箱格式验证
- **字符串处理**：`EncodeFileName()` 文件名编码，`CleanInput()` 清理输入字符串，`CleanNewline()` 清理换行符

### 数据脱敏（tools.go）
- **手机号脱敏**：`MaskPhoneNumber()` 智能手机号码脱敏，保留前3位和后4位，中间用*代替

### 系统操作（os.go）
- **目录操作**：`EnsureDir()` 确保目录存在，不存在则创建
