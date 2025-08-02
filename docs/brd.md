# Business requirements document

## Overview

Tool để tự động hóa developer tasks, user của tool chính là developer.

- User sẽ log 1 cái task ở status TODO.
- Sau đó bấm start planning, task sẽ chuyển từ TODO sang PLANNING và AI Agent CLI sẽ bốc task đi plan.
- Sau khi plan xong, task sẽ chuyển từ PLANNING sang PLAN REVIEWING, nếu plan okay thì user sẽ bấm start implement.
- Khi start implement, task sẽ chuyển từ PLAN REVIEWING sang IMPLEMENTING.
- Khi implement xong, task sẽ chuyển từ IMPLEMENTING sang CODE REVIEWING.
- Sau khi Pull request được merge, task sẽ chuyển từ CODE REVIEWING sang DONE.
- Nếu task đang làm, user có thể bấm cancel để dừng không làm task đó nữa.

## Features

- Là user, tôi có thể tạo và cấu hình được projects.
- Là user, tôi có thể log task vào cột TODO.
- Là user, tôi có thể start planning task.
- Là user, tôi có thể review plan của task và start implement.
- Là user, tôi muốn các task sẽ được implement riêng lẽ ở những branch riêng biệt.
