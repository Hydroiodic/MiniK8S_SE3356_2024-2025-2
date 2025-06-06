from openai import OpenAI
from PIL import Image
import base64
import os
import requests
from io import BytesIO
from typing import Dict, Any
def handle(param) -> Dict[str, Any]:
    """
    预处理图片（当前仅转换格式，可扩展）
    参数:
        param: 包含图片数据的字典
    返回:
        更新后的处理字典（添加预处理后的图片）
    """
    encoded_image_string = param.get("image_base64")
    
    decoded_image_bytes = base64.b64decode(encoded_image_string)

        # 2. 使用 PIL 从二进制数据中加载图片对象
    image = Image.open(BytesIO(decoded_image_bytes))
    
    # 确保图片是RGB模式（如果需要）
    if image.mode != "RGB":
        image = image.convert("RGB")
    
    # 保存预处理后的临时文件
  # 4. 将预处理后的 PIL Image 对象重新编码为 Base64 字符串
    output_buffer = BytesIO()
    # 统一保存为PNG格式，保证一致性，如果你有其他偏好，也可以改
    image.save(output_buffer, format="PNG")
    processed_encoded_image = base64.b64encode(output_buffer.getvalue()).decode("utf-8")
    
    return {
        "processed_image_base64": processed_encoded_image,
    }
