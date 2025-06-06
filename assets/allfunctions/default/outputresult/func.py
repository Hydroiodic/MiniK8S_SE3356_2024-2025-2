from openai import OpenAI
from PIL import Image
import base64
import os
import requests
from io import BytesIO
from typing import Dict, Any

def handle(param) -> Dict[str, Any]:
    """
    清理临时文件
    参数:
        param: 处理字典
    返回:
        清理后的字典（可选移除临时文件引用）
    """
    print("Image Description:")
    print(param["description"])
    return param
        