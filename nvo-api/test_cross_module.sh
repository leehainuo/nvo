#!/bin/bash

echo "=== 测试简化后的服务注入架构 ==="
echo ""

# 1. 测试基础 API
echo "1. 测试角色列表 API..."
curl -s http://localhost:8080/api/v1/roles | python3 -m json.tool
echo ""

# 2. 测试用户列表 API
echo "2. 测试用户列表 API..."
curl -s http://localhost:8080/api/v1/users | python3 -m json.tool
echo ""

# 3. 测试跨模块调用 - 获取用户及其角色详情
echo "3. 测试跨模块调用 - GET /users/1/roles..."
curl -s http://localhost:8080/api/v1/users/1/roles | python3 -m json.tool
echo ""

echo "=== 测试完成 ==="
