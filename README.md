# PVM

PVM 是一个基于 Go 和 libvirt 的虚拟化管理工具，提供三套入口：

- CLI 命令行工具
- gRPC API
- HTTP/JSON API

项目面向 Linux KVM/QEMU + libvirt 运行环境，覆盖以下核心能力：

- 虚拟机定义、查询和生命周期管理
- 从 ISO 安装新虚拟机
- 从已安装镜像直接启动虚拟机
- 存储池和卷管理
- libvirt 虚拟网络管理
- host interface 管理
- 域快照管理
- ZFS 卷真快照管理

## 特性概览

- 默认连接 `qemu:///system`，支持 `--uri` 或 `LIBVIRT_DEFAULT_URI` 覆盖
- CLI 和 HTTP 都使用 `snake_case` 字段名
- `pvm serve` 默认同时监听：
  - HTTP: `127.0.0.1:8080`
  - gRPC: `127.0.0.1:9090`
- API 统一使用 Bearer Token 鉴权
- VM 启动支持两种模式：
  - `clone`: 复制源镜像或源卷到目标卷，再创建 VM
  - `direct`: 直接挂载源镜像或源卷启动，不做复制
- ZFS 卷快照通过 `zfs` 命令完成实际 snapshot / rollback，libvirt 负责资源发现和状态协调

## 项目结构

```text
cmd/pvm                 CLI 入口
api/proto/pvm/v1        protobuf 定义
api/gen/pvm/v1          生成的 protobuf / gRPC 代码
internal/cli            CLI 命令实现
internal/server         gRPC 与 HTTP 服务
internal/service        共享业务层
internal/backend        libvirt 后端与非 Linux stub
internal/xmlbuild       libvirtxml 组装
internal/zfs            ZFS 快照封装
```

## 运行要求

正式运行环境：

- Linux
- libvirt
- QEMU / KVM
- cgo 可用

可选能力：

- ZFS
- `zfs` 命令行工具

开发环境：

- Go `1.21.x`
- `protoc`
- `protoc-gen-go`
- `protoc-gen-go-grpc`

说明：

- macOS 可以用于开发、生成代码和跑大部分单元测试。
- 非 Linux 或未启用 cgo 时，项目仍可编译，但 libvirt 后端会返回 `unsupported`。

## 依赖安装

Ubuntu / Debian 示例：

```bash
sudo apt-get update
sudo apt-get install -y \
  libvirt-daemon-system \
  libvirt-clients \
  libvirt-dev \
  qemu-kvm \
  pkg-config \
  protobuf-compiler
```

如果需要 ZFS 卷快照：

```bash
sudo apt-get install -y zfsutils-linux
```

安装 protobuf Go 插件：

```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.35.1
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.5.1
```

## 构建与测试

```bash
go mod download
go test ./...
go build -o bin/pvm ./cmd/pvm
```

生成 protobuf 代码：

```bash
go generate ./api/proto/pvm/v1
```

或直接执行：

```bash
./scripts/genproto.sh
```

## CLI 快速开始

查看帮助：

```bash
./bin/pvm --help
./bin/pvm vm --help
./bin/pvm serve --help
```

全局参数：

- `--uri`: libvirt URI，默认 `qemu:///system`
- `--username`: libvirt 用户名
- `--password-stdin`: 从标准输入读取 libvirt 密码
- `-o, --output text|json`: 输出格式

相关环境变量：

- `LIBVIRT_DEFAULT_URI`
- `LIBVIRT_USERNAME`
- `LIBVIRT_PASSWORD`
- `PVM_API_TOKEN`

## 常用 CLI 示例

从 ISO 安装并启动虚拟机：

```bash
./bin/pvm vm install \
  --name ubuntu-iso \
  --pool tank \
  --disk-name ubuntu-iso-root \
  --disk-size-gib 40 \
  --disk-format raw \
  --iso-path /var/lib/libvirt/boot/ubuntu-24.04.iso \
  --memory-mib 4096 \
  --vcpu 4 \
  --network default \
  --graphics-type vnc \
  --autostart
```

从已有镜像克隆后启动：

```bash
./bin/pvm vm launch \
  --name ubuntu-clone \
  --mode clone \
  --image-path /var/lib/libvirt/images/ubuntu-base.qcow2 \
  --target-pool tank \
  --target-volume ubuntu-clone-root \
  --target-format qcow2 \
  --memory-mib 4096 \
  --vcpu 4 \
  --network default \
  --graphics-type vnc
```

