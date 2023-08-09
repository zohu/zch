### zch
Two-level cache (memory + redis)

```
github.com/zohu/zch
```
```
// 初始化
type Options struct {
	Expiration    time.Duration             // 二级缓存的超时时间
	CleanInterval time.Duration             // 二级缓存的刷新时间
	PrefixL2      string                    // 二级缓存的前缀
	RedisOptions  *redis.UniversalOptions   // redis的配置
}
zch.NewL2(Options)
```
```
zch.L().XXX  带二级的缓存 
zch.C().XXX  纯内存
zch.R().XXX  纯redis
```

### CHANGELOG:
#### 2023-08-10
- L2缓存增加Flush方法，可实现一键释放L2的所有缓存，不影响正常Redis其他缓存