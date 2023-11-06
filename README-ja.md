# ikesu

:construction: **このリポジトリは開発中のため実用には耐えません** :construction:

## 概要

Mackerelで管理しているホストのメトリックの一定時間以上の途絶を検知して、チェック監視結果としてアラート通知します。  
通常は死活監視（Connectivity）により提供される機能ですが、このツールではクラウドインテグレーションやカスタムメトリックのような死活監視の対象とならない任意のメトリックの途絶の検知を目的としています。

なお、次のような特徴があります。

- 通常のコマンドして動作します。チェック監視プラグインとしては使用できません。
  - cronやAWS Lambdaなどから定期的に実行されることを想定しています。
- YAMLで指定したサービス・ロールに所属する任意のホストをチェックします。
  - Mackerelのホスト管理のプラクティスに則ることで、監視ルールの管理がとても楽になります。
  - サービス・ロールに所属するホストのうち、チェック対象を特定のプロバイダー（EC2やRDSなどの各種クラウド製品）に限定できます。
- 一部のプロバイダーを除き、検証するメトリックが自動的に決定されます。
  - もちろん任意のメトリックを追加してチェックできます。その場合は複数のメトリックのうち、いずれかが投稿されていればOKとなります。
- 現在から過去最大30日まで遡ってチェックできます。デフォルトでは24時間以上の途絶があるとアラートが発報します。
- 通知はCRITICALアラートのみです。WARNINGからCRITICALにアラートレベルが変わるような段階的な通知には非対応です。
- サービスメトリックの途切れ監視はサポートしません。

細やかな設定やサービスメトリックの途切れ監視が必要な場合は [mackerelio-labs/check-mackerel-metric](https://github.com/mackerelio-labs/check-mackerel-metric) の使用をオススメします。

## インストール方法

### Homebrewでインストール

```
brew install tukaelu/tap/ikesu
```

### バイナリを使用する

[リリースページ](https://github.com/tukaelu/ikesu/releases)から使用する環境にあったZipアーカイブをダウンロードしてご使用ください。

## 使用方法

```
NAME:
   ikesu check - Detects disruptions in posted metrics and notifies the host as a CRITICAL alert.

USAGE:
   ikesu check -config <config file> [-dry-run]

OPTIONS:
   --config value, -c value  Specify the path to the configuration file. [$IKESU_CHECK_CONFIG]
   --show-providers          List the inspection metric names corresponding to the provider for each integration. (default: false)
   --dry-run                 Only a simplified display of the check results is performed, and no alerts are issued. (default: false)
   --help, -h                show help
```

- MackerelのAPIキーの指定が必要です。いずれかの方法で指定してください。
  - `MACKEREL_APIKEY`もしくは`IKESU_MACKEREL_APIKEY`の環境変数に指定する。
  - `-apikey`オプションで指定する。

コマンドラインでの実行方法は次のようになります。

```
# APIキーが環境変数 MACKEREL_APIKEY もしくは IKESU_MACKEREL_APIKEY に設定されている場合
ikesu check --conf check.yaml

# 設定をS3バケットから読み込む場合（regionHintを指定しない場合は`ap-northeast-1`として扱います）
ikesu check --conf s3://your_s3_bucket/check.yaml?regionHint=ap-northeast-1

# APIキーをオプションで指定する場合
ikesu -apikey <your api key> check --conf check.yaml
```

## 設定方法

次のようなYAML形式で設定します。各項目については表を確認してください。

```
---
check:
  - name: front-web
    service: blog
    roles:
      - web
    interrupted_interval: 24h
    providers:
      - ec2
      - agent-ec2
    inspection_metrics:
      ec2:
        - "custom.foo.bar"
      agent-ec2:
        - "custom.foo.bar"
  - name: backend
    service: blog
    roles:
      - db
      - functions
    interrupted_interval: 30m
```

| 項目                 | 必須/固定 | 説明                                                        | 初期値 |
| -------------------- | --------- | ----------------------------------------------------------- | ------ |
| check                | 固定      | -                                                           | -      |
| name                 | 必須      | 監視ルール名                                                | -      |
| service              | 必須      | 監視対象とするサービス名                                    | -      |
| roles                | 任意      | 監視対象とするロール名（複数指定可）                        | -      |
| interrupted_interval | 任意      | 途絶を検知する経過時間 *1                                   | 24h    |
| providers            | 任意      | ホストのうちチェック対象を行うプロバイダー *2（複数指定可） | -      |
| inspection_metrics   | 任意      | プロバイダーごとに途絶を検知するメトリック名（複数指定可）  | *3     |

- *1 `10m`や`1h`のような書式で定義してください。最大で30日間（`720h`）まで指定可能です。
- *2 プロバイダーは基本的には[ホスト情報](https://mackerel.io/ja/api-docs/entry/hosts#get)に含まれる`host.meta.cloud.provider`に対応しています。
- *3 プロバイダーによって自動的に検知対象になるメトリックは`--show-provider`を実行して確認してください。
  - mackerel-agentがインストールされたオンプレ環境や、mackerel-container-agentが稼働する環境ではプロバイダーは次のようになります。
    - `agent`
    - `container-agent`
  - mackerel-agentとクラウドインテグレーションが併用される場合、プロバイダー名が次のようになります。
    - `agent-ec2` (Amazon EC2)
    - `agent-virtualmachine` (Azure VM)
    - `agent-computeengine` (Google Complute Engine)

## 注意

- 自動的にメトリックが検知対象となるかはプロバイダーに依存します。
  - メトリック名の定義にワイルドカードを含まず、常時投稿されるメトリックをもつプロバイダーが対象です。
  - `--show-provider`オプションでプロバイダーごとの対応が確認できます。
  - 必要に応じて`inspection_metrics`でメトリック名を直接指定してください。
- 次に該当する場合はチェックを行いません。
  - ホストがmackerel-agent(`provider=agent`) かつ `inspection_metrics` が定義されていない場合はチェックしません。
    - 通常であれば死活監視で検知されるからです。
  - mackerel-container-agent(`provider=container-agent`) かつ `inspection_metrics` が定義されていない場合はチェックしません。
- サービス側の仕様変更により、本ツールが動作が不安定になったり仕様が変更となる場合があります。

## ライセンス

Copyright 2023 tukaelu (Tsukasa NISHIYAMA)

Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with the License. You may obtain a copy of the License at

```
http://www.apache.org/licenses/LICENSE-2.0
```

Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.