从已有 libvirt volume 直接启动：

```bash
./bin/pvm vm launch \
  --name ubuntu-direct \
  --mode direct \
  --source-volume tank/ubuntu-golden \
  --memory-mib 4096 \
  --vcpu 4 \
  --network default \
  --graphics-type spice
```

查看图形连接信息：

```bash
./bin/pvm vm graphics ubuntu-direct
```

生命周期管理：

```bash
./bin/pvm vm list
./bin/pvm vm start ubuntu-direct
./bin/pvm vm shutdown ubuntu-direct
./bin/pvm vm reboot ubuntu-direct
./bin/pvm vm pause ubuntu-direct
./bin/pvm vm resume ubuntu-direct
./bin/pvm vm undefine ubuntu-direct
```

给虚拟机增删改网卡：

```bash
./bin/pvm vm nic add ubuntu-direct --network default --alias net1 --model virtio
./bin/pvm vm nic update ubuntu-direct --alias net1 --bridge br0 --model virtio
./bin/pvm vm nic remove ubuntu-direct --alias net1
```

定义 ZFS 存储池：

```bash
./bin/pvm pool define \
  --name tank \
  --type zfs \
  --source-name tank/vm \
  --target-path /tank/vm \
  --autostart
```

创建卷：

```bash
./bin/pvm volume create \
  --pool tank \
  --name demo-root \
  --capacity-bytes 42949672960 \
  --format raw
```

定义虚拟网络：

```bash
./bin/pvm network define \
  --name isolated \
  --bridge virbr10 \
  --forward-mode nat \
  --ipv4-cidr 192.168.100.1/24 \
  --dhcp-start 192.168.100.10 \
  --dhcp-end 192.168.100.200 \
  --autostart
```

定义 host interface：

```bash
./bin/pvm iface define \
  --name br-test0 \
  --type bridge \
  --member eth0 \
  --member eth1 \
  --mtu 9000 \
  --start-mode onboot
```

域快照：

```bash
./bin/pvm snapshot vm create --vm ubuntu-direct --name before-upgrade --memory
./bin/pvm snapshot vm list --vm ubuntu-direct
./bin/pvm snapshot vm revert --vm ubuntu-direct --name before-upgrade
./bin/pvm snapshot vm delete --vm ubuntu-direct --name before-upgrade
```

ZFS 卷快照：

```bash
./bin/pvm snapshot volume create --pool tank --volume demo-root --name pre-patch
./bin/pvm snapshot volume list --pool tank --volume demo-root
./bin/pvm snapshot volume rollback --pool tank --volume demo-root --name pre-patch
./bin/pvm snapshot volume delete --pool tank --volume demo-root --name pre-patch
```

JSON 输出：

```bash
./bin/pvm --output json vm list
```

## HTTP / gRPC 服务

启动服务：

```bash
export PVM_API_TOKEN='replace-with-a-long-random-token'

./bin/pvm serve \
  --http-addr 127.0.0.1:8080 \
  --grpc-addr 127.0.0.1:9090
```

也可以通过文件传入 Token：

```bash
./bin/pvm serve --token-file /etc/pvm/token
```

说明：

- 服务默认只监听 `127.0.0.1`
- 首版不包含 TLS / mTLS
- HTTP 使用 `Authorization: Bearer <token>`
- gRPC 使用 metadata `authorization: Bearer <token>`

### HTTP 接口

公开健康检查：

```bash
curl http://127.0.0.1:8080/healthz
```

查询系统信息：

```bash
curl -H "Authorization: Bearer ${PVM_API_TOKEN}" \
  http://127.0.0.1:8080/v1/system
```

查询虚拟机列表：

```bash
curl -H "Authorization: Bearer ${PVM_API_TOKEN}" \
  http://127.0.0.1:8080/v1/vms
```

通过 HTTP 从已有镜像克隆启动 VM：

```bash
curl -X POST \
  -H "Authorization: Bearer ${PVM_API_TOKEN}" \
  -H "Content-Type: application/json" \
  http://127.0.0.1:8080/v1/vms:launch \
  -d '{
    "spec": {
      "name": "api-clone",
      "mode": "clone",
      "memory_mib": "4096",
      "vcpu": 4,
      "image_path": "/var/lib/libvirt/images/ubuntu-base.qcow2",
      "target_pool": "tank",
      "target_volume_name": "api-clone-root",
      "target_format": "qcow2",
      "networks": [
        {
          "network": "default",
          "model": "virtio"
        }
      ],
      "graphics": [
        {
          "type": "vnc",
          "listen": "127.0.0.1",
          "auto_port": true
        }
      ]
    }
  }'
```

