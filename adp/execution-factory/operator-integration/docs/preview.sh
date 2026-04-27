#!/bin/bash

PORT=8000
DIR="$(dirname "$0")"

if ! command -v python3 &> /dev/null; then
    echo "错误：需要Python3环境"
    exit 1
fi
echo "先加载文档..."
./build.sh

# 创建一个临时的Python服务器脚本
cat > "$DIR/html_only_server.py" << 'EOF'
import http.server
import socketserver
import os
import urllib.parse
from pathlib import Path

# 从环境变量获取端口，默认为8000
PORT = int(os.environ.get('PORT', 8000))

class HTMLOnlyHandler(http.server.SimpleHTTPRequestHandler):
    def list_directory(self, path):
        """自定义目录列表，只显示HTML文件"""
        try:
            list = os.listdir(path)
        except OSError:
            self.send_error(404, "No permission to list directory")
            return None

        list.sort(key=lambda a: a.lower())
        displaypath = urllib.parse.unquote(self.path)

        # 直接发送响应，而不是调用send_head()
        self.send_response(200)
        self.send_header("Content-type", "text/html; charset=utf-8")
        self.end_headers()
        # 自定义HTML输出
        self.wfile.write(b'<!DOCTYPE html>\n')
        self.wfile.write(b'<meta charset="utf-8">\n')
        self.wfile.write(b'h1 { color: #333; }\n')
        self.wfile.write(b'table { border-collapse: collapse; width: 100%; }\n')
        self.wfile.write(b'td, th { border: 1px solid #ddd; padding: 8px; text-align: left; }\n')
        self.wfile.write(b'tr:nth-child(even) { background-color: #f2f2f2; }\n')
        self.wfile.write(b'a { text-decoration: none; color: #0066cc; }\n')
        self.wfile.write(b'a:hover { text-decoration: underline; }\n')
        self.wfile.write(b'</style>\n')
        self.wfile.write(b'</head>\n<body>\n')
        self.wfile.write(b'<h1>Directory listing for %s</h1>\n' % displaypath.encode('utf-8'))
        self.wfile.write(b'<hr>\n')
        self.wfile.write(b'<table>\n')
        # 添加返回上级目录链接
        if displaypath != '/':
            parent_path = os.path.dirname(displaypath.rstrip('/'))
            if parent_path == '':
                parent_path = '/'
            self.wfile.write(b'<tr><td><a href="%s">Parent Directory</a></td></tr>\n' % parent_path.encode('utf-8'))
        # 只显示HTML文件和目录
        for name in list:
            fullname = os.path.join(path, name)
            displayname = linkname = name

            # 如果是目录
            if os.path.isdir(fullname):
                self.wfile.write(b'<tr><td><a href="%s/">%s/</a></td></tr>\n' %
                       (urllib.parse.quote(linkname).encode('utf-8'), displayname.encode('utf-8')))
            # 如果是HTML文件
            elif name.lower().endswith('.html'):
                self.wfile.write(b'<tr><td><a href="%s">%s</a></td></tr>\n' %
                       (urllib.parse.quote(linkname).encode('utf-8'), displayname.encode('utf-8')))

        self.wfile.write(b'</table>\n')
        self.wfile.write(b'</body>\n</html>\n')
        return None

    def translate_path(self, path):
        """转换路径，确保只能访问指定目录"""
        path = path.split('?',1)[0]
        path = path.split('#',1)[0]
        path = urllib.parse.unquote(path)

        # 移除开头的斜杠
        if path.startswith('/'):
            path = path[1:]

        # 构建完整路径
        path = os.path.join(self.directory, path)

        # 规范化路径
        path = os.path.normpath(path)

        # 确保路径在允许的目录内
        if not path.startswith(self.directory):
            path = self.directory

        return path

    def end_headers(self):
        """添加CORS头部，允许跨域访问"""
        self.send_header('Access-Control-Allow-Origin', '*')
        super().end_headers()

# 启动服务器
Handler = HTMLOnlyHandler
socketserver.TCPServer(("", PORT), Handler).serve_forever()
EOF

echo "启动文档预览服务(仅HTML文件)..."
echo "访问地址: http://localhost:${PORT}"

# 启动Python服务器，只显示HTML文件
cd "$DIR/apis"
PORT=$PORT python3 ../html_only_server.py

# 清理临时Python脚本
rm -f "$DIR/html_only_server.py"