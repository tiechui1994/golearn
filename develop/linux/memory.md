# Linux 下的 mmp

### 共享文件映射

```cgo
#include <fcntl.h>
#include <stdio.h>
#include <string.h>
#include <sys/mman.h>
#include <sys/stat.h>
#include <unistd.h>

int main(int argc, char* argv[]) {
    int fd;
    struct stat sb;
    char* p;
    if ((fd=open(argv[1], O_RDWR)) < 0) {
        perror("open");
    }

    if ((fstat(fd, &sb)) == -1) {
        perror("fstat");
    }

    if ((p = (char*)mmap(NULL, sb.st_size, PROT_READ | PROT_WRITE, MAP_SHARED, fd, 0)) == (void*)-1) {
        perror("mmap");
    }

    // must be execute
    memset(p, 'c', sb.st_size);
    sleep(100);
    return 0;
}
```

当前系统内存使用情况:

> ![image](/images/develop_linux_mmp_share_1.bmp)


调用程序之后, 进行共享文件映射, 内存的状况:

> ![image](/images/develop_linux_mmp_share_2.bmp)

使用的内存是 `buffer/cache`

### 私有文件映射

```cgo
#include <fcntl.h>
#include <stdio.h>
#include <string.h>
#include <sys/mman.h>
#include <sys/stat.h>
#include <unistd.h>

int main(int argc, char* argv[]) {
    int fd;
    struct stat sb;
    char* p;

    if ((fd = open(argv[1], 0)) < 0) {
        perror("open");
    }

    if ((fstat(fd, &sb)) == -1) {
        perror("fstat");
    }

    if ((p = (char*)mmap(NULL, sb.st_size, PROT_READ|PROT_WRITE, MAP_PRIVATE, fd, 0)) == (void*)-1) {
        perror("mmap");
    }

    memset(p, 'c', sb.st_size);
    sleep(100);

    return 0;
}
```

当前系统内存使用情况:

> ![image](/images/develop_linux_mmp_private_1.png)


调用程序时, 进行私有文件映射, 内存的状况:

> ![image](/images/develop_linux_mmp_private_2.png)

调用前后 `used` 和 `.

### 私有匿名映射

```cgo
#include <stdio.h>
#include <string.h>
#include <sys/mman.h>
#include <unistd.h>
#define SIZE 1024*1024*1024

int main(int argc, char* argv[]) {
    char* p;
    if ((p = (char*)mmap(NULL, SIZE, PROT_READ|PROT_WRITE, MAP_PRIVATE|MAP_ANONYMOUS, -1, 0)) == (void*)-1) {
        perror("mmap");
    }

    memset(p, 'c', SIZE);
    sleep(100);

    return 0;
}
```

当前系统内存使用情况:

> ![image](/images/develop_linux_mmp_any_1.png)


调用程序时, 进行匿名共享映射, 内存的状况:

> ![image](/images/develop_linux_mmp_any_2.png)

`used` 增加了 1G, 而 buff/cache 并没有增长. 说明匿名共享时, 并没有占所有 cache.

### 共享匿名映射

```cgo
#include <stdio.h>
#include <string.h>
#include <sys/mman.h>
#include <unistd.h>
#define SIZE 1024*1024*1024

int main(int argc, char* argv[]) {
    char* p;
    if ((p = (char*)mmap(NULL, SIZE, PROT_READ|PROT_WRITE, MAP_SHARED|MAP_ANONYMOUS, -1, 0)) == (void*)-1) {
        perror("mmap");
    }

    memset(p, 'c', SIZE);
    sleep(100);

    return 0;
}
```

当前系统内存使用情况:

> ![image](/images/develop_linux_mmp_shareany_1.png)


调用程序时, 进行共享匿名映射, 内存的状况:

> ![image](/images/develop_linux_mmp_shareany_2.png)

`buffer/cache` 增长了1G,  `shared` 也增长了1G. 即当进行共享匿名映射时, 是从 cache 上当中申请内存.