# clipper

## 概要
youtubeチャンネルの動画情報とコメントを収集して、そこからムービークリップリストを作成し再生するプログラムです。

## インストール

### 前提
- clipperはオプションでrunModeを指定することで情報を収集して静的ページを作るクローラーと静的ページを公開するWEBサーバーになります。
- clipperをWEBサーバーにする場合は証明書を用意する必要があります。
- clipperをWEBサーバーとして外部に公開する場合はドメインを取得する必要があります。
- クローラーが作った静的ページをapache等で公開することも可能です。
- youtube apiを利用するためapi keyを取得する必要があります。
- 更新の通知にtwitter apiを利用するためtwitterのapi key, api key secret, access token, access tokeb secretを取得する必要があります。

### インストール 
- デフォルトで/usr/local/clipperにインストールされます。
- 共通インストール
```
go get -u github.com/munoudesu/clipper
cd $GOPATH/src/github.com/munoudesu/clipper
sudo -E ./install.sh
sudo vi /usr/local/clipper/etc/clipper.conf
sudo vi /usr/local/clipper/etc/youtube_data_api_key_file
sudo vi /usr/local/clipper/etc/twitter_api_key_file
sudo systemctl daemon-reload
```
- clipper webサーバーインストール
```
sudo systemctl enable clipper.service
sudo systemctl start clipper.service
```
- コマンドサンプル
  - NOTEにGCP-GCEのubuntuにインストールした際のコマンドリストがあります。

## 動作
- クローラー
  - cronで1日1回情報を収集し、静的ページを更新します。データに変更があればtwitterに通知します。
- webサーバー
  - httpsで静的ページを公開します。https://<domain>/root/index.htmlでアクセスできるようになります。

## 設定

### youtube_data_api_key_file
- 1行に１つapi keyを書きます。
- 複数のapi keyを書いた場合はローテーションしながら使われます。
```
<api key1>
<api key2>
```

### twitter_api_key_file
- api key, api secret key, access token, access token secretを1行づつ書きます。
- パーミッションは600にしてください
```
<api key>
<api secret key>
<access token>
<access token secret>
```

### clipper.conf
- youtubeセクション
  - maxVideos
    - 収集する最大動画数を指定します。
- youtube.channelsセクション
  - name
    - 任意の名前を指定します。
  - channelId
    - チャンネルのIDを指定します。
- twitterセクション
  - tweetLinkRoot
    - ツイートに含めるリンクのURLを指定します。
  - tweetComment
    - ツイートに含めるコメントを指定します。
- twitter.usersセクション
  - twitter.users.任意の名前
    - 任意の名前を指定します。youtube.channelsセクションの任意の名前と同じになる必要があります。
  - tags
    - ツイートに含めたいタグリストを指定します。
- buulderセクション
  - maxDuration
    - ムービークリップの時間の最大を指定します。これ以上長い場合はこの時間に調整されます。
  - adjustStartTimeSpan
    - 開始時間の誤差を吸収する時間を指定します。開始時間がこの範囲のものがある場合は同一のものとして扱われます。
- webセクション
  - addrPort
    - webサーバーが使用するアドレスとポートを指定します。
  - tlsCertPath
    - 中間証明書も連結した完全な証明書のパスを指定します。
  - tlsKeyPath
    - 証明書のプライベートキーを指定します。

## コマンド引数
- checkChannelModified
  - チャンネルの変更をチェックしません。DB上にデータが存在すれば次に進みます。
- checkVideoModified
  - どうがの変更をチェックしません。 DB上にデータが存在すれば次に進みます。
- checkCommentModified
  - コメントの変更をチェックしません。 DB上にデータが存在すれば次に進みます。
- config
  - 読み込む設定ファイルを指定します。
- rebild
  - 情報に変更がない場合でも静的ページの更新を行います。
- renotify
  - 情報に変更がない場合でもtwitterへ通知を行います。
- runMode
  - 動作モードを指定します。
- searchChannel
  - チャンネルの検索をします。falseを指定するとスキップできます。
- searchVideo
  - 動画の検索をします。falseを指定するとスキップできます。
- searchComment
  - コメントの検索をします。falseを指定するとスキップできます。
- skipSearch
  - 情報収集処理をスキップします。
- skipBuild
  - 静的ページ作成処理をスキップします。
- skipNotify
  - 通知処理をスキップします。
-　verbose
  - 動作ログが出力されるようになります。
  
## 注意
設定で取得動画数を少なくしない限り、ほとんどの場合youtube data api v3 のquota制限に達するため、割り当てを増加させたapi keyが必要になります。

