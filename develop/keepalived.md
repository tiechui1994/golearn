## keepalived

VRRP 是Virtual Router Redundancy Protocol(虚拟路由冗余协议)的缩写, VRRP协议将两台或多台路由器设备虚拟成一个设
备, 对外提供虚拟路由器IP(一个或多个).

双机高可用的两种模式:

- 主从模式: 一台主服务器, 一台热备服务器, 正常情况下, 主服务器绑定公网虚拟IP, 提供负载均衡服务, 热备服务器处于空闲状
态; 当主服务器发生故障时, 热备服务器j接管主服务器的公网虚拟IP, 提供负载均衡服务.


- 主主模式: 两台负载均衡服务器, 互为主备, 且都处于活动状态, 同时各自绑定一个公网虚拟IP, 提供负载均衡服务; 当其中一
台发生故障时, 另外一台接管发生故障服务服务器的公网虚拟IP(此时活动的机器承担所有请求)


### 主从配置

vip: 192.168.200.100

// master config
```
// master
global_defs {
    router_id master       
}

# check nginx runing
vrrp_script chk_nginx {
    script "/etc/keepalived/nginx_check.sh"
    interval 2
    weight -20
}

vrrp_instance VI_1 {
    state MASTER   
    interface eth0
    virtual_router_id 50
    nopreempt
    priority 100
    advert_int 1
    virtual_ipaddress {
        192.168.200.100
    }
    track_script {
        chk_nginx
    }
}
```

// backup config
```
// backup
global_defs {
    router_id backup       
}

# check nginx runing
vrrp_script chk_nginx {
    script "/etc/keepalived/nginx_check.sh"
    interval 2
    weight -20
}

vrrp_instance VI_1 {
    state BACKUP   
    interface eth0
    virtual_router_id 50
    nopreempt
    priority 80
    advert_int 1
    virtual_ipaddress {
        192.168.200.100
    }
    track_script {
        chk_nginx
    }
}
```

### 主从配置升级到双主配置

vip1: 192.168.200.100
vip2: 192.168.200.200

// one
```
// master
global_defs {
    router_id master       
}

# check nginx runing
vrrp_script chk_nginx {
    script "/etc/keepalived/nginx_check.sh"
    interval 2
    weight -20
}

vrrp_instance VI_1 {
    state MASTER   
    interface eth0
    virtual_router_id 50
    nopreempt
    priority 100
    advert_int 1
    virtual_ipaddress {
        192.168.200.100
    }
    track_script {
        chk_nginx
    }
}

vrrp_instance VI_2 {
    state BACKUP   
    interface eth0
    virtual_router_id 51
    nopreempt
    priority 80
    advert_int 1
    virtual_ipaddress {
        192.168.200.200
    }
    track_script {
        chk_nginx
    }
}
```

// another
```
// backup
global_defs {
    router_id backup       
}

# check nginx runing
vrrp_script chk_nginx {
    script "/etc/keepalived/nginx_check.sh"
    interval 2
    weight -20
}

vrrp_instance VI_1 {
    state BACKUP   
    interface eth0
    virtual_router_id 50
    nopreempt
    priority 80
    advert_int 1
    virtual_ipaddress {
        192.168.200.100
    }
    track_script {
        chk_nginx
    }
}

vrrp_instance VI_2 {
    state MASTER
    interface eth0
    virtual_router_id 51
    nopreempt
    priority 100
    advert_int 1
    virtual_ipaddress {
        192.168.200.200
    }
    track_script {
        chk_nginx
    }
}
```