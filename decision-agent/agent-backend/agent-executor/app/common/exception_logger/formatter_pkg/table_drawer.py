# -*- coding:utf-8 -*-
"""
表格绘制工具
"""

from typing import Optional, List

from ..constants import COLORS, RESET, BOLD


class TableDrawer:
    """表格绘制器"""

    # 表格边框字符
    CORNER_TL = "┌"
    CORNER_TR = "┐"
    CORNER_BL = "└"
    CORNER_BR = "┘"
    HORIZONTAL = "─"
    VERTICAL = "│"
    T_DOWN = "┬"
    T_UP = "┴"
    T_RIGHT = "├"
    T_LEFT = "┤"
    CROSS = "┼"

    @classmethod
    def draw_table(
        cls,
        headers: List[str],
        rows: List[List[str]],
        col_widths: Optional[List[int]] = None,
        colorize: bool = False,
        indent: int = 2,
    ) -> str:
        """
        绘制表格

        Args:
            headers: 表头列表
            rows: 数据行列表
            col_widths: 列宽列表（可选）
            colorize: 是否添加颜色
            indent: 缩进空格数

        Returns:
            str: 表格字符串
        """
        if not headers:
            return ""

        # 计算列宽
        if col_widths is None:
            col_widths = []
            for i, header in enumerate(headers):
                max_width = len(header)
                for row in rows:
                    if i < len(row):
                        max_width = max(max_width, len(str(row[i])))
                col_widths.append(min(max_width + 2, 50))  # 最大50字符

        indent_str = " " * indent
        lines = []

        # 顶部边框
        top_border = cls.CORNER_TL
        for i, width in enumerate(col_widths):
            top_border += cls.HORIZONTAL * width
            top_border += cls.T_DOWN if i < len(col_widths) - 1 else cls.CORNER_TR

        if colorize:
            lines.append(f"{indent_str}{COLORS['border']}{top_border}{RESET}")
        else:
            lines.append(f"{indent_str}{top_border}")

        # 表头
        header_row = cls.VERTICAL
        for i, (header, width) in enumerate(zip(headers, col_widths)):
            cell = f" {header} ".ljust(width)
            if colorize:
                header_row += f"{BOLD}{COLORS['key']}{cell}{RESET}{COLORS['border']}{cls.VERTICAL}{RESET}"
            else:
                header_row += f"{cell}{cls.VERTICAL}"

        if colorize:
            lines.append(
                f"{indent_str}{COLORS['border']}{cls.VERTICAL}{RESET}{header_row[1:]}"
            )
        else:
            lines.append(f"{indent_str}{header_row}")

        # 表头分隔线
        sep_line = cls.T_RIGHT
        for i, width in enumerate(col_widths):
            sep_line += cls.HORIZONTAL * width
            sep_line += cls.CROSS if i < len(col_widths) - 1 else cls.T_LEFT

        if colorize:
            lines.append(f"{indent_str}{COLORS['border']}{sep_line}{RESET}")
        else:
            lines.append(f"{indent_str}{sep_line}")

        # 数据行
        for row in rows:
            data_row = cls.VERTICAL
            for i, width in enumerate(col_widths):
                value = str(row[i]) if i < len(row) else ""
                # 截断过长的值
                if len(value) > width - 2:
                    value = value[: width - 5] + "..."
                cell = f" {value} ".ljust(width)
                if colorize:
                    data_row += f"{COLORS['value']}{cell}{RESET}{COLORS['border']}{cls.VERTICAL}{RESET}"
                else:
                    data_row += f"{cell}{cls.VERTICAL}"

            if colorize:
                lines.append(
                    f"{indent_str}{COLORS['border']}{cls.VERTICAL}{RESET}{data_row[1:]}"
                )
            else:
                lines.append(f"{indent_str}{data_row}")

        # 底部边框
        bottom_border = cls.CORNER_BL
        for i, width in enumerate(col_widths):
            bottom_border += cls.HORIZONTAL * width
            bottom_border += cls.T_UP if i < len(col_widths) - 1 else cls.CORNER_BR

        if colorize:
            lines.append(f"{indent_str}{COLORS['border']}{bottom_border}{RESET}")
        else:
            lines.append(f"{indent_str}{bottom_border}")

        return "\n".join(lines)
