## keepalived

双机高可用的两种模式:

- 主从模式: 一台主服务器, 一台热备服务器, 正常情况下, 主服务器绑定公网虚拟IP, 提供负载均衡服务, 热备服务器处于空闲状
态; 当主服务器发生故障时, 热备服务器j接管主服务器的公网虚拟IP, 提供负载均衡服务.


- 主主模式: 两台负载均衡服务器, 互为主备, 且都处于活动状态, 同时各自绑定一个公网虚拟IP, 提供负载均衡服务; 当其中一
台发生故障时, 另外一台接管发生故障服务服务器的公网虚拟IP(此时活动的机器承担所有请求)


主从:

```
// master
global_defs {
    router_id master       
}

vrrp_instance VI_1 {
    state MASTER   
    interface eth0
    virtual_router_id 50
    nopreempt
    priority 100
    advert_int 1
    virtual_ipaddress {
        192.168.200.11
    }
}


// backup
global_defs {
    router_id backup       
}

vrrp_instance VI_1 {
    state BACKUP   
    interface eth0
    virtual_router_id 50
    nopreempt
    priority 80
    advert_int 1
    virtual_ipaddress {
        192.168.200.11
    }
}
```