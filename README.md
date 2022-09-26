### zch
Two-level cache (memory + redis)

```
github.com/zohu/zch
```
```
// 初始化的前两个入参：一级缓存的超时时间，一级缓存的刷新时间
zch.NewL2(time.Hour, 5*time.Minute, &redis.UniversalOptions{
Addrs:    zrdsConf.Host,
Password: zrdsConf.Password,
DB:       0,
})
```
```
zch.L().XXX  带二级的缓存 
zch.C().XXX  纯内存
zch.R().XXX  纯redis
```