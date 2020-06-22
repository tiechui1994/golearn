# Google drive

[Google Drive API](https://developers.google.com/drive/api/v3/reference/?apix=true)

> 要想使用 `Google Drive API`, 需要先获取相应 scope 的Token, 然后才能访问.

## google 大文件下载

### 1. Get the file ID:

- go to your [**Google Drive**](https://drive.google.com/drive/my-drive)

- right-click the file you want to download

- click on **Get shareable link**

- link looks like this: `https://drive.google.com/open?id=xxx` where `xxx` is the ID you will need 
after few more steps

### 2. Get an OAuth token:

- go to [**OAuth 2.0 Playground**](https://developers.google.com/oauthplayground/)

- under **Select the Scope box**, scroll down to **Drive API v3**

- expand it and select `https://www.googleapis.com/auth/drive.readonly`

- click on the blue `Authorize APIs` button

- login if you are prompted and then click on `Exchange authorization code for tokens` to get a token

- copy the `Access token` for further use

### Download the file

- on Linx Or OS X

```
curl -C - -H "Authorization: Bearer ttt" https://www.googleapis.com/drive/v3/files/xxx?alt=media 
     -o zzz
```
