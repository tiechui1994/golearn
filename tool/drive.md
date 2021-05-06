# Google Drive

[Google Drive API](https://developers.google.com/drive/api/v3/reference/?apix=true)

> 要想使用 `Google Drive API`, 需要先获取相应 scope 的Token, 然后才能访问.

## google 大文件下载

### 1. 获取要下载的文件ID:

- 先登录到谷歌云盘[**Google Drive**](https://drive.google.com/drive/my-drive)

- 鼠标移动到要下载的文件, 然后右击, 然后点击 **获取链接**, 如下图所示:

![image](/images/develop_api_google_getlink.png)

然后, 获得了文件的分享链接:

![image](/images/develop_api_google_link.png)

- 链接: `https://drive.google.com/file/d/xxx/view?usp=sharing`, 其中 `xxx` 就是需要获取的文件ID, 复制出来,
以供后面使用.

### 2.获取 OAuth 的 token:

- 进入谷歌开发官网, [**OAuth 2.0 Playground**](https://developers.google.com/oauthplayground/)

- 在 `Step 1 Select & authorize APIs` 当中选择向下滚动, 选择 **Drive API v3** 选项, 并展开, 然后选择 
`https://www.googleapis.com/auth/drive.readonly` 选项.

- 点击下方蓝色的 `Authorize APIs` 按钮, 跳转到Google的授权页面, 登录Google账号并同意授权.

- 授权完成后, 会再次跳回到谷歌开发官网, 点击 `Step 2 Exchange authorization code for tokens` 当中蓝色按钮,
`Exchange authorization code for tokens`, 发起 code 换 token 的请求.

- 复制出 `Access token` 当中的内容, 后面会使用到的.

### 3.下载文件

- 在 Linx 或者 OS X 当中输入如下命令

```
curl -C - -H "Authorization: Bearer {TOKEN}" https://www.googleapis.com/drive/v3/files/{ID}?alt=media 
     -o {FILE}
```

> `{ID}` 是第1步获取到的文件ID, `{TOKEN}` 是第2步获取的授权Token. `{FILE}` 是对下载的文件的命名.
> 注: 如果文件特别大, 1个小时下载不完, 只需要根据第2步获取一个 `{TOKEN}`, 然后继续按照上述的命令继续下载文件, 直到文件
> 下载完成.
