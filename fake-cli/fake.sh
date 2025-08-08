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
    echo "$(date '+%H:%M:%S') - Input: $input"
    # create file dummy_code.txt with random content
    echo "This is dummy code ${i}" > dummy_code_${i}.txt
    echo "Created file dummy_code_${i}.txt with random content"
    sleep 2
done

exit 0