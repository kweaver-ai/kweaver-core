#!/usr/bin/python3
from dotenv import load_dotenv
from driveradapters import start_server

def main():
    # 加载.env文件
    load_dotenv()
    start_server()



if __name__ == "__main__":
    main()