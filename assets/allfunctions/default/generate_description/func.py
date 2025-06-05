
from openai import OpenAI
from PIL import Image
import base64
import os
import requests
from io import BytesIO
from typing import Dict, Any

def handle(param) -> Dict[str, Any]:
    client = OpenAI(
        base_url="http://47.242.151.133:24576/v1/",
        api_key="ml2025",
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
    chat_response = client.chat.completions.create(
        model="Qwen/Qwen2.5-VL-3B-Instruct",
        messages=[
            {"role": "system", "content": "You are a helpful assistant that describes images in detail."},
            {
                "role": "user",
                "content": [
                    {
                        "type": "image_url",
                        "image_url": {
                            "url": base64_image
                        },
                    },
                    {"type": "text", "text": "Please describe this image in detail."},
                ],
            },
        ],
    )
    
    description = chat_response.choices[0].message.content
    # param["description"] = description
    return {
        "description": description
    }