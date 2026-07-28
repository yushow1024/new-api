
## 4. 创建视频任务

接口：

```
POST /v1/videos
```

### 4.1 文生视频

```bash
curl -X POST "https://api.example.com/api/v1/videos" \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "sora-2",
    "prompt": "一只可爱的小猫在阳光下玩耍，电影感，阳光明亮",
    "aspect_ratio": "16:9",
    "resolution": "720p",
    "seconds": "10",
    "metadata": {
      "order_id": "order_10001"
    }
  }'
```

### 4.2 图生视频

```bash
curl -X POST "https://api.example.com/api/v1/videos" \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "sora-2",
    "prompt": "让参考图中的人物自然微笑并向镜头挥手，真实摄影风格",
    "aspect_ratio": "16:9",
    "resolution": "720p",
    "seconds": "10",
    "images": [
      "https://example.com/reference.png"
    ],
    "metadata": {
      "order_id": "order_10002"
    }
  }'
```

### 4.3 多媒体参考视频

部分模型支持参考视频或参考音频。是否支持以 `/v1/models` 返回的模型参数为准。

bash

```
curl -X POST "https://api.example.com/api/v1/videos" \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "sora-2",
    "prompt": "参考视频的主体动作，生成一段商业广告风格视频",
    "videos": [
      "https://example.com/source.mp4"
    ],
    "audios": [
      "https://example.com/music.mp3"
    ],
    "resolution": "720p",
    "seconds": "10"
  }'
```

### 4.5 视频请求字段

| 字段           | 类型   | 必填 | 说明                                             |
| :------------- | :----- | :--- | :----------------------------------------------- |
| `model`        | string | 是   | 模型名，使用 `/v1/models` 返回的 `id`            |
| `prompt`       | string | 是   | 视频生成提示词                                   |
| `seconds`      | string | 否   | 视频时长，单位秒                                 |
| `duration`     | string | 否   | `seconds` 的兼容别名                             |
| `aspect_ratio` | string | 否   | 画面比例，如 `16:9`、`9:16`、`1:1`               |
| `resolution`   | string | 否   | 分辨率档位，如 `720p`、`1080p`                   |
| `size`         | string | 否   | 尺寸，如 `1280x720`                              |
| `images`       | array  | 否   | 参考图片 URL/Base64 数组                         |
| `videos`       | array  | 否   | 参考视频 URL 数组                                |
| `audios`       | array  | 否   | 参考音频 URL 数组                                |
| `metadata`     | object | 否   | 业务自定义数据，会随任务保存                     |
| `parameters`   | object | 否   | 模型公开扩展参数，字段名以 `/v1/models` 返回为准 |

兼容字段：`image_urls`、`video_urls`、`video_url`、`audio_urls`、`audio_url`、`aspectRatio` 也可用。新接入建议优先使用 `images`、`videos`、`audios`、`aspect_ratio`。

### 4.6 创建响应示例

json

```
{
  "id": "task_xxx",
  "object": "video",
  "model": "sora-2",
  "status": "queued",
  "progress": 0,
  "created_at": 1776843503,
  "size": "16:9",
  "seconds": "10"
}
```

## 5. 查询视频任务

接口：

text

```
GET /v1/videos/{taskId}
```

### 5.1 请求

bash

```
curl -X GET "https://api.example.com/api/v1/videos/task_xxx" \
  -H "Authorization: Bearer YOUR_API_KEY"
```

### 5.2 处理中响应

json

```
{
  "id": "task_xxx",
  "object": "video",
  "model": "sora-2",
  "status": "processing",
  "progress": 45,
  "created_at": 1776843503,
  "size": "16:9",
  "seconds": "10"
}
```

### 5.3 完成响应

json

```
{
  "id": "task_xxx",
  "object": "video",
  "model": "sora-2",
  "status": "completed",
  "progress": 100,
  "created_at": 1776843503,
  "size": "16:9",
  "seconds": "10",
  "video_url": "https://example.com/result.mp4",
  "result": {
    "video_url": "https://example.com/result.mp4"
  },
  "completed_at": 1776843803,
  "processing_time": 300
}
```

### 5.4 失败响应

json

```
{
  "id": "task_xxx",
  "object": "video",
  "model": "sora-2",
  "status": "failed",
  "progress": 100,
  "created_at": 1776843503,
  "error": {
    "message": "generation failed",
    "code": "generation_failed"
  },
  "failed_at": 1776843603
}
```

## 9. 错误处理

常见 HTTP 状态：

| 状态码 | 说明                               |
| :----- | :--------------------------------- |
| `400`  | 请求参数错误                       |
| `401`  | API Key 缺失、无效、未激活或已过期 |
| `402`  | 余额不足                           |
| `404`  | 任务不存在                         |
| `429`  | 请求频率或并发超过限制             |
| `500`  | 服务内部错误                       |

错误示例：

json

```
{
  "code": 401,
  "message": "API Key 无效",
  "error": {
    "type": "authentication_error",
    "details": "API Key 不存在"
  }
}
```

## 11. 完整示例流程

### 11.1 视频完整流程

bash

```
# 1. 创建视频任务
curl -X POST "https://api.example.com/api/v1/videos" \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "sora-2",
    "prompt": "一只小猫在阳光下奔跑",
    "aspect_ratio": "16:9",
    "resolution": "720p",
    "seconds": "10"
  }'

# 2. 查询视频任务
curl -X GET "https://api.example.com/api/v1/videos/task_xxx" \
  -H "Authorization: Bearer YOUR_API_KEY"
```