#!/bin/bash

# CLI tool nhận input từ stdin và in ra stdout mỗi 3 giây
# Usage: echo "Hello" | ./fake.sh

echo "CLI tool đã khởi động. Nhận input từ stdin và in ra mỗi 3 giây..."
echo "Nhấn Ctrl+C để thoát"
echo ""

# Đọc input từ stdin
input=$(cat)

# In ra input đã nhận được
echo "Đã nhận input: $input"
echo ""

# in ra input mỗi 2 giây, tối đa 5 lần
for i in {1..5}; do
    echo "thinking..."
    sleep 2
done

echo "This is implement plan"
echo "12345566"

exit 0