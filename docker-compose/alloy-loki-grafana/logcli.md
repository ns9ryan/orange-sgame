#### https://www.bilibili.com/video/BV1vP9gYbEh6/
#### Download page https://github.com/grafana/loki/releases/
#### https://grafana.com/docs/loki/latest/query/logcli/
#### https://grafana.com/docs/loki/latest/query/logcli/getting-started/#installation
#### https://grafana.com/docs/loki/latest/query/logcli/logcli-tutorial/

> Grafana+Loki+Alloy快速构建企业日志系统
https://www.bilibili.com/video/BV1eoyVY7Eag/

#### wget https://github.com/grafana/loki/releases/download/v3.5.0/logcli-linux-amd64.zip
#### wget https://github.com/grafana/loki/releases/download/v3.5.0/logcli-windows-amd64.exe.zip

```bash
export LOKI_ORG_ID="tenant1"
export LOKI_ADDR=http://192.168.0.116:3100
logcli query '{service_name="ns9-partner-bo-service"}' --tail --no-labels
logcli query '{service_name="ns9-game-user-service"}' --tail --no-labels
```