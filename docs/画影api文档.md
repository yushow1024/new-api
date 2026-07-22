# API 接入文档

Base URL:
http://ai-studio.aixyzz.com/v1

所有接口都使用 JSON。
所有请求都需要携带请求头:
Authorization: Bearer <API_KEY>

API Key 在 API 控制台创建。完整密钥只显示一次，请创建后立即保存。

一、获取可用模型

curl http://ai-studio.aixyzz.com/v1/models \
-H "Authorization: Bearer <API_KEY>"

返回示例:
{
"object": "list",
"data": [
{
"id": "seedance-2.0",
"object": "model",
"display_name": "Seedance 2.0",
"status": "active",
"pricing": {
"billingMode": "per_item",
"unitPrice": "1",
"currency": "POINT"
},
"inputModes": ["text", "image", "video", "audio"],
"mediaLimits": { "images": 9, "videos": 3, "audios": 3 },
"durations": ["5s", "6s", "7s", "8s", "9s", "10s", "11s", "12s", "13s", "14s", "15s"],
"resolutions": ["720p"],
"aspectRatios": ["16:9", "9:16"],
"defaults": { "duration": "5s", "resolution": "720p", "aspectRatio": "16:9" }
}
]
}

二、创建视频生成任务

示例 1：文生视频

curl http://ai-studio.aixyzz.com/v1/videos/generations \
-H "Authorization: Bearer <API_KEY>" \
-H "Content-Type: application/json" \
-d '{
"model": "seedance-2.0",
"prompt": "原创科幻城市夜景，一辆无品牌悬浮飞车沿高空赛道高速前进，电影感，光影真实，无文字，无 logo",
"duration": 5,
"resolution": "720p",
"aspectRatio": "16:9"
}'

示例 2：全能参数

curl http://ai-studio.aixyzz.com/v1/videos/generations \
-H "Authorization: Bearer <API_KEY>" \
-H "Content-Type: application/json" \
-d '{
"model": "seedance-2.0",
"requestId": "order_20260628_0001",
"prompt": "参考图片主体、参考视频运镜和参考音频节奏，生成一段原创产品展示视频，画面干净，无文字，无 logo",
"images": [
"https://example.com/images/01.png",
"https://example.com/images/02.png",
"https://example.com/images/03.png",
"https://example.com/images/04.png",
"https://example.com/images/05.png",
"https://example.com/images/06.png",
"https://example.com/images/07.png",
"https://example.com/images/08.png",
"https://example.com/images/09.png"
],
"videos": [
"https://example.com/videos/01.mp4",
"https://example.com/videos/02.mp4",
"https://example.com/videos/03.mp4"
],
"audios": [
"https://example.com/audios/01.mp3",
"https://example.com/audios/02.mp3",
"https://example.com/audios/03.mp3"
],
"duration": 8,
"resolution": "720p",
"aspectRatio": "16:9"
}'

示例 3：首尾帧

curl http://ai-studio.aixyzz.com/v1/videos/generations \
-H "Authorization: Bearer <API_KEY>" \
-H "Content-Type: application/json" \
-d '{
"model": "seedance-2.0",
"prompt": "镜头从首帧画面自然运动到尾帧画面，光影稳定，动作连贯，无文字，无 logo",
"firstFrame": "https://example.com/start.png",
"lastFrame": "https://example.com/end.png",
"duration": 5,
"resolution": "720p",
"aspectRatio": "16:9"
}'

创建任务返回示例:
{
"id": "job_xxx",
"object": "video.generation",
"created": 1760000000,
"model": "seedance-2.0",
"status": "processing",
"request_id": "req_xxx",
"usage": {
"cost": "5",
"currency": "POINT"
},
"balance": {
"available": "95",
"reserved": "5",
"currency": "POINT"
},
"data": []
}

三、查询任务结果

curl http://ai-studio.aixyzz.com/v1/videos/<JOB_ID> \
-H "Authorization: Bearer <API_KEY>"

成功返回示例:
{
"id": "job_xxx",
"object": "video.generation",
"created": 1760000000,
"model": "seedance-2.0",
"status": "succeeded",
"data": [
{
"url": "https://example.com/video.mp4",
"cover_url": "https://example.com/cover.jpg"
}
]
}

字段参数说明:
| 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| model | string | 是 | 模型 ID，使用 /models 返回的 data[].id |
| prompt | string | 是 | 视频提示词 |
| duration | number/string | 是 | 视频时长。可传 5，也可传 "5s"。必须在模型 durations 内 |
| resolution | string | 是 | 分辨率。必须在模型 resolutions 内 |
| aspectRatio | string | 是 | 画面比例。必须在模型 aspectRatios 内 |
| requestId | string | 否 | 你的请求编号。重复提交同一个 requestId 会返回同一任务，避免重复创建和重复扣费 |
| firstFrame | string | 否 | 首帧图片 URL。只做首帧图生视频时传这个字段 |
| lastFrame | string | 否 | 尾帧图片 URL。做首尾帧时和 firstFrame 一起传；不能单独只传 lastFrame |
| images | string[] | 否 | 图片 URL 列表，数量不能超过模型 mediaLimits.images |
| videos | string[] | 否 | 视频 URL 列表，数量不能超过模型 mediaLimits.videos |
| audios | string[] | 否 | 音频 URL 列表，数量不能超过模型 mediaLimits.audios |

状态说明:
| status | 含义 | 下一步 |
|---|---|---|
| processing | 任务处理中 | 继续查询 |
| succeeded | 任务成功 | 读取 data[].url |
| failed | 任务失败 | 查看 error |
| canceled | 任务已取消 | 不再查询 |

错误返回格式:
{
"error": {
"message": "错误说明",
"type": "invalid_request_error",
"code": "invalid_input",
"request_id": "req_xxx"
}
}

常见错误:
| code | 含义 |
|---|---|
| invalid_api_key | API Key 缺失或错误 |
| insufficient_balance | 余额不足 |
| model_not_found | 模型不存在或未授权 |
| model_unavailable | 模型未启用 |
| invalid_input | 参数不符合模型能力 |
| rate_limit_exceeded | 请求过快或并发过高 |

扣费说明:
- 创建任务会占用或扣除账户积分。
- 任务成功后消耗会记入任务记录。
- 任务失败时不应产生最终扣费，具体以任务记录和额度流水为准。

接入建议:
- 先调用 /models 获取模型和参数能力，再创建任务。
- 每次业务请求建议生成唯一 requestId。网络超时后用同一个 requestId 重试。
- 创建任务后保存返回的 id，用它查询任务结果。
- 不要把 API Key 放到前端网页或客户端 App 内。
