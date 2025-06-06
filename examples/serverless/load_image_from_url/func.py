from openai import OpenAI
from PIL import Image
import base64
import os
import requests
from io import BytesIO
from typing import Dict, Any
def handle(param) -> Dict[str, Any]:
    """
    从HTTP地址加载图片并返回初始处理字典
    参数:
        image_url: 图片的HTTP地址
    返回:
        包含图片和其他初始信息的字典
    """
    try:
        response = requests.get(param["image_url"])
        response.raise_for_status()  # 检查请求是否成功
        
        image = Image.open(BytesIO(response.content))
        buffered = BytesIO()
        image.save(buffered, format="PNG")
        encoded_image = base64.b64encode(buffered.getvalue()).decode("utf-8")

        return {
            "image_base64": encoded_image,
            "temp_files":[]
        }
    except requests.RequestException as e:
        raise Exception(f"Failed to download image from URL: {image_url}. Error: {str(e)}")