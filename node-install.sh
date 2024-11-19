#!/bin/bash

# 创建 'masa' 用户并设置主目录
useradd -m masa

# 设置 RPC_URL
RPC_URL=https://ethereum-sepolia.publicnode.com 

# 将 RPC_URL 追加到 masa 用户的 .bash_profile
echo "export RPC_URL=${RPC_URL}" | tee -a /home/masa/.bash_profile

# 设置 masa 用户主目录的权限
chown masa:masa /home/masa/.bash_profile

# 安装 Node.js 和 Yarn
curl -fsSL https://deb.nodesource.com/setup_lts.x | bash -
curl -sS https://dl.yarnpkg.com/debian/pubkey.gpg | gpg --dearmor -o /usr/share/keyrings/yarn-archive-keyring.gpg
echo "deb [signed-by=/usr/share/keyrings/yarn-archive-keyring.gpg] https://dl.yarnpkg.com/debian/ stable main" | tee /etc/apt/sources.list.d/yarn.list
apt-get update -y && apt-get install -y yarn nodejs jq

# 构建 go 二进制文件
make build
cp masa-node /usr/local/bin/masa-node
chmod +x /usr/local/bin/masa-node

# 确定全局 npm 模块路径并设置 NODE_PATH
GLOBAL_NODE_MODULES=$(npm root -g)
export NODE_PATH=$GLOBAL_NODE_MODULES

MASANODE_CMD="/usr/bin/masa-node --port=4001 --udp=true --tcp=false --start --bootnodes=${BOOTNODES}"

# 为 masa-node 创建 systemd 服务文件
cat <<EOF | sudo tee /etc/systemd/system/masa-node.service
[Unit]
Description=MASA Node 服务
After=network.target

[Service]
User=masa
WorkingDirectory=/home/masa
Environment="RPC_URL=${RPC_URL}"
ExecStart=$MASANODE_CMD
Restart=always

[Install]
WantedBy=multi-user.target
EOF

# 确保服务文件由 root 拥有
sudo chown root:root /etc/systemd/system/masa-node.service

# 重新加载 systemd 守护进程
sudo systemctl daemon-reload

# 启用并启动 masa-node 服务
sudo systemctl enable masa-node
sudo systemctl start masa-node
