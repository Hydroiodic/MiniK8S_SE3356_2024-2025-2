
from openai import OpenAI
from PIL import Image
import base64
import os
import requests
from io import BytesIO
from typing import Dict, Any

def handle(param) -> Dict[str, Any]:
    client = OpenAI(
        base_url="https://dashscope.aliyuncs.com/compatible-mode/v1",
        api_key="sk-dd41976c22094389bfa54cfe1aa04e34",
    )

    """
    调用模型生成图片描述
    参数:
        param: 包含base64编码图片的字典
    返回:
        添加模型生成的描述的字典
    """
    encoded_image_text = param["processed_image_base64"]
    base64_image = f"data:image;base64,{encoded_image_text}"
    completion = client.chat.completions.create(
        model="qwen-vl-max-latest", # 此处以qwen-vl-max-latest为例，可按需更换模型名称。模型列表：https://help.aliyun.com/model-studio/getting-started/models
        messages=[
            {
                "role": "system",
                "content": [{"type": "text", "text": "You are a helpful assistant."}],
            },
            {
                "role": "user",
                "content": [
                    {
                        "type": "image_url",
                        "image_url": {
                            "url": base64_image
                        },
                    },
                    {"type": "text", "text": "请描述图中景象"},
                ],
            },
        ],
    )
    description = completion.choices[0].message.content
    # param["description"] = description
    return {
        "description": description
    }