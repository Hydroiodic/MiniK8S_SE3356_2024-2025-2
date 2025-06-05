from openai import OpenAI
from PIL import Image
import base64
import os
import requests
from io import BytesIO
from typing import Dict, Any

def handle(param) -> Dict[str, Any]:
    """
    将图片编码为base64格式
    参数:
        param: 包含图片数据的字典
    返回:
        添加base64编码后的字典
    """
    temp_file = param["temp_files"][-1]  # 获取最新保存的临时文件
    
    with open(temp_file, "rb") as f:
        encoded_image = base64.b64encode(f.read())
    
    encoded_image_text = encoded_image.decode("utf-8")
    base64_image = f"param:image;base64,{encoded_image_text}"
    
    param["base64_image"] = base64_image
    return param