常用 REST 风格路径：

- `GET /v1/system`
- `GET /v1/vms`
- `GET /v1/vms/{name}`
- `POST /v1/vms:define`
- `POST /v1/vms:install`
- `POST /v1/vms:launch`
- `POST /v1/vms/{name}:start`
- `POST /v1/vms/{name}:shutdown`
- `GET /v1/pools`
- `GET /v1/pools/{pool}/volumes`
- `GET /v1/networks`
- `GET /v1/interfaces`
- `POST /v1/vms/{vm}/snapshots`
- `POST /v1/pools/{pool}/volumes/{volume}/snapshots`

说明：

- HTTP 层使用 protobuf JSON。
- 为了与 CLI 对齐，字段名使用 `snake_case`。
- 按 protobuf JSON 规范，`uint64` 字段会序列化为字符串，例如 `memory_mib`、`capacity_bytes`。

### gRPC 接口

服务名：

- `pvm.v1.SystemService`
- `pvm.v1.VMService`
- `pvm.v1.StorageService`
- `pvm.v1.NetworkService`
- `pvm.v1.InterfaceService`
- `pvm.v1.SnapshotService`

使用 `grpcurl` 查询虚拟机列表：

```bash
grpcurl \
  -plaintext \
  -H "authorization: Bearer ${PVM_API_TOKEN}" \
  -import-path ./api/proto \
  -proto pvm/v1/vm.proto \
  127.0.0.1:9090 \
  pvm.v1.VMService/ListVMs
```

Go 客户端可直接复用 `api/gen/pvm/v1` 里的生成代码。

## 镜像启动模式说明

`pvm vm launch` 与对应 API 支持两类源：

- `image_path`: 本地 qcow2/raw 镜像文件
- `source_volume`: 已存在的 libvirt 卷，格式为 `pool/name`

`clone` 模式：

- 必须提供 `target_pool` 和 `target_volume_name`
- 如果源是镜像文件，会先导入目标池
- 如果源是 libvirt 卷，会通过 libvirt 复制为新卷
- VM 使用新卷启动

`direct` 模式：

- 不会复制源镜像
- VM 直接引用源文件或源卷
- 返回结果会标记 `shared_source = true`

建议：

- 需要隔离和可回收性时使用 `clone`
- 需要最快启动速度或做 golden image 共享时使用 `direct`

## ZFS 卷快照说明

卷快照只支持 ZFS zvol-backed volume。

当前约束：

- 需要系统安装 `zfs` 命令
- 只支持能映射到 `/dev/zvol/...` 的卷路径
- 创建和回滚快照时，会尝试暂停正在使用该卷且处于运行状态的虚拟机，完成后恢复
- 非 ZFS 后端不支持卷真快照

示例流程：

1. 创建 ZFS pool 并在 libvirt 中定义对应存储池
2. 创建 volume
3. 使用 VM 挂载该 volume
4. 执行 `pvm snapshot volume create`
5. 必要时执行 `rollback`

## 已知限制

- 正式运行仅支持 Linux + cgo + libvirt
- HTTP 服务默认本地监听，不包含远程 TLS/mTLS
- gRPC 服务当前未开启 reflection，`grpcurl` 需要显式提供 `.proto`
- CLI 的 host interface 参数目前主要覆盖名称、类型、成员、MAC、MTU、启动模式；更复杂的协议字段更适合走 gRPC / HTTP
- `vm undefine` 不会自动删除关联卷
- `volume delete` 会检查卷是否仍被虚拟机使用
- `direct` 模式会共享底层镜像，请确认上层业务接受共享写入或自行配合只读策略

## 开发备注

- libvirt XML 统一通过 `libvirtxml` 构造，不手写长 XML
- CLI 和服务端共享 `internal/service`
- HTTP 层复用 `grpc-gateway/runtime`，但路由注册由项目手工维护
- `go test ./...` 在非 Linux 环境主要覆盖 CLI、服务层、gateway、XML 生成和参数校验

## 许可证

如需对外发布，请按你的实际开源策略补充许可证文件。
