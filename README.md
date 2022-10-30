#  合约事件监听

监听链上事件，可以保存到数据库，可以推送到指定服务。

当前只支持EVM的链，ethereum、polygon、BSC......

示例代码可以查看examples文件夹。

## 说明

### Event

1. 通过`event := NewEvent(...)`，创建订阅
   1. 参数`SubscriptionConf`
      1. Alias：可以是任意字符串
      2. Contract：合约地址，可以多个
      3. ABIFile：智能合约的abi文件
         1. 可以直接用`erc20`/`erc721`/`erc1155`作为配置项的值，将使用默认自带的abi
         2. 可以自己修改abi中参数的名称，从而实现自定义收到的数据
      4. EventName：要监听的事件
         1. 如果为空，则表示监听合约的所有事件
         2. 不允许监听无法识别的事件
      5. Filter：要过滤的参数，`map[string]string`
         1. key就是abi事件的参数名
         2. value默认为hex字符串，要匹配的值
            1. 比如from:0x...，用于监听指定地址的转出事件
            2. 如果参数为int/uint的，也支持数字
            3. 比如要监听指定tokenID的NFT
         3. 如果本身abi参数有indexed修饰，则会在查询节点时，就增加该过滤
   2. `type EventCallback func(alias string, info map[string]interface{}) error`
      1. 回调函数，监听到的事件，将通过回调通知到业务模块
      2. alias就是配置中的Alias
      3. info携带了具体的事件内容
         1. 包含基础内容：`KAlias`/`KBlock`/`KTX`等信息
         2. 包含事件对应的具体信息，如`Transfer`：
            1. from/to/amount
2. 通过`event.Run()`执行查询指定范围区块的事件
   1. 结果将通过callback通知到业务模块
   2. 如果callback返回error，表示异常，将退出

### DBItem

1. 可以将数据存储到数据库，不同的alias存储在不同的表里
2. tx+logIndex创建了索引，所以不允许重复
3. 例子可以查看`examples/2.save`